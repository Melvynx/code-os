package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/melvynx/code-os/internal/config"
)

const apiRoot = "https://api.cloudflare.com/client/v4"

var defaultRunTokenPath = "/etc/cloudflared/code-os.token"

type ProvisionResult struct {
	TunnelID  string
	RunToken  string
	TokenFile string
	UnitPath  string
	Dashboard string
	PortHost  string
	Wildcard  bool
	Message   string
}

type apiClient struct {
	token  string
	base   string
	client *http.Client
}

func Provision(cfg config.Config) (config.Config, ProvisionResult, error) {
	result := ProvisionResult{Dashboard: cfg.Cloudflare.DashboardHost, PortHost: cfg.PublicPortHost, Wildcard: false}
	if strings.Contains(cfg.PublicPortHost, "*") {
		return cfg, result, fmt.Errorf("publicPortHost must not be a wildcard; use port{port}.<zone>")
	}
	if cfg.Cloudflare.TokenFile == "" || cfg.Cloudflare.AccountID == "" || cfg.Cloudflare.ZoneID == "" {
		return cfg, result, fmt.Errorf("cloudflare token, account id, and zone id are required to provision the tunnel")
	}
	if _, err := connectorBinary(); err != nil {
		return cfg, result, err
	}
	if tunnelID := strings.TrimSpace(cfg.Cloudflare.TunnelID); tunnelID != "" {
		if _, err := os.Stat(tokenFileOrDefault("")); err != nil {
			return cfg, result, fmt.Errorf("tunnel %s has no run token at /etc/cloudflared/code-os.token; omit --cloudflare-tunnel-id so setup can create the tunnel", tunnelID)
		}
	}
	token, err := os.ReadFile(cfg.Cloudflare.TokenFile)
	if err != nil {
		return cfg, result, fmt.Errorf("read Cloudflare token: %w", err)
	}
	client := apiClient{token: strings.TrimSpace(string(token)), client: http.DefaultClient}
	tunnelID := cfg.Cloudflare.TunnelID
	runToken := ""
	if tunnelID == "" {
		tunnelID, runToken, err = client.createTunnel(cfg.Cloudflare.AccountID, "code-os")
		if err != nil {
			return cfg, result, err
		}
		cfg.Cloudflare.TunnelID = tunnelID
	}
	if runToken != "" {
		tokenPath := defaultRunTokenPath
		if err := writePrivateFile(tokenPath, runToken); err != nil {
			return cfg, result, err
		}
		result.TokenFile = tokenPath
	}
	if err := client.putIngress(cfg); err != nil {
		return cfg, result, err
	}
	if err := client.ensureCNAME(cfg.Cloudflare.ZoneID, cfg.Cloudflare.DashboardHost, tunnelID+".cfargotunnel.com"); err != nil {
		return cfg, result, err
	}
	unitPath, err := installConnector(tokenFileOrDefault(result.TokenFile))
	if err != nil {
		return cfg, result, err
	}
	result.TunnelID = tunnelID
	result.UnitPath = unitPath
	result.Message = fmt.Sprintf("Provisioned tunnel %s for %s; ports use %s (no wildcard)", tunnelID, cfg.Cloudflare.DashboardHost, emptyPortHost(cfg.PublicPortHost))
	return cfg, result, nil
}

