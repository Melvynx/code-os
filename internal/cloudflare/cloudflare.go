package cloudflare

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/melvynx/code-os/internal/config"
)

type IngressRule struct {
	Hostname string `json:"hostname,omitempty"`
	Service  string `json:"service"`
}

type TunnelConfiguration struct {
	Mode           string        `json:"mode"`
	DashboardHost  string        `json:"dashboardHost"`
	PublicPortHost string        `json:"publicPortHost,omitempty"`
	Wildcard       bool          `json:"wildcard"`
	ManagedIngress []IngressRule `json:"managedIngress"`
	Fallback       *IngressRule  `json:"fallback,omitempty"`
	Instructions   string        `json:"instructions"`
}

func Configuration(cfg config.Config) (TunnelConfiguration, error) {
	hostname := strings.TrimSpace(cfg.Cloudflare.DashboardHost)
	if hostname == "" {
		return TunnelConfiguration{}, fmt.Errorf("dashboard hostname is not configured")
	}
	if strings.Contains(cfg.PublicPortHost, "*") {
		return TunnelConfiguration{}, fmt.Errorf("publicPortHost must not be a wildcard; use port{port}.<zone>")
	}
	mode := cfg.Cloudflare.TunnelMode
	if mode == "" {
		mode = "shared"
	}
	portHost := strings.TrimSpace(cfg.PublicPortHost)
	instructions := "Dashboard hostname only in ingress. Publish each healthy Portly port as an individual CNAME from publicPortHost. Never add a wildcard."
	if portHost != "" {
		instructions = fmt.Sprintf("Dashboard hostname only in ingress. Individual CNAMEs matching %s — never *.", portHost)
	}
	configuration := TunnelConfiguration{
		Mode: mode, DashboardHost: hostname, PublicPortHost: portHost, Wildcard: false,
		ManagedIngress: []IngressRule{{
			Hostname: hostname,
			Service:  "http://" + cfg.Address,
		}},
		Instructions: instructions,
	}
	if mode == "dedicated" {
		configuration.Fallback = &IngressRule{Service: "http_status:404"}
		configuration.Instructions = instructions + " Dedicated tunnel: keep http_status:404 last."
	}
	return configuration, nil
}

func Render(cfg config.Config) (string, error) {
	configuration, err := Configuration(cfg)
	if err != nil {
		return "", err
	}
	payload, err := json.MarshalIndent(configuration, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode tunnel configuration: %w", err)
	}
	return string(payload) + "\n", nil
}
