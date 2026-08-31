package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type trustedIPStore struct {
	path      string
	mutex     sync.RWMutex
	addresses map[netip.Addr]struct{}
}

func newTrustedIPStore(path string) (*trustedIPStore, error) {
	if path == "" {
		return nil, nil
	}
	if err := ensureTrustedIPFile(path); err != nil {
		return nil, err
	}
	addresses, err := readTrustedIPs(path)
	if err != nil {
		return nil, err
	}
	return &trustedIPStore{path: path, addresses: addresses}, nil
}

func ensureTrustedIPFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create trusted IP directory: %w", err)
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		file, createErr := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if createErr != nil {
			return fmt.Errorf("create trusted IP file: %w", createErr)
		}
		return file.Close()
	}
	if err != nil {
		return fmt.Errorf("inspect trusted IP file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("trusted IP file must be a regular file")
	}
	if info.Mode().Perm() != 0o600 {
		return fmt.Errorf("trusted IP file permissions are %04o, want 0600", info.Mode().Perm())
	}
	return nil
}

func readTrustedIPs(path string) (map[netip.Addr]struct{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open trusted IP file: %w", err)
	}
	defer file.Close()

	addresses := make(map[netip.Addr]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		value := strings.TrimSpace(scanner.Text())
		if value == "" || strings.HasPrefix(value, "#") {
			continue
		}
		address, parseErr := parseIPAddress(value)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid trusted IP %q: %w", value, parseErr)
		}
		addresses[address] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read trusted IP file: %w", err)
	}
	return addresses, nil
}

func (store *trustedIPStore) Contains(address netip.Addr) bool {
	if store == nil || !address.IsValid() {
		return false
	}
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	_, exists := store.addresses[address.Unmap()]
	return exists
}

func (store *trustedIPStore) Count() int {
	if store == nil {
		return 0
	}
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	return len(store.addresses)
}

func (store *trustedIPStore) Add(address netip.Addr) error {
	if store == nil {
		return errors.New("trusted IP storage is not configured")
	}
	address = address.Unmap()
	if !address.IsValid() {
		return errors.New("client IP is invalid")
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	_, existed := store.addresses[address]
	store.addresses[address] = struct{}{}
	if err := store.writeLocked(); err != nil {
		if !existed {
			delete(store.addresses, address)
		}
		return err
	}
	return nil
}

func (store *trustedIPStore) Remove(address netip.Addr) error {
	if store == nil {
		return errors.New("trusted IP storage is not configured")
	}
	address = address.Unmap()
	if !address.IsValid() {
		return errors.New("client IP is invalid")
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	_, existed := store.addresses[address]
	delete(store.addresses, address)
	if err := store.writeLocked(); err != nil {
		if existed {
			store.addresses[address] = struct{}{}
		}
		return err
	}
	return nil
}

func (store *trustedIPStore) writeLocked() error {
	values := make([]string, 0, len(store.addresses))
	for address := range store.addresses {
		values = append(values, address.String())
	}
	sort.Strings(values)
	content := ""
	if len(values) > 0 {
		content = strings.Join(values, "\n") + "\n"
	}

	temporary, err := os.CreateTemp(filepath.Dir(store.path), ".trusted-ips-*")
	if err != nil {
		return fmt.Errorf("create trusted IP update: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("secure trusted IP update: %w", err)
	}
	if _, err := temporary.WriteString(content); err != nil {
		temporary.Close()
		return fmt.Errorf("write trusted IP update: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync trusted IP update: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close trusted IP update: %w", err)
	}
	if err := os.Rename(temporaryPath, store.path); err != nil {
		return fmt.Errorf("install trusted IP update: %w", err)
	}
	return nil
}

func clientIPAddress(request *http.Request) (netip.Addr, error) {
	if forwarded := strings.TrimSpace(request.Header.Get("CF-Connecting-IP")); forwarded != "" {
		if address, err := parseIPAddress(forwarded); err == nil {
			return address, nil
		}
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		host = request.RemoteAddr
	}
	return parseIPAddress(strings.TrimSpace(host))
}

func parseIPAddress(value string) (netip.Addr, error) {
	address, err := netip.ParseAddr(value)
	if err != nil {
		return netip.Addr{}, errors.New("not a valid IP address")
	}
	return address.Unmap(), nil
}
