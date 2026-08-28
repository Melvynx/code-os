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
	ManagedIngress []IngressRule `json:"managedIngress"`
	Fallback       *IngressRule  `json:"fallback,omitempty"`
	Instructions   string        `json:"instructions"`
}

func Configuration(cfg config.Config) (TunnelConfiguration, error) {
	hostname := strings.TrimSpace(cfg.Cloudflare.DashboardHost)
	if hostname == "" {
		return TunnelConfiguration{}, fmt.Errorf("dashboard hostname is not configured")
	}
	mode := cfg.Cloudflare.TunnelMode
	if mode == "" {
		mode = "shared"
	}
	configuration := TunnelConfiguration{
		Mode: mode,
		ManagedIngress: []IngressRule{{
			Hostname: hostname,
			Service:  "http://" + cfg.Address,
		}},
		Instructions: "Insert the managed ingress before the tunnel's final fallback and preserve every unrelated rule.",
	}
	if mode == "dedicated" {
		configuration.Fallback = &IngressRule{Service: "http_status:404"}
		configuration.Instructions = "Use the managed ingress followed by the fallback in this dedicated Code OS tunnel."
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
