package server

import (
	"net/http"

	"github.com/melvynx/code-os/internal/diagnostics"
	"github.com/melvynx/code-os/internal/skills"
)

func (server HTTPServer) status(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, diagnostics.Inspect(server.Config, diagnostics.Options{ConfigPath: server.ConfigPath}))
}

func (server HTTPServer) skillsSyncStatus(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, skills.Inspect(server.Config.Skills))
}

func (server HTTPServer) runSkillsSync(response http.ResponseWriter, request *http.Request) {
	if !isSameOrigin(request) {
		writeJSON(response, http.StatusForbidden, map[string]string{"error": "same-origin request required"})
		return
	}
	result, err := skills.Sync(server.Config.Skills)
	if err != nil {
		writeJSON(response, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(response, http.StatusOK, map[string]any{
		"result": result,
		"status": skills.Inspect(server.Config.Skills),
	})
}
