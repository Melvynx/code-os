package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/melvynx/code-os/internal/config"
	"github.com/melvynx/code-os/internal/dashboard"
	"github.com/melvynx/code-os/internal/model"
	"github.com/melvynx/code-os/internal/site"
)

//go:embed web/*
var webAssets embed.FS

type HTTPServer struct {
	Config     config.Config
	ConfigPath string
	Service    *Service
	Logger     *slog.Logger
}

func (server HTTPServer) Handler() (http.Handler, error) {
	loginAssets, err := fs.Sub(webAssets, "web")
	if err != nil {
		return nil, fmt.Errorf("load embedded login: %w", err)
	}
	dashboardAssets, err := dashboard.Assets()
	if err != nil {
		return nil, err
	}
	siteAssets, err := site.Assets()
	if err != nil {
		return nil, err
	}
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/overview", server.overview)
	protected.HandleFunc("GET /api/snapshot", server.snapshot)
	protected.HandleFunc("POST /api/refresh", server.refresh)
	protected.HandleFunc("POST /api/applications/{id}/stop", server.stopApplication)
	protected.HandleFunc("POST /api/agents/{id}/terminate", server.terminateAgent)
	protected.HandleFunc("GET /api/health", server.health)
	protected.HandleFunc("GET /api/settings", server.settings)
	protected.HandleFunc("PUT /api/settings", server.updateSettings)
	protected.HandleFunc("GET /media/{id}", server.media)
	protected.Handle("GET /app/", http.StripPrefix("/app", singlePageApplication(dashboardAssets)))
	protected.HandleFunc("GET /app", func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, "/app/", http.StatusPermanentRedirect)
	})
	publicSite := publicSecurityHeaders(http.FileServer(http.FS(siteAssets)))
	root := http.NewServeMux()
	var auth *authenticator
	if server.Config.Auth.PasswordFile != "" {
		password, err := os.ReadFile(server.Config.Auth.PasswordFile)
		if err != nil {
			return nil, fmt.Errorf("read dashboard password: %w", err)
		}
		secret := strings.TrimSpace(string(password))
		if secret == "" {
			return nil, fmt.Errorf("dashboard password file is empty")
		}
		bypassKey := ""
		if server.Config.Auth.BypassKeyFile != "" {
			bypass, err := os.ReadFile(server.Config.Auth.BypassKeyFile)
			if err != nil {
				return nil, fmt.Errorf("read media bypass key: %w", err)
			}
			bypassKey = strings.TrimSpace(string(bypass))
			if bypassKey == "" {
				return nil, fmt.Errorf("media bypass key file is empty")
			}
		}
		sessionKey, err := os.ReadFile(server.Config.Auth.SessionKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read session signing key: %w", err)
		}
		auth, err = newAuthenticator(loginAssets, server.Config.Auth.Username, secret, bypassKey, server.Config.Auth.TrustedIPsFile, sessionKey)
		if err != nil {
			return nil, err
		}
		root.HandleFunc("GET /login", auth.loginPage)
		root.HandleFunc("GET /login.css", auth.loginStyles)
		root.HandleFunc("GET /favicon.ico", auth.loginFavicon)
		root.HandleFunc("POST /auth/login", auth.login)
		root.HandleFunc("POST /auth/logout", auth.logout)
		root.HandleFunc("GET /trust-ip", auth.trustIPPage)
		root.HandleFunc("POST /auth/trust-ip", auth.trustIP)
		root.Handle("GET /api/trusted-ip", auth.protect(http.HandlerFunc(auth.trustedIPStatus)))
		root.Handle("POST /api/trusted-ip", auth.protect(http.HandlerFunc(auth.trustCurrentIP)))
		root.Handle("DELETE /api/trusted-ip", auth.protect(http.HandlerFunc(auth.untrustIP)))
		if server.Config.FilesRoot != "" {
			files := auth.protectFiles(http.HandlerFunc(server.privateFile))
			root.Handle("GET /files/{path...}", files)
			root.Handle("HEAD /files/{path...}", files)
		}
		root.Handle("/app", auth.protect(protected))
		root.Handle("/app/", auth.protect(protected))
		root.Handle("/api/", auth.protect(protected))
		root.Handle("/media/", auth.protect(protected))
	} else {
		root.Handle("/app", protected)
		root.Handle("/app/", protected)
		root.Handle("/api/", protected)
		root.Handle("/media/", protected)
	}
	root.Handle("/", http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" && (request.Method == http.MethodGet || request.Method == http.MethodHead) {
			http.Redirect(response, request, "/app/", http.StatusTemporaryRedirect)
			return
		}
		publicSite.ServeHTTP(response, request)
	}))
	dashboardHandler := server.securityHeaders(server.requestLog(root))
	if auth != nil && server.Config.PublicPortHost != "" {
		return server.hostRouter(dashboardHandler, auth), nil
	}
	return dashboardHandler, nil
}

func publicSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self' 'unsafe-inline'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(response, request)
	})
}

func (server HTTPServer) overview(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, model.Summarize(server.Service.Snapshot()))
}

func (server HTTPServer) snapshot(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, server.Service.Snapshot())
}

func (server HTTPServer) refresh(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := contextWithTimeout(request, 30*time.Second)
	defer cancel()
	writeJSON(response, http.StatusOK, server.Service.Refresh(ctx))
}

func (server HTTPServer) stopApplication(response http.ResponseWriter, request *http.Request) {
	if !isSameOrigin(request) {
		writeJSON(response, http.StatusForbidden, map[string]string{"error": "same-origin request required"})
		return
	}
	ctx, cancel := contextWithTimeout(request, 30*time.Second)
	defer cancel()
	if err := server.Service.StopApplication(ctx, request.PathValue("id")); err != nil {
		writeJSON(response, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(response, http.StatusOK, server.Service.Refresh(ctx))
}

func (server HTTPServer) terminateAgent(response http.ResponseWriter, request *http.Request) {
	if !isSameOrigin(request) {
		writeJSON(response, http.StatusForbidden, map[string]string{"error": "same-origin request required"})
		return
	}
	if err := server.Service.TerminateAgent(request.PathValue("id")); err != nil {
		writeJSON(response, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(response, http.StatusAccepted, map[string]bool{"ok": true})
}

func (server HTTPServer) health(response http.ResponseWriter, _ *http.Request) {
	snapshot := server.Service.Snapshot()
	writeJSON(response, http.StatusOK, map[string]any{
		"ok": true, "environment": server.Config.EnvironmentName,
		"generatedAt": snapshot.GeneratedAt,
	})
}

func (server HTTPServer) media(response http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	path, exists := server.Service.MediaPath(id)
	if !exists {
		http.NotFound(response, request)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		http.NotFound(response, request)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(response, request)
		return
	}
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if contentType != "" {
		response.Header().Set("Content-Type", contentType)
	}
	if strings.EqualFold(filepath.Ext(path), ".svg") {
		response.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; style-src 'unsafe-inline'; img-src data:")
	}
	if isMediaBypass(request) {
		response.Header().Set("Cache-Control", "private, no-store")
		response.Header().Set("Pragma", "no-cache")
	} else {
		response.Header().Set("Cache-Control", "private, max-age=300")
	}
	http.ServeContent(response, request, info.Name(), info.ModTime(), file)
}

func (server HTTPServer) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; frame-ancestors 'none'")
		response.Header().Set("Referrer-Policy", "no-referrer")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(response, request)
	})
}

func (server HTTPServer) requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(response, request)
		if server.Logger != nil && strings.HasPrefix(request.URL.Path, "/api/") {
			server.Logger.Debug("http request", "method", request.Method, "path", request.URL.Path, "duration", time.Since(started))
		}
	})
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}