func (client apiClient) createTunnel(accountID, name string) (string, string, error) {
	var payload struct {
		Success bool `json:"success"`
		Result  struct {
			ID    string `json:"id"`
			Token string `json:"token"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := client.do(http.MethodPost, "/accounts/"+accountID+"/cfd_tunnel", map[string]string{
		"name": name, "config_src": "cloudflare",
	}, &payload); err != nil {
		return "", "", err
	}
	if !payload.Success || payload.Result.ID == "" {
		return "", "", fmt.Errorf("create tunnel: %s", firstErrorMessage(payload.Errors, "failed"))
	}
	return payload.Result.ID, payload.Result.Token, nil
}

func (client apiClient) putIngress(cfg config.Config) error {
	configuration, err := Configuration(cfg)
	if err != nil {
		return err
	}
	ingress := make([]map[string]string, 0, len(configuration.ManagedIngress)+1)
	for _, rule := range configuration.ManagedIngress {
		ingress = append(ingress, map[string]string{"hostname": rule.Hostname, "service": rule.Service})
	}
	ingress = append(ingress, map[string]string{"service": "http_status:404"})
	var payload struct {
		Success bool `json:"success"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := client.do(http.MethodPut, "/accounts/"+cfg.Cloudflare.AccountID+"/cfd_tunnel/"+cfg.Cloudflare.TunnelID+"/configurations", map[string]any{
		"config": map[string]any{"ingress": ingress},
	}, &payload); err != nil {
		return err
	}
	if !payload.Success {
		return fmt.Errorf("put ingress: %s", firstErrorMessage(payload.Errors, "failed"))
	}
	return nil
}

func (client apiClient) ensureCNAME(zoneID, name, target string) error {
	var listed struct {
		Success bool `json:"success"`
		Result  []struct {
			ID      string `json:"id"`
			Content string `json:"content"`
			Proxied bool   `json:"proxied"`
		} `json:"result"`
	}
	query := "/zones/" + zoneID + "/dns_records?type=CNAME&name=" + name
	if err := client.do(http.MethodGet, query, nil, &listed); err != nil {
		return err
	}
	body := map[string]any{"type": "CNAME", "name": name, "content": target, "ttl": 1, "proxied": true}
	if len(listed.Result) > 0 {
		record := listed.Result[0]
		if record.Content == target && record.Proxied {
			return nil
		}
		return client.do(http.MethodPut, "/zones/"+zoneID+"/dns_records/"+record.ID, body, nil)
	}
	return client.do(http.MethodPost, "/zones/"+zoneID+"/dns_records", body, nil)
}

func (client apiClient) do(method, path string, body any, dest any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	base := client.base
	if base == "" {
		base = apiRoot
	}
	request, err := http.NewRequest(method, base+path, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+client.token)
	request.Header.Set("Content-Type", "application/json")
	httpClient := client.client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode >= 300 {
		return fmt.Errorf("cloudflare %s %s: %s", method, path, strings.TrimSpace(string(data)))
	}
	if dest == nil {
		return nil
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("decode cloudflare response: %w", err)
	}
	return nil
}

func connectorBinary() (string, error) {
	candidates := []string{"/usr/local/bin/cloudflared"}
	if resolved, err := exec.LookPath("cloudflared"); err == nil {
		candidates = append([]string{resolved}, candidates...)
	}
	for _, path := range candidates {
		info, err := os.Stat(path)
		if err == nil && info.Mode().IsRegular() {
			return path, nil
		}
	}
	return "", fmt.Errorf("cloudflared is not installed; put it at /usr/local/bin/cloudflared before setup")
}

func installConnector(tokenFile string) (string, error) {
	if _, err := os.Stat(tokenFile); err != nil {
		return "", fmt.Errorf("cloudflared run token missing at %s", tokenFile)
	}
	binary, err := connectorBinary()
	if err != nil {
		return "", err
	}
	unitPath := "/etc/systemd/system/cloudflared-code-os.service"
	if err := os.WriteFile(unitPath, []byte(connectorUnit(binary, tokenFile)), 0o644); err != nil {
		return "", fmt.Errorf("write cloudflared unit (need root): %w", err)
	}
	_ = exec.Command("systemctl", "daemon-reload").Run()
	if err := exec.Command("systemctl", "enable", "--now", "cloudflared-code-os.service").Run(); err != nil {
		return unitPath, fmt.Errorf("enable cloudflared-code-os: %w", err)
	}
	return unitPath, nil
}

func connectorUnit(binary, tokenFile string) string {
	return fmt.Sprintf(`[Unit]
Description=Cloudflare Tunnel for Code OS
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s tunnel --no-autoupdate run --token-file %s
Restart=always
RestartSec=5
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadOnlyPaths=%s

[Install]
WantedBy=multi-user.target
`, binary, tokenFile, tokenFile)
}

func writePrivateFile(path, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimSpace(value)+"\n"), 0o600)
}

func tokenFileOrDefault(path string) string {
	if path != "" {
		return path
	}
	return defaultRunTokenPath
}

func firstErrorMessage(errors []struct {
	Message string `json:"message"`
}, fallback string) string {
	if len(errors) == 0 || errors[0].Message == "" {
		return fallback
	}
	return errors[0].Message
}

func emptyPortHost(value string) string {
	if strings.TrimSpace(value) == "" {
		return "port{port}.<zone>"
	}
	return value
}
