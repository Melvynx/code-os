package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/melvynx/code-os/internal/model"
)

const (
	iconCacheTTL   = 15 * time.Minute
	iconFetchLimit = 256 << 10
)

var (
	defaultIconCache  = newIconCache()
	htmlIconHref      = regexp.MustCompile(`(?i)<link[^>]+>`)
	htmlIconRel       = regexp.MustCompile(`(?i)rel=["']([^"']+)["']`)
	htmlIconReference = regexp.MustCompile(`(?i)href=["']([^"']+)["']`)
	htmlIconType      = regexp.MustCompile(`(?i)type=["']([^"']+)["']`)
	htmlIconSizes     = regexp.MustCompile(`(?i)sizes=["']([^"']+)["']`)
)

type cachedIcon struct {
	body        []byte
	contentType string
	expires     time.Time
	ok          bool
}

type iconCache struct {
	mutex sync.Mutex
	items map[string]cachedIcon
}

func newIconCache() *iconCache {
	return &iconCache{items: map[string]cachedIcon{}}
}

func (cache *iconCache) get(key string) (cachedIcon, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	item, exists := cache.items[key]
	if !exists || time.Now().After(item.expires) {
		return cachedIcon{}, false
	}
	return item, true
}

func (cache *iconCache) set(key string, item cachedIcon) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.items[key] = item
}

func (server HTTPServer) iconStore() *iconCache {
	if server.Icons != nil {
		return server.Icons
	}
	return defaultIconCache
}

