package server

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (server HTTPServer) privateFile(response http.ResponseWriter, request *http.Request) {
	path, ok := resolvePrivateFile(server.Config.FilesRoot, request.PathValue("path"))
	if !ok {
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

	extension := strings.ToLower(filepath.Ext(path))
	if contentType := mime.TypeByExtension(extension); contentType != "" {
		response.Header().Set("Content-Type", contentType)
	}
	response.Header().Set("Cache-Control", "private, no-store")
	response.Header().Set("Pragma", "no-cache")
	response.Header().Set("Referrer-Policy", "no-referrer")
	response.Header().Set("X-Content-Type-Options", "nosniff")
	response.Header().Set("Content-Disposition", "inline")
	if bypass, _ := request.Context().Value(artifactBypassContextKey{}).(bool); bypass {
		response.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
	}
	if extension == ".svg" {
		response.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; style-src 'unsafe-inline'; img-src data:")
	}
	http.ServeContent(response, request, info.Name(), info.ModTime(), file)
}

func resolvePrivateFile(root, requested string) (string, bool) {
	if strings.ContainsRune(requested, '\x00') {
		return "", false
	}
	clean := filepath.Clean(filepath.FromSlash("/" + requested))
	clean = strings.TrimPrefix(clean, string(filepath.Separator))
	if clean == "" || clean == "." {
		clean = "index.html"
	}
	for _, segment := range strings.Split(clean, string(filepath.Separator)) {
		if segment == ".." || strings.HasPrefix(segment, ".") {
			return "", false
		}
	}
	candidate := filepath.Join(root, clean)
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		candidate = filepath.Join(candidate, "index.html")
	}
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", false
	}
	candidateResolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", false
	}
	relative, err := filepath.Rel(rootResolved, candidateResolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", false
	}
	return candidateResolved, true
}
