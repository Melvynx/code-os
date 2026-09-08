package cloudflare

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/melvynx/code-os/internal/config"
)

func TestProvisionRejectsExistingTunnelWithoutRunToken(t *testing.T) {
	if _, err := connectorBinary(); err != nil {
		t.Skip(err.Error())
	}
	previous := defaultRunTokenPath
	defaultRunTokenPath = filepath.Join(t.TempDir(), "missing-run.token")
	t.Cleanup(func() { defaultRunTokenPath = previous })
	token := filepath.Join(t.TempDir(), "api-token")
	if err := os.WriteFile(token, []byte("test-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := Provision(config.Config{
		Address:        "127.0.0.1:7890",
		PublicPortHost: "port{port}.example.com",
		Cloudflare: config.Cloudflare{
			DashboardHost: "code-os.example.com",
			TokenFile:     token,
			AccountID:     "account",
			ZoneID:        "zone",
			TunnelID:      "existing-tunnel",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "run token") {
		t.Fatalf("Provision() error = %v, want run token error", err)
	}
}

func TestConnectorUnitUsesResolvedBinary(t *testing.T) {
	t.Parallel()
	unit := connectorUnit("/usr/local/bin/cloudflared", "/etc/cloudflared/code-os.token")
	if !strings.Contains(unit, "ExecStart=/usr/local/bin/cloudflared tunnel") {
		t.Fatalf("unit = %s", unit)
	}
}

func TestDoRejectsHTTPErrorsEvenWhenDecoding(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(response, `{"success":false,"errors":[{"message":"nope"}]}`)
	}))
	t.Cleanup(server.Close)
	client := apiClient{token: "test", base: server.URL, client: server.Client()}
	var payload struct {
		Success bool `json:"success"`
	}
	err := client.do(http.MethodGet, "/unused", nil, &payload)
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("do() error = %v, want HTTP error containing nope", err)
	}
}
