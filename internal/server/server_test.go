package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/melvynx/code-os/internal/config"
)

type authFixture struct {
	handler   http.Handler
	mediaID   string
	bypassKey string
}

func TestPublicLandingStaysAvailableWithoutAuthentication(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if challenge := response.Header().Get("WWW-Authenticate"); challenge != "" {
		t.Fatalf("WWW-Authenticate = %q, want no browser dialog", challenge)
	}
}

func TestAnonymousDashboardGetsLoginPageInsteadOfBasicAuthDialog(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/app/projects", nil))
	if response.Code != http.StatusSeeOther || !strings.HasPrefix(response.Header().Get("Location"), "/login?next=") {
		t.Fatalf("dashboard response = %d location %q", response.Code, response.Header().Get("Location"))
	}
	if challenge := response.Header().Get("WWW-Authenticate"); challenge != "" {
		t.Fatalf("WWW-Authenticate = %q, want no browser dialog", challenge)
	}
}

func TestMediaBypassKeyRendersImageWithoutSession(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	request := httptest.NewRequest(http.MethodGet, "/media/"+fixture.mediaID+"?bp="+fixture.bypassKey, nil)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if cacheControl := response.Header().Get("Cache-Control"); cacheControl != "private, no-store" {
		t.Fatalf("Cache-Control = %q, want private, no-store", cacheControl)
	}
}

func TestBypassKeyCannotAccessDashboardAPI(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	request := httptest.NewRequest(http.MethodGet, "/api/overview?bp="+fixture.bypassKey, nil)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestWrongBypassKeyCannotAccessMedia(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	request := httptest.NewRequest(http.MethodGet, "/media/"+fixture.mediaID+"?bp=wrong-key", nil)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestLoginPageSupportsPasswordManagers(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/login", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	body := response.Body.String()
	for _, expected := range []string{`name="username"`, `autocomplete="username"`, `name="password"`, `autocomplete="current-password"`} {
		if !strings.Contains(body, expected) {
			t.Errorf("login page missing %s", expected)
		}
	}
	if challenge := response.Header().Get("WWW-Authenticate"); challenge != "" {
		t.Fatalf("WWW-Authenticate = %q, want no browser dialog", challenge)
	}
}

func TestLoginRejectsExternalRedirect(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	body := strings.NewReader("username=code-os&password=correct-horse&next=https%3A%2F%2Fevil.example")
	request := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, request)

	if location := response.Header().Get("Location"); location != "/app/" {
		t.Fatalf("Location = %q, want /app/", location)
	}
}

func TestAuthenticatedClientGetsSPAForRouterPath(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	sessionCookie := loginFixture(t, fixture)
	request := httptest.NewRequest(http.MethodGet, "/app/projects", nil)
	request.AddCookie(sessionCookie)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if !strings.Contains(response.Body.String(), `<div id="root"></div>`) {
		t.Fatal("router path did not serve the dashboard shell")
	}
}

func TestDashboardCSPAllowsRadixRuntimeStylesButKeepsScriptsStrict(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/login", nil))

	policy := response.Header().Get("Content-Security-Policy")
	if !strings.Contains(policy, "style-src 'self' 'unsafe-inline'") {
		t.Fatalf("Content-Security-Policy = %q, want Radix inline styles allowed", policy)
	}
	if !strings.Contains(policy, "script-src 'self'") || strings.Contains(policy, "script-src 'self' 'unsafe-inline'") {
		t.Fatalf("Content-Security-Policy = %q, want strict same-origin scripts", policy)
	}
}

func TestAuthenticatedUnknownAPIStaysNotFound(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	sessionCookie := loginFixture(t, fixture)
	request := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
	request.AddCookie(sessionCookie)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestLoginFormCreatesSessionForPasswordManagerFlow(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	body := strings.NewReader("username=code-os&password=correct-horse&next=%2F")
	request := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303; body = %s", response.Code, response.Body.String())
	}
	if len(response.Result().Cookies()) != 1 {
		t.Fatalf("cookies = %d, want 1", len(response.Result().Cookies()))
	}
	sessionCookie := response.Result().Cookies()[0]
	if !sessionCookie.HttpOnly || sessionCookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("session cookie flags = HttpOnly:%t SameSite:%d", sessionCookie.HttpOnly, sessionCookie.SameSite)
	}
	authorizedRequest := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	authorizedRequest.AddCookie(sessionCookie)
	authorized := httptest.NewRecorder()
	fixture.handler.ServeHTTP(authorized, authorizedRequest)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authenticated status = %d, want 200", authorized.Code)
	}
}

func loginFixture(t *testing.T, fixture authFixture) *http.Cookie {
	t.Helper()
	body := strings.NewReader("username=code-os&password=correct-horse&next=%2F")
	request := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther || len(response.Result().Cookies()) != 1 {
		t.Fatalf("login failed with status %d", response.Code)
	}
	return response.Result().Cookies()[0]
}

func newAuthFixture(t *testing.T) authFixture {
	t.Helper()
	directory := t.TempDir()
	passwordFile := filepath.Join(directory, "password")
	bypassFile := filepath.Join(directory, "bypass")
	sessionKeyFile := filepath.Join(directory, "session-key")
	imagePath := filepath.Join(directory, "image.png")
	writeFixtureFile(t, passwordFile, "correct-horse\n")
	writeFixtureFile(t, bypassFile, "media-only-key\n")
	writeFixtureFile(t, sessionKeyFile, "0123456789abcdef0123456789abcdef\n")
	writeFixtureFile(t, imagePath, "not-a-real-png")
	service := &Service{media: map[string]string{"image": imagePath}}
	httpServer := HTTPServer{
		Config: config.Config{
			EnvironmentName: "test",
			Auth: config.Auth{
				Username: "code-os", PasswordFile: passwordFile, BypassKeyFile: bypassFile, SessionKeyFile: sessionKeyFile,
			},
		},
		Service: service,
	}
	handler, err := httpServer.Handler()
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	return authFixture{handler: handler, mediaID: "image", bypassKey: "media-only-key"}
}

func writeFixtureFile(t *testing.T, path, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}