func (server HTTPServer) iconHTTPClient() *http.Client {
	if server.IconClient != nil {
		return server.IconClient
	}
	return &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(request *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func (server HTTPServer) applicationOrigin(application model.Application) string {
	if server.IconOrigin != nil {
		return server.IconOrigin(application)
	}
	if application.Port <= 0 {
		return ""
	}
	return fmt.Sprintf("http://127.0.0.1:%d", application.Port)
}

func (server HTTPServer) applicationIcon(response http.ResponseWriter, request *http.Request) {
	application, exists := server.Service.Application(request.PathValue("id"))
	if !exists {
		http.NotFound(response, request)
		return
	}
	key := iconCacheKey(application)
	if cached, ok := server.iconStore().get(key); ok {
		writeCachedIcon(response, cached)
		return
	}
	body, contentType, err := server.fetchApplicationIcon(application)
	cached := cachedIcon{body: body, contentType: contentType, expires: time.Now().Add(iconCacheTTL), ok: err == nil}
	server.iconStore().set(key, cached)
	writeCachedIcon(response, cached)
}

func writeCachedIcon(response http.ResponseWriter, cached cachedIcon) {
	if !cached.ok {
		http.Error(response, "application icon not found", http.StatusNotFound)
		return
	}
	response.Header().Set("Content-Type", cached.contentType)
	response.Header().Set("Cache-Control", "private, max-age=300")
	response.Header().Set("X-Content-Type-Options", "nosniff")
	if strings.Contains(cached.contentType, "svg") {
		response.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; style-src 'unsafe-inline'")
	}
	_, _ = response.Write(cached.body)
}

func iconCacheKey(application model.Application) string {
	return strings.Join([]string{application.ID, application.State, strconv.Itoa(application.Port), application.Domain}, "|")
}

func (server HTTPServer) fetchApplicationIcon(application model.Application) ([]byte, string, error) {
	client := server.iconHTTPClient()
	for _, candidate := range iconCandidates(application, server.applicationOrigin(application)) {
		if !iconTargetAllowed(application, candidate, server.applicationOrigin(application)) {
			continue
		}
		body, contentType, err := fetchIconURL(client, candidate)
		if err != nil {
			continue
		}
		if strings.Contains(contentType, "text/html") {
			for _, href := range parseHTMLIconHrefs(body, candidate) {
				if !iconTargetAllowed(application, href, server.applicationOrigin(application)) {
					continue
				}
				icon, iconType, iconErr := fetchIconURL(client, href)
				if iconErr == nil && isIconMedia(iconType, icon) {
					return icon, iconType, nil
				}
			}
			continue
		}
		if isIconMedia(contentType, body) {
			return body, iconContentType(contentType, body), nil
		}
	}
	return nil, "", fmt.Errorf("application icon not found")
}

func iconCandidates(application model.Application, origin string) []string {
	var candidates []string
	if origin != "" {
		candidates = append(candidates,
			origin+"/favicon.svg",
			origin+"/favicon-32x32.png",
			origin+"/apple-icon.png",
			origin+"/apple-touch-icon.png",
			origin+"/favicon-dev.png",
			origin+"/favicon.ico",
			origin+"/",
		)
	}
	if application.Domain == "" {
		return candidates
	}
	domain := strings.ToLower(application.Domain)
	return append(candidates,
		"https://"+domain+"/favicon.svg",
		"https://"+domain+"/favicon.ico",
		"https://icons.duckduckgo.com/ip3/"+domain+".ico",
		"https://www.google.com/s2/favicons?domain="+url.QueryEscape(domain)+"&sz=64",
	)
}

func iconTargetAllowed(application model.Application, raw, origin string) bool {
	target, err := url.Parse(raw)
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") {
		return false
	}
	host := strings.ToLower(target.Hostname())
	if origin != "" {
		if parsedOrigin, originErr := url.Parse(origin); originErr == nil && sameOrigin(parsedOrigin, target) {
			return true
		}
	}
	if host == "127.0.0.1" || host == "localhost" || host == "::1" {
		return localIconPortAllowed(application, target)
	}
	if application.Domain != "" && host == strings.ToLower(application.Domain) {
		return target.Scheme == "https"
	}
	return isIconCDN(target)
}

func localIconPortAllowed(application model.Application, target *url.URL) bool {
	if application.Port <= 0 {
		return false
	}
	port := target.Port()
	if port == "" {
		if target.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return port == strconv.Itoa(application.Port)
}

func sameOrigin(left, right *url.URL) bool {
	return strings.EqualFold(left.Scheme, right.Scheme) && strings.EqualFold(left.Host, right.Host)
}

func isIconCDN(target *url.URL) bool {
	if target.Scheme != "https" {
		return false
	}
	host := strings.ToLower(target.Hostname())
	if host == "icons.duckduckgo.com" && strings.HasPrefix(target.Path, "/ip3/") {
		return true
	}
	return host == "www.google.com" && strings.HasPrefix(target.Path, "/s2/favicons")
}

func fetchIconURL(client *http.Client, raw string) ([]byte, string, error) {
	request, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return nil, "", err
	}
	request.Header.Set("Accept", "image/svg+xml,image/png,image/x-icon,image/*,text/html;q=0.2")
	request.Header.Set("User-Agent", "Code-OS/1.0")
	response, err := client.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("status %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, iconFetchLimit+1))
	if err != nil {
		return nil, "", err
	}
	if len(body) == 0 || len(body) > iconFetchLimit {
		return nil, "", fmt.Errorf("invalid icon size")
	}
	contentType := response.Header.Get("Content-Type")
	if mediaType, _, _ := strings.Cut(contentType, ";"); mediaType != "" {
		contentType = strings.TrimSpace(mediaType)
	}
	return body, contentType, nil
}

func parseHTMLIconHrefs(body []byte, pageURL string) []string {
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil
	}
	type ranked struct {
		href  string
		score int
	}
	var found []ranked
	for _, tag := range htmlIconHref.FindAll(body, 16) {
		relMatch := htmlIconRel.FindSubmatch(tag)
		if len(relMatch) < 2 {
			continue
		}
		rel := strings.ToLower(string(relMatch[1]))
		if !strings.Contains(rel, "icon") {
			continue
		}
		hrefMatch := htmlIconReference.FindSubmatch(tag)
		if len(hrefMatch) < 2 {
			continue
		}
		resolved, err := base.Parse(string(hrefMatch[1]))
		if err != nil {
			continue
		}
		score := 10
		if strings.Contains(rel, "apple-touch") {
			score += 8
		}
		typeMatch := htmlIconType.FindSubmatch(tag)
		if len(typeMatch) >= 2 {
			switch strings.ToLower(string(typeMatch[1])) {
			case "image/svg+xml":
				score += 20
			case "image/png":
				score += 12
			}
		}
		if sizes := htmlIconSizes.FindSubmatch(tag); len(sizes) >= 2 {
			if width, _, ok := strings.Cut(string(sizes[1]), "x"); ok {
				if parsed, err := strconv.Atoi(width); err == nil && parsed >= 32 {
					score += 6
				}
			}
		}
		found = append(found, ranked{href: resolved.String(), score: score})
	}
	for index := 0; index < len(found); index++ {
		for next := index + 1; next < len(found); next++ {
			if found[next].score > found[index].score {
				found[index], found[next] = found[next], found[index]
			}
		}
	}
	hrefs := make([]string, 0, len(found))
	seen := map[string]struct{}{}
	for _, item := range found {
		if _, exists := seen[item.href]; exists {
			continue
		}
		seen[item.href] = struct{}{}
		hrefs = append(hrefs, item.href)
	}
	return hrefs
}

func isIconMedia(contentType string, body []byte) bool {
	switch {
	case strings.Contains(contentType, "image/"), strings.Contains(contentType, "icon"):
		return true
	case strings.HasPrefix(strings.ToLower(string(body)), "<svg"), looksLikePNG(body), looksLikeICO(body):
		return true
	default:
		return false
	}
}

func iconContentType(contentType string, body []byte) string {
	if strings.HasPrefix(contentType, "image/") || strings.Contains(contentType, "icon") {
		return contentType
	}
	if strings.HasPrefix(strings.ToLower(string(body)), "<svg") {
		return "image/svg+xml"
	}
	if looksLikePNG(body) {
		return "image/png"
	}
	if looksLikeICO(body) {
		return "image/x-icon"
	}
	return "application/octet-stream"
}

func looksLikePNG(body []byte) bool {
	return len(body) >= 8 && string(body[:8]) == "\x89PNG\r\n\x1a\n"
}

func looksLikeICO(body []byte) bool {
	return len(body) >= 4 && body[0] == 0 && body[1] == 0 && (body[2] == 1 || body[2] == 2)
}
