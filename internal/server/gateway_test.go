package server

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/melvynx/code-os/internal/config"
	"github.com/melvynx/code-os/internal/model"
)

func TestGatewayProtectsAndProxiesHealthyPort(t *testing.T) {
	t.Parallel()
	origin, port := loopbackOrigin(t, func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-Origin-Host", request.Host)
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte("proxied"))
	})
	defer origin.Close()
	fixture := newGatewayFixture(t, port)
	host := fmt.Sprintf("port%d.example.com", port)

	unauthorized := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://"+host+"/private", nil)
	request.Host = host
	fixture.handler.ServeHTTP(unauthorized, request)
	if unauthorized.Code != http.StatusSeeOther || !strings.HasPrefix(unauthorized.Header().Get("Location"), "/_code-os/login?") {
		t.Fatalf("unauthorized response = %d %q", unauthorized.Code, unauthorized.Header().Get("Location"))
	}

	cookie := gatewayLoginFixture(t, fixture.handler, host)
	authorized := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "http://"+host+"/private", nil)
	request.Host = host
	request.AddCookie(cookie)
	fixture.handler.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK || authorized.Body.String() != "proxied" {
		t.Fatalf("authorized response = %d %q", authorized.Code, authorized.Body.String())
	}
	if got := authorized.Header().Get("X-Origin-Host"); got != "localhost:"+strconv.Itoa(port) {
		t.Fatalf("origin host = %q", got)
	}
}

func TestGatewayRewritesLocalhostRedirect(t *testing.T) {
	t.Parallel()
	var originPort int
	origin, port := loopbackOrigin(t, func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Location", fmt.Sprintf("http://localhost:%d/auth/signin", originPort))
		response.WriteHeader(http.StatusTemporaryRedirect)
	})
	defer origin.Close()
	originPort = port
	fixture := newGatewayFixture(t, port)
	host := fmt.Sprintf("port%d.example.com", port)
	cookie := gatewayLoginFixture(t, fixture.handler, host)

	request := httptest.NewRequest(http.MethodGet, "http://"+host+"/orgs", nil)
	request.Host = host
	request.Header.Set("X-Forwarded-Proto", "https")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)

	if location := response.Header().Get("Location"); location != "https://"+host+"/auth/signin" {
		t.Fatalf("Location = %q", location)
	}
}

func TestGatewayRejectsPortMissingFromHealthySnapshot(t *testing.T) {
	t.Parallel()
	fixture := newGatewayFixture(t, 43210)
	host := "port43211.example.com"
	cookie := gatewayLoginFixture(t, fixture.handler, host)
	request := httptest.NewRequest(http.MethodGet, "http://"+host+"/", nil)
	request.Host = host
	request.AddCookie(cookie)
	response := httptest.NewRecorder()

	fixture.handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestPrivateFileBypassIsImageOnly(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeFixtureFile(t, filepath.Join(directory, "image.png"), "png-bytes")
	writeFixtureFile(t, filepath.Join(directory, "notes.txt"), "private")
	fixture := newFileFixture(t, directory)

	image := httptest.NewRecorder()
	fixture.handler.ServeHTTP(image, httptest.NewRequest(http.MethodGet, "/files/image.png?bp=media-only-key", nil))
	if image.Code != http.StatusOK || image.Body.String() != "png-bytes" {
		t.Fatalf("image response = %d %q", image.Code, image.Body.String())
	}
	if image.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("Cache-Control = %q", image.Header().Get("Cache-Control"))
	}

	nonImage := httptest.NewRecorder()
	fixture.handler.ServeHTTP(nonImage, httptest.NewRequest(http.MethodGet, "/files/notes.txt?bp=media-only-key", nil))
	if nonImage.Code != http.StatusUnauthorized {
		t.Fatalf("non-image status = %d, want 401", nonImage.Code)
	}

	withoutKey := httptest.NewRecorder()
	fixture.handler.ServeHTTP(withoutKey, httptest.NewRequest(http.MethodGet, "/files/image.png", nil))
	if withoutKey.Code != http.StatusSeeOther {
		t.Fatalf("anonymous image status = %d, want 303", withoutKey.Code)
	}
}

