package server

import (
	"net/http"
	"net/url"
)

type trustedIPStatus struct {
	CurrentIP  string `json:"currentIP"`
	Trusted    bool   `json:"trusted"`
	Configured bool   `json:"configured"`
	Count      int    `json:"count"`
}

type trustIPPageData struct {
	StylesheetPath string
	FormAction     string
	Next           string
	ClientIP       string
	Context        string
}

func (auth *authenticator) trustIPPage(response http.ResponseWriter, request *http.Request) {
	auth.trustIPPageFor(response, request, dashboardAuthSurface)
}

func (auth *authenticator) gatewayTrustIPPage(response http.ResponseWriter, request *http.Request) {
	auth.trustIPPageFor(response, request, gatewayAuthSurface)
}

func (auth *authenticator) trustIPPageFor(response http.ResponseWriter, request *http.Request, surface authSurface) {
	nextPath := safeNextPath(request.URL.Query().Get("next"))
	if !auth.hasValidSessionFor(request, surface.CookieName) {
		http.Redirect(response, request, surface.LoginPath+"?next="+url.QueryEscape(nextPath), http.StatusSeeOther)
		return
	}
	if auth.trustedIPs == nil {
		http.Redirect(response, request, nextPath, http.StatusSeeOther)
		return
	}
	address, err := clientIPAddress(request)
	if err != nil {
		http.Error(response, "Unable to determine your IP address", http.StatusBadRequest)
		return
	}
	if auth.trustedIPs.Contains(address) {
		http.Redirect(response, request, nextPath, http.StatusSeeOther)
		return
	}
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Header().Set("Cache-Control", "no-store")
	data := trustIPPageData{
		StylesheetPath: surface.StylesheetPath, FormAction: surface.TrustAction,
		Next: nextPath, ClientIP: address.String(), Context: surface.Context,
	}
	if err := auth.templates.ExecuteTemplate(response, "trust-ip.html", data); err != nil {
		http.Error(response, "Unable to render IP trust confirmation", http.StatusInternalServerError)
	}
}

func (auth *authenticator) trustIP(response http.ResponseWriter, request *http.Request) {
	auth.trustIPFor(response, request, dashboardAuthSurface)
}

func (auth *authenticator) gatewayTrustIP(response http.ResponseWriter, request *http.Request) {
	auth.trustIPFor(response, request, gatewayAuthSurface)
}

func (auth *authenticator) trustedIPStatus(response http.ResponseWriter, request *http.Request) {
	address, err := clientIPAddress(request)
	if err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "unable to determine client IP"})
		return
	}
	writeJSON(response, http.StatusOK, trustedIPStatus{
		CurrentIP: address.String(), Trusted: auth.trustedIPs.Contains(address),
		Configured: auth.trustedIPs != nil, Count: auth.trustedIPs.Count(),
	})
}

func (auth *authenticator) untrustIP(response http.ResponseWriter, request *http.Request) {
	if !isSameOrigin(request) {
		writeJSON(response, http.StatusForbidden, map[string]string{"error": "same-origin request required"})
		return
	}
	address, err := clientIPAddress(request)
	if err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "unable to determine client IP"})
		return
	}
	if err := auth.trustedIPs.Remove(address); err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "could not revoke trusted IP"})
		return
	}
	writeJSON(response, http.StatusOK, trustedIPStatus{
		CurrentIP: address.String(), Trusted: false,
		Configured: auth.trustedIPs != nil, Count: auth.trustedIPs.Count(),
	})
}

func (auth *authenticator) trustIPFor(response http.ResponseWriter, request *http.Request, surface authSurface) {
	if !auth.hasValidSessionFor(request, surface.CookieName) {
		http.Error(response, "Authentication required", http.StatusUnauthorized)
		return
	}
	if !isSameOrigin(request) {
		http.Error(response, "Same-origin request required", http.StatusForbidden)
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, 8<<10)
	if err := request.ParseForm(); err != nil {
		http.Error(response, "Invalid form", http.StatusBadRequest)
		return
	}
	nextPath := safeNextPath(request.FormValue("next"))
	address, err := clientIPAddress(request)
	if err != nil {
		http.Error(response, "Unable to determine your IP address", http.StatusBadRequest)
		return
	}
	if err := auth.trustedIPs.Add(address); err != nil {
		http.Error(response, "Unable to trust this IP address", http.StatusInternalServerError)
		return
	}
	http.Redirect(response, request, nextPath, http.StatusSeeOther)
}
