package server

import "net/http"

func (server HTTPServer) resources(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, server.Service.ResourceHistory())
}
