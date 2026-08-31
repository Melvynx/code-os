package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

type gatewayTargetContextKey struct{}
type artifactBypassContextKey struct{}

var bypassableArtifactExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".webp": true, ".avif": true, ".svg": true,
}

func (server HTTPServer) hostRouter(dashboard http.Handler, auth *authenticator) http.Handler {
	apps := server.gatewayRoutes(auth, server.appGateway())

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		host := requestHostname(request)
		if port, ok := portFromPublicHost(server.Config.PublicPortHost, host); ok {
			ctx := context.WithValue(request.Context(), gatewayTargetContextKey{}, port)
			apps.ServeHTTP(response, request.WithContext(ctx))
			return
		}
		dashboard.ServeHTTP(response, request)
	})
}

func (server HTTPServer) gatewayRoutes(auth *authenticator, protected http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /_code-os/login", auth.gatewayLoginPage)
	mux.HandleFunc("GET /_code-os/login.css", auth.loginStyles)
	mux.HandleFunc("GET /_code-os/favicon.svg", auth.loginFavicon)
	mux.HandleFunc("POST /_code-os/auth/login", auth.gatewayLogin)
	mux.HandleFunc("POST /_code-os/auth/logout", auth.gatewayLogout)
	mux.HandleFunc("GET /_code-os/trust-ip", auth.gatewayTrustIPPage)
	mux.HandleFunc("POST /_code-os/auth/trust-ip", auth.gatewayTrustIP)
	mux.Handle("/", auth.protectGateway(protected))
	return gatewayAuthSecurityHeaders(mux)
}

func (server HTTPServer) appGateway() http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		port, ok := request.Context().Value(gatewayTargetContextKey{}).(int)
		if !ok || !server.Service.IsHealthyApplicationPort(port) {
			http.Error(response, "Development application is not available", http.StatusNotFound)
			return
		}
		address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
		proxyToLoopback(address, requestHostname(request)).ServeHTTP(response, request)
	})
}

func (auth *authenticator) hasValidArtifactBypass(request *http.Request) bool {
	if auth.bypassKey == "" || (request.Method != http.MethodGet && request.Method != http.MethodHead) {
		return false
	}
	if !bypassableArtifactExtensions[strings.ToLower(filepath.Ext(request.URL.Path))] {
		return false
	}
	return constantTimeEqual(request.URL.Query().Get("bp"), auth.bypassKey)
}

func proxyToLoopback(address, publicHost string) http.Handler {
	target := &url.URL{Scheme: "http", Host: address}
	return &httputil.ReverseProxy{
		Rewrite: func(proxyRequest *httputil.ProxyRequest) {
			forwardedProtocol := externalProtocol(proxyRequest.In)
			proxyRequest.SetURL(target)
			proxyRequest.SetXForwarded()
			proxyRequest.Out.Host = "localhost:" + target.Port()
			proxyRequest.Out.Header.Set("X-Forwarded-Host", publicHost)
			proxyRequest.Out.Header.Set("X-Forwarded-Proto", forwardedProtocol)
		},
		ModifyResponse: func(response *http.Response) error {
			rewriteProxyLocation(response, publicHost, externalProtocol(response.Request))
			if artifactBypass, _ := response.Request.Context().Value(artifactBypassContextKey{}).(bool); artifactBypass {
				response.Header.Set("Cache-Control", "private, no-store")
				response.Header.Set("Pragma", "no-cache")
				response.Header.Set("Referrer-Policy", "no-referrer")
				if strings.EqualFold(filepath.Ext(response.Request.URL.Path), ".svg") {
					response.Header.Set("Content-Security-Policy", "sandbox; default-src 'none'; style-src 'unsafe-inline'; img-src data:")
				}
			}
			return nil
		},
		ErrorHandler: func(response http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(response, "Development application is unavailable", http.StatusBadGateway)
		},
	}
}

func rewriteProxyLocation(response *http.Response, publicHost, protocol string) {
	location := response.Header.Get("Location")
	if location == "" {
		return
	}
	parsed, err := url.Parse(location)
	if err != nil || !parsed.IsAbs() {
		return
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "localhost" && host != "127.0.0.1" && host != "::1" {
		return
	}
	parsed.Scheme = protocol
	parsed.Host = publicHost
	response.Header.Set("Location", parsed.String())
}

func portFromPublicHost(template, hostname string) (int, bool) {
	before, after, ok := strings.Cut(strings.ToLower(strings.TrimSpace(template)), "{port}")
	if !ok || strings.Contains(after, "{port}") {
		return 0, false
	}
	hostname = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(hostname), "."))
	if !strings.HasPrefix(hostname, before) || !strings.HasSuffix(hostname, after) {
		return 0, false
	}
	portText := strings.TrimSuffix(strings.TrimPrefix(hostname, before), after)
	if portText == "" {
		return 0, false
	}
	port, err := strconv.Atoi(portText)
	return port, err == nil && port >= 1024 && port <= 65535
}

func requestHostname(request *http.Request) string {
	host := strings.TrimSpace(request.Host)
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		return strings.Trim(hostname, "[]")
	}
	return strings.Trim(host, "[]")
}

func externalProtocol(request *http.Request) string {
	protocol := strings.TrimSpace(strings.Split(request.Header.Get("X-Forwarded-Proto"), ",")[0])
	if strings.EqualFold(protocol, "https") {
		return "https"
	}
	if request.TLS != nil {
		return "https"
	}
	return "http"
}

func gatewayAuthSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/_code-os/") {
			response.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; frame-ancestors 'none'")
			response.Header().Set("Cache-Control", "no-store")
			response.Header().Set("Referrer-Policy", "no-referrer")
			response.Header().Set("X-Content-Type-Options", "nosniff")
			response.Header().Set("X-Frame-Options", "DENY")
		}
		next.ServeHTTP(response, request)
	})
}

func constantTimeEqual(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	var result byte
	for index := range left {
		result |= left[index] ^ right[index]
	}
	return result == 0
}