func TestPrivateFileResolverRejectsEscapeAndHiddenFiles(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeFixtureFile(t, filepath.Join(directory, "image.png"), "ok")
	for _, path := range []string{"../secret", ".env", "folder/.secret"} {
		if resolved, ok := resolvePrivateFile(directory, path); ok {
			t.Fatalf("resolvePrivateFile(%q) = %q, want rejection", path, resolved)
		}
	}
}

func TestLoginRateLimitAppliesAfterRepeatedFailures(t *testing.T) {
	t.Parallel()
	fixture := newAuthFixture(t)
	for attempt := 0; attempt < maxLoginAttempts; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("username=bad&password=bad"))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		response := httptest.NewRecorder()
		fixture.handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d", attempt+1, response.Code)
		}
	}
	request := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("username=code-os&password=correct-horse"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusTooManyRequests || response.Header().Get("Retry-After") == "" {
		t.Fatalf("rate-limited response = %d Retry-After=%q", response.Code, response.Header().Get("Retry-After"))
	}
}

type gatewayFixture struct {
	handler http.Handler
}

func newGatewayFixture(t *testing.T, healthyPort int) gatewayFixture {
	t.Helper()
	directory := t.TempDir()
	passwordFile := filepath.Join(directory, "password")
	bypassFile := filepath.Join(directory, "bypass")
	sessionKeyFile := filepath.Join(directory, "session-key")
	writeFixtureFile(t, passwordFile, "correct-horse\n")
	writeFixtureFile(t, bypassFile, "media-only-key\n")
	writeFixtureFile(t, sessionKeyFile, "0123456789abcdef0123456789abcdef\n")
	healthy := true
	service := &Service{media: map[string]string{}, snapshot: model.Snapshot{Apps: []model.Application{{Port: healthyPort, State: "running", Healthy: &healthy}}}}
	server := HTTPServer{Config: config.Config{
		EnvironmentName: "test", PublicPortHost: "port{port}.example.com",
		Auth: config.Auth{Username: "code-os", PasswordFile: passwordFile, BypassKeyFile: bypassFile, SessionKeyFile: sessionKeyFile},
	}, Service: service}
	handler, err := server.Handler()
	if err != nil {
		t.Fatal(err)
	}
	return gatewayFixture{handler: handler}
}

func newFileFixture(t *testing.T, filesRoot string) authFixture {
	t.Helper()
	directory := t.TempDir()
	passwordFile := filepath.Join(directory, "password")
	bypassFile := filepath.Join(directory, "bypass")
	sessionKeyFile := filepath.Join(directory, "session-key")
	writeFixtureFile(t, passwordFile, "correct-horse\n")
	writeFixtureFile(t, bypassFile, "media-only-key\n")
	writeFixtureFile(t, sessionKeyFile, "0123456789abcdef0123456789abcdef\n")
	server := HTTPServer{Config: config.Config{
		EnvironmentName: "test", FilesRoot: filesRoot,
		Auth: config.Auth{Username: "code-os", PasswordFile: passwordFile, BypassKeyFile: bypassFile, SessionKeyFile: sessionKeyFile},
	}, Service: &Service{media: map[string]string{}}}
	handler, err := server.Handler()
	if err != nil {
		t.Fatal(err)
	}
	return authFixture{handler: handler, bypassKey: "media-only-key"}
}

func gatewayLoginFixture(t *testing.T, handler http.Handler, host string) *http.Cookie {
	t.Helper()
	body := strings.NewReader("username=code-os&password=correct-horse&next=%2F")
	request := httptest.NewRequest(http.MethodPost, "http://"+host+"/_code-os/auth/login", body)
	request.Host = host
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Forwarded-Proto", "https")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther || len(response.Result().Cookies()) != 1 {
		t.Fatalf("gateway login failed: %d %s", response.Code, response.Body.String())
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Name != gatewaySessionCookieName || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("gateway cookie flags = %#v", cookie)
	}
	return cookie
}

func loopbackOrigin(t *testing.T, handler http.HandlerFunc) (*httptest.Server, int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	return server, listenerPort(t, listener.Addr())
}

func listenerPort(t *testing.T, address net.Addr) int {
	t.Helper()
	_, portText, err := net.SplitHostPort(address.String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	return port
}
