package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/melvynx/code-os/internal/config"
)

type authFixture struct {
	handler        http.Handler
	mediaID        string
	bypassKey      string
	trustedIPsPath string
}

func TestDashboardRootRedirectsToCommandCenter(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusTemporaryRedirect || response.Header().Get("Location") != "/app/" {
		t.Fatalf("root response = %d location %q, want 307 /app/", response.Code, response.Header().Get("Location"))
	}
	if challenge := response.Header().Get("WWW-Authenticate"); challenge != "" {
		t.Fatalf("WWW-Authenticate = %q, want no browser dialog", challenge)
	}
}

func TestPublicDocumentationStaysAvailableWithoutAuthentication(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/docs/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("documentation status = %d, want 200", response.Code)
	}
	if challenge := response.Header().Get("WWW-Authenticate"); challenge != "" {
		t.Fatalf("WWW-Authenticate = %q, want no browser dialog", challenge)
	}
}

func TestAuthenticationFaviconIsPublic(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	request := httptest.NewRequest(http.MethodGet, "https://code-os.example/favicon.ico", nil)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("favicon response = %d content-type %q", response.Code, response.Header().Get("Content-Type"))
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

func TestTamperedSessionCookieIsRejected(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	cookie := loginFixture(t, fixture)
	cookie.Value += "tampered"
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestExpiredSessionCookieIsRejected(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	auth := authenticator{username: "code-os", sessionKey: []byte("0123456789abcdef0123456789abcdef\n")}
	payload := base64.RawURLEncoding.EncodeToString([]byte("code-os\n" + fmt.Sprint(time.Now().Add(-time.Hour).Unix())))
	cookie := &http.Cookie{Name: sessionCookieName, Value: payload + "." + auth.sign(payload)}
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	request.AddCookie(cookie)
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

func TestProcessMutationRequiresSameOriginAndKnownTarget(t *testing.T) {
	t.Parallel()
	server := HTTPServer{Service: &Service{}}

	crossOrigin := httptest.NewRequest(http.MethodPost, "https://code-os.example/api/agents/12:34/terminate", nil)
	crossOrigin.Host = "code-os.example"
	crossOrigin.Header.Set("Origin", "https://attacker.example")
	crossOrigin.Header.Set("X-Forwarded-Proto", "https")
	crossOriginResponse := httptest.NewRecorder()
	server.terminateAgent(crossOriginResponse, crossOrigin)
	if crossOriginResponse.Code != http.StatusForbidden {
		t.Fatalf("cross-origin termination = %d, want 403", crossOriginResponse.Code)
	}

	unknown := httptest.NewRequest(http.MethodPost, "https://code-os.example/api/agents/12:34/terminate", nil)
	unknown.Host = "code-os.example"
	unknown.Header.Set("Origin", "https://code-os.example")
	unknown.Header.Set("X-Forwarded-Proto", "https")
	unknown.SetPathValue("id", "12:34")
	unknownResponse := httptest.NewRecorder()
	server.terminateAgent(unknownResponse, unknown)
	if unknownResponse.Code != http.StatusConflict {
		t.Fatalf("unknown agent termination = %d, want 409", unknownResponse.Code)
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

func TestAuthenticatedClientCanTrustAndRevokeExactIP(t *testing.T) {
	t.Parallel()
	fixture := newTrustedAuthFixture(t)
	loginBody := strings.NewReader("username=code-os&password=correct-horse&next=%2Fapp%2Fprojects")
	loginRequest := httptest.NewRequest(http.MethodPost, "https://code-os.example/auth/login", loginBody)
	loginRequest.Host = "code-os.example"
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRequest.Header.Set("X-Forwarded-Proto", "https")
	loginRequest.Header.Set("CF-Connecting-IP", "203.0.113.7")
	loginResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusSeeOther || loginResponse.Header().Get("Location") != "/trust-ip?next=%2Fapp%2Fprojects" {
		t.Fatalf("login response = %d location %q", loginResponse.Code, loginResponse.Header().Get("Location"))
	}
	if len(loginResponse.Result().Cookies()) != 1 {
		t.Fatalf("login cookies = %d, want 1", len(loginResponse.Result().Cookies()))
	}
	sessionCookie := loginResponse.Result().Cookies()[0]

	pageRequest := httptest.NewRequest(http.MethodGet, "https://code-os.example/trust-ip?next=%2Fapp%2Fprojects", nil)
	pageRequest.Host = "code-os.example"
	pageRequest.Header.Set("X-Forwarded-Proto", "https")
	pageRequest.Header.Set("CF-Connecting-IP", "203.0.113.7")
	pageRequest.AddCookie(sessionCookie)
	pageResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(pageResponse, pageRequest)
	if pageResponse.Code != http.StatusOK || !strings.Contains(pageResponse.Body.String(), "203.0.113.7") {
		t.Fatalf("trust page = %d %q", pageResponse.Code, pageResponse.Body.String())
	}

	trustBody := strings.NewReader("next=%2Fapp%2Fprojects")
	trustRequest := httptest.NewRequest(http.MethodPost, "https://code-os.example/auth/trust-ip", trustBody)
	trustRequest.Host = "code-os.example"
	trustRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	trustRequest.Header.Set("Origin", "https://code-os.example")
	trustRequest.Header.Set("X-Forwarded-Proto", "https")
	trustRequest.Header.Set("CF-Connecting-IP", "203.0.113.7")
	trustRequest.AddCookie(sessionCookie)
	trustResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(trustResponse, trustRequest)
	if trustResponse.Code != http.StatusSeeOther || trustResponse.Header().Get("Location") != "/app/projects" {
		t.Fatalf("trust response = %d location %q", trustResponse.Code, trustResponse.Header().Get("Location"))
	}
	assertAPIStatusForIP(t, fixture.handler, "203.0.113.7", http.StatusOK)
	assertAPIStatusForIP(t, fixture.handler, "203.0.113.8", http.StatusUnauthorized)

	stored, err := os.ReadFile(fixture.trustedIPsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != "203.0.113.7\n" {
		t.Fatalf("trusted IP storage = %q", stored)
	}
	info, err := os.Stat(fixture.trustedIPsPath)
	if err != nil {
		t.Fatalf("stat trusted IP storage: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("trusted IP storage mode = %v, want 0600", info.Mode().Perm())
	}

	revokeRequest := httptest.NewRequest(http.MethodDelete, "https://code-os.example/api/trusted-ip", nil)
	revokeRequest.Host = "code-os.example"
	revokeRequest.Header.Set("Origin", "https://code-os.example")
	revokeRequest.Header.Set("X-Forwarded-Proto", "https")
	revokeRequest.Header.Set("CF-Connecting-IP", "203.0.113.7")
	revokeResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(revokeResponse, revokeRequest)
	if revokeResponse.Code != http.StatusOK || !strings.Contains(revokeResponse.Body.String(), `"trusted":false`) {
		t.Fatalf("revoke response = %d %q", revokeResponse.Code, revokeResponse.Body.String())
	}
	assertAPIStatusForIP(t, fixture.handler, "203.0.113.7", http.StatusUnauthorized)
}

func TestTrustIPMutationRequiresSessionAndSameOrigin(t *testing.T) {
	t.Parallel()
	fixture := newTrustedAuthFixture(t)
	unauthenticated := httptest.NewRequest(http.MethodPost, "https://code-os.example/auth/trust-ip", strings.NewReader("next=%2Fapp%2F"))
	unauthenticated.Host = "code-os.example"
	unauthenticated.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	unauthenticated.Header.Set("Origin", "https://code-os.example")
	unauthenticated.Header.Set("X-Forwarded-Proto", "https")
	unauthenticated.Header.Set("CF-Connecting-IP", "203.0.113.7")
	unauthenticatedResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(unauthenticatedResponse, unauthenticated)
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated trust status = %d, want 401", unauthenticatedResponse.Code)
	}

	cookie := loginFixtureForIP(t, fixture, "203.0.113.7")
	crossOrigin := httptest.NewRequest(http.MethodPost, "https://code-os.example/auth/trust-ip", strings.NewReader("next=%2Fapp%2F"))
	crossOrigin.Host = "code-os.example"
	crossOrigin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	crossOrigin.Header.Set("Origin", "https://attacker.example")
	crossOrigin.Header.Set("X-Forwarded-Proto", "https")
	crossOrigin.Header.Set("CF-Connecting-IP", "203.0.113.7")
	crossOrigin.AddCookie(cookie)
	crossOriginResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(crossOriginResponse, crossOrigin)
	if crossOriginResponse.Code != http.StatusForbidden {
		t.Fatalf("cross-origin trust status = %d, want 403", crossOriginResponse.Code)
	}
	assertAPIStatusForIP(t, fixture.handler, "203.0.113.7", http.StatusUnauthorized)
}

func TestAuthenticatedSettingsCanTrustCurrentIP(t *testing.T) {
	t.Parallel()
	fixture := newTrustedAuthFixture(t)
	cookie := loginFixtureForIP(t, fixture, "203.0.113.7")

	request := httptest.NewRequest(http.MethodPost, "https://code-os.example/api/trusted-ip", nil)
	request.Host = "code-os.example"
	request.Header.Set("Origin", "https://code-os.example")
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Header.Set("CF-Connecting-IP", "203.0.113.7")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"trusted":true`) {
		t.Fatalf("trust current IP response = %d %q", response.Code, response.Body.String())
	}
	assertAPIStatusForIP(t, fixture.handler, "203.0.113.7", http.StatusOK)
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

func newTrustedAuthFixture(t *testing.T) authFixture {
	t.Helper()
	directory := t.TempDir()
	passwordFile := filepath.Join(directory, "password")
	bypassFile := filepath.Join(directory, "bypass")
	sessionKeyFile := filepath.Join(directory, "session-key")
	trustedIPsFile := filepath.Join(directory, "trusted-ips")
	imagePath := filepath.Join(directory, "image.png")
	writeFixtureFile(t, passwordFile, "correct-horse\n")
	writeFixtureFile(t, bypassFile, "media-only-key\n")
	writeFixtureFile(t, sessionKeyFile, "0123456789abcdef0123456789abcdef\n")
	writeFixtureFile(t, trustedIPsFile, "")
	writeFixtureFile(t, imagePath, "not-a-real-png")
	service := &Service{media: map[string]string{"image": imagePath}}
	httpServer := HTTPServer{
		Config: config.Config{
			EnvironmentName: "test",
			Auth: config.Auth{
				Username: "code-os", PasswordFile: passwordFile, BypassKeyFile: bypassFile,
				SessionKeyFile: sessionKeyFile, TrustedIPsFile: trustedIPsFile,
			},
		},
		Service: service,
	}
	handler, err := httpServer.Handler()
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	return authFixture{handler: handler, mediaID: "image", bypassKey: "media-only-key", trustedIPsPath: trustedIPsFile}
}

func loginFixtureForIP(t *testing.T, fixture authFixture, address string) *http.Cookie {
	t.Helper()
	body := strings.NewReader("username=code-os&password=correct-horse&next=%2Fapp%2F")
	request := httptest.NewRequest(http.MethodPost, "https://code-os.example/auth/login", body)
	request.Host = "code-os.example"
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Header.Set("CF-Connecting-IP", address)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther || len(response.Result().Cookies()) != 1 {
		t.Fatalf("login failed with status %d", response.Code)
	}
	return response.Result().Cookies()[0]
}

func assertAPIStatusForIP(t *testing.T, handler http.Handler, address string, expected int) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "https://code-os.example/api/health", nil)
	request.Host = "code-os.example"
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Header.Set("CF-Connecting-IP", address)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != expected {
		t.Fatalf("API status for %s = %d, want %d", address, response.Code, expected)
	}
}

func writeFixtureFile(t *testing.T, path, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}
