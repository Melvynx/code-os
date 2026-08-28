package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/melvynx/stackenv/internal/config"
)

func TestHandlerRequiresConfiguredAuthentication(t *testing.T) {
	t.Parallel()
	passwordFile := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(passwordFile, []byte("correct-horse\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	httpServer := HTTPServer{
		Config:  config.Config{EnvironmentName: "test", Auth: config.Auth{Username: "stackenv", PasswordFile: passwordFile}},
		Service: &Service{media: make(map[string]string)},
	}
	handler, err := httpServer.Handler()
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", unauthorized.Code)
	}

	authorizedRequest := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	authorizedRequest.SetBasicAuth("stackenv", "correct-horse")
	authorized := httptest.NewRecorder()
	handler.ServeHTTP(authorized, authorizedRequest)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authenticated status = %d, want 200", authorized.Code)
	}
}
