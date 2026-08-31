package server

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
)

func TestTrustedIPStorePersistsNormalizedExactAddresses(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "trusted-ips")
	store, err := newTrustedIPStore(path)
	if err != nil {
		t.Fatal(err)
	}
	address := netip.MustParseAddr("::ffff:203.0.113.7")
	if err := store.Add(address); err != nil {
		t.Fatal(err)
	}
	if !store.Contains(netip.MustParseAddr("203.0.113.7")) || store.Contains(netip.MustParseAddr("203.0.113.8")) {
		t.Fatal("trusted IP store did not preserve exact normalized membership")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "203.0.113.7\n" {
		t.Fatalf("trusted IP file = %q", content)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("trusted IP mode = %04o, want 0600", info.Mode().Perm())
	}
	if err := store.Remove(address); err != nil {
		t.Fatal(err)
	}
	if store.Contains(address) {
		t.Fatal("removed IP remains trusted")
	}
}

func TestTrustedIPStoreRejectsInvalidOrInsecureFiles(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	invalid := filepath.Join(directory, "invalid")
	if err := os.WriteFile(invalid, []byte("not-an-ip\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := newTrustedIPStore(invalid); err == nil {
		t.Fatal("invalid trusted IP file was accepted")
	}
	insecure := filepath.Join(directory, "insecure")
	if err := os.WriteFile(insecure, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := newTrustedIPStore(insecure); err == nil {
		t.Fatal("insecure trusted IP file was accepted")
	}
}

func TestClientIPAddressPrefersValidCloudflareAddressAndFallsBack(t *testing.T) {
	t.Parallel()
	cloudflareRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	cloudflareRequest.RemoteAddr = "127.0.0.1:1234"
	cloudflareRequest.Header.Set("CF-Connecting-IP", "::ffff:203.0.113.7")
	address, err := clientIPAddress(cloudflareRequest)
	if err != nil || address.String() != "203.0.113.7" {
		t.Fatalf("Cloudflare address = %q error=%v", address, err)
	}

	fallbackRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	fallbackRequest.RemoteAddr = "127.0.0.1:4321"
	fallbackRequest.Header.Set("CF-Connecting-IP", "invalid")
	address, err = clientIPAddress(fallbackRequest)
	if err != nil || address.String() != "127.0.0.1" {
		t.Fatalf("fallback address = %q error=%v", address, err)
	}
}
