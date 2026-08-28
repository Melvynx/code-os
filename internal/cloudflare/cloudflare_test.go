package cloudflare

import (
	"strings"
	"testing"

	"github.com/melvynx/stackenv/internal/config"
)

func TestRenderUsesConfiguredHostnameAndLoopback(t *testing.T) {
	t.Parallel()
	cfg := config.Config{Address: "127.0.0.1:7890", Cloudflare: config.Cloudflare{DashboardHost: "stackenv.example.com"}}
	payload, err := Render(cfg)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, expected := range []string{"stackenv.example.com", "http://127.0.0.1:7890", "preserve every unrelated rule"} {
		if !strings.Contains(payload, expected) {
			t.Fatalf("Render() = %s, want %q", payload, expected)
		}
	}
	if strings.Contains(payload, "http_status:404") {
		t.Fatalf("shared tunnel output must not inject a fallback: %s", payload)
	}
}
