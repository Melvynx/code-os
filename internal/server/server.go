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

	"github.com/melvynx/stackenv/internal/config"
	"github.com/melvynx/stackenv/internal/dashboard"
	"github.com/melvynx/stackenv/internal/model"
)

//go:embed web/*
var webAssets embed.FS

type HTTPServer struct {
	Config  config.Config
	Service *Service
	Logger  *slog.Logger
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
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/overview", server.overview)
	protected.HandleFunc("GET /api/snapshot", server.snapshot)
	protected.HandleFunc("POST /api/refresh", server.refresh)
	protected.HandleFunc("GET /api/health", server.health)
	protected.HandleFunc("GET /media/{id}", server.media)
	protected.Handle("GET /", singlePageApplication(dashboardAssets))
	var handler http.Handler = protected
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
		auth, err := newAuthenticator(loginAssets, server.Config.Auth.Username, secret, bypassKey)
		if err != nil {
			return nil, err
		}
		public := http.NewServeMux()
		public.HandleFunc("GET /login", auth.loginPage)
		public.HandleFunc("GET /login.css", auth.loginStyles)
		public.HandleFunc("POST /auth/login", auth.login)
		public.HandleFunc("POST /auth/logout", auth.logout)
		public.Handle("/", auth.protect(protected))
		handler = public
	}
	return server.securityHeaders(server.requestLog(handler)), nil
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
