package server

import (
	"io/fs"
	"net/http"
	"strings"
)

func singlePageApplication(assets fs.FS) http.Handler {
	files := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assetPath := strings.TrimPrefix(request.URL.Path, "/")
		if isReservedPath(assetPath) {
			http.NotFound(response, request)
			return
		}
		if assetPath == "" {
			files.ServeHTTP(response, request)
			return
		}
		info, err := fs.Stat(assets, assetPath)
		if err == nil && !info.IsDir() {
			files.ServeHTTP(response, request)
			return
		}
		fallback := request.Clone(request.Context())
		fallback.URL.Path = "/"
		fallback.URL.RawPath = ""
		files.ServeHTTP(response, fallback)
	})
}

func isReservedPath(path string) bool {
	for _, prefix := range []string{"api/", "auth/", "media/"} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
