package config

import "strings"

// DefaultPublicPortHost derives port{port}.<zone> from the dashboard hostname.
// code-os.example.com and example.com both become port{port}.example.com.
// The zone is never hardcoded; any DNS name works.
func DefaultPublicPortHost(dashboardHost string) string {
	host := strings.ToLower(strings.TrimSpace(dashboardHost))
	host = strings.TrimSuffix(host, ".")
	if host == "" || !strings.Contains(host, ".") {
		return ""
	}
	parts := strings.Split(host, ".")
	zone := host
	if len(parts) >= 3 {
		zone = strings.Join(parts[1:], ".")
	}
	return "port{port}." + zone
}
