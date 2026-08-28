package server

import (
	"crypto/subtle"
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
	assets, err := fs.Sub(webAssets, "web")
	if err != nil {
		return nil, fmt.Errorf("load embedded dashboard: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/overview", server.overview)
	mux.HandleFunc("GET /api/snapshot", server.snapshot)
	mux.HandleFunc("POST /api/refresh", server.refresh)
	mux.HandleFunc("GET /api/health", server.health)
	mux.HandleFunc("GET /media/{id}", server.media)
	mux.Handle("GET /", http.FileServer(http.FS(assets)))
	handler := server.securityHeaders(server.requestLog(mux))
	if server.Config.Auth.PasswordFile != "" {
		password, err := os.ReadFile(server.Config.Auth.PasswordFile)
		if err != nil {
			return nil, fmt.Errorf("read dashboard password: %w", err)
		}
		secret := strings.TrimSpace(string(password))
		if secret == "" {
			return nil, fmt.Errorf("dashboard password file is empty")
		}
		handler = basicAuth(handler, server.Config.Auth.Username, secret)
	}
	return handler, nil
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
	response.Header().Set("Cache-Control", "private, max-age=300")
	http.ServeContent(response, request, info.Name(), info.ModTime(), file)
}

func (server HTTPServer) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; connect-src 'self'; frame-ancestors 'none'")
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

func basicAuth(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		providedUser, providedPassword, ok := request.BasicAuth()
		userMatches := subtle.ConstantTimeCompare([]byte(providedUser), []byte(username)) == 1
		passwordMatches := subtle.ConstantTimeCompare([]byte(providedPassword), []byte(password)) == 1
		if !ok || !userMatches || !passwordMatches {
			response.Header().Set("WWW-Authenticate", `Basic realm="StackEnv Command Center", charset="UTF-8"`)
			http.Error(response, "Authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(response, request)
	})
}
