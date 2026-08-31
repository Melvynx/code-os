package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	sessionCookieName        = "code-os_session"
	gatewaySessionCookieName = "code-os_gateway_session"
	sessionDuration          = 12 * time.Hour
	loginAttemptWindow       = 5 * time.Minute
	maxLoginAttempts         = 5
)

type mediaBypassContextKey struct{}

type authenticator struct {
	username        string
	password        string
	bypassKey       string
	sessionKey      []byte
	templates       *template.Template
	loginStylesheet []byte
	favicon         []byte
	trustedIPs      *trustedIPStore
	attemptsMutex   sync.Mutex
	attempts        map[string]loginAttempt
}

type loginPageData struct {
	InvalidCredentials bool
	Next               string
	StylesheetPath     string
	FaviconPath        string
	FormAction         string
	PageTitle          string
	Heading            string
	Description        string
	Context            string
}

type loginAttempt struct {
	Count       int
	WindowStart time.Time
}

type authSurface struct {
	CookieName     string
	LoginPath      string
	StylesheetPath string
	FaviconPath    string
	LoginAction    string
	LogoutPath     string
	TrustPath      string
	TrustAction    string
	PageTitle      string
	Heading        string
	Description    string
	Context        string
	SameSite       http.SameSite
	APIsUseJSON401 bool
}

var dashboardAuthSurface = authSurface{
	CookieName: sessionCookieName, LoginPath: "/login", StylesheetPath: "/login.css?v=black-grid-2", FaviconPath: "/favicon.svg",
	LoginAction: "/auth/login", LogoutPath: "/auth/logout", TrustPath: "/trust-ip", TrustAction: "/auth/trust-ip", PageTitle: "Sign in — Code OS",
	Heading: "Sign in to Code OS", Description: "Use the credentials configured on this environment.",
	Context: "Private development environment", SameSite: http.SameSiteStrictMode, APIsUseJSON401: true,
}

var gatewayAuthSurface = authSurface{
	CookieName: gatewaySessionCookieName, LoginPath: "/_code-os/login", StylesheetPath: "/_code-os/login.css?v=black-grid-2", FaviconPath: "/_code-os/favicon.svg",
	LoginAction: "/_code-os/auth/login", LogoutPath: "/_code-os/auth/logout", TrustPath: "/_code-os/trust-ip", TrustAction: "/_code-os/auth/trust-ip", PageTitle: "Unlock application — Code OS",
	Heading: "Unlock development app", Description: "Authenticate before opening this private development port.",
	Context: "Protected by Code OS", SameSite: http.SameSiteLaxMode,
}

func newAuthenticator(assets fs.FS, username, password, bypassKey, trustedIPsFile string, sessionKey []byte) (*authenticator, error) {
	templates, err := template.ParseFS(assets, "login.html", "trust-ip.html")
	if err != nil {
		return nil, fmt.Errorf("parse login page: %w", err)
	}
	loginStylesheet, err := fs.ReadFile(assets, "login.css")
	if err != nil {
		return nil, fmt.Errorf("read login stylesheet: %w", err)
	}
	favicon, err := fs.ReadFile(assets, "favicon.svg")
	if err != nil {
		return nil, fmt.Errorf("read login favicon: %w", err)
	}
	if len(sessionKey) < 32 {
		return nil, fmt.Errorf("session signing key must contain at least 32 bytes")
	}
	trustedIPs, err := newTrustedIPStore(trustedIPsFile)
	if err != nil {
		return nil, fmt.Errorf("load trusted IPs: %w", err)
	}
	return &authenticator{
		username: username, password: password, bypassKey: bypassKey,
		sessionKey: append([]byte(nil), sessionKey...), templates: templates, loginStylesheet: loginStylesheet, favicon: favicon, trustedIPs: trustedIPs,
		attempts: make(map[string]loginAttempt),
	}, nil
}

func (auth *authenticator) protect(next http.Handler) http.Handler {
	return auth.protectSurface(next, dashboardAuthSurface)
}

func (auth *authenticator) protectGateway(next http.Handler) http.Handler {
	return auth.protectSurface(next, gatewayAuthSurface)
}

func (auth *authenticator) protectFiles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if auth.hasTrustedIPAddress(request) || auth.hasValidSession(request) {
			next.ServeHTTP(response, request)
			return
		}
		if auth.hasValidArtifactBypass(request) {
			query := request.URL.Query()
			query.Del("bp")
			ctx := context.WithValue(request.Context(), artifactBypassContextKey{}, true)
			clone := request.Clone(ctx)
			clone.URL.RawQuery = query.Encode()
			next.ServeHTTP(response, clone)
			return
		}
		if request.URL.Query().Has("bp") {
			writeJSON(response, http.StatusUnauthorized, map[string]string{"error": "invalid file bypass"})
			return
		}
		nextPath := request.URL.RequestURI()
		if nextPath == "" {
			nextPath = "/"
		}
		http.Redirect(response, request, dashboardAuthSurface.LoginPath+"?next="+url.QueryEscape(nextPath), http.StatusSeeOther)
	})
}

func (auth *authenticator) protectSurface(next http.Handler, surface authSurface) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if auth.hasTrustedIPAddress(request) || auth.hasValidSessionFor(request, surface.CookieName) {
			next.ServeHTTP(response, request)
			return
		}
		if surface.CookieName == sessionCookieName && auth.hasValidMediaBypass(request) {
			ctx := context.WithValue(request.Context(), mediaBypassContextKey{}, true)
			next.ServeHTTP(response, request.WithContext(ctx))
			return
		}
		if surface.APIsUseJSON401 && (strings.HasPrefix(request.URL.Path, "/api/") || strings.HasPrefix(request.URL.Path, "/media/")) {
			writeJSON(response, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			return
		}
		nextPath := request.URL.RequestURI()
		if nextPath == "" {
			nextPath = "/"
		}
		http.Redirect(response, request, surface.LoginPath+"?next="+url.QueryEscape(nextPath), http.StatusSeeOther)
	})
}

func (auth *authenticator) loginPage(response http.ResponseWriter, request *http.Request) {
	auth.loginPageFor(response, request, dashboardAuthSurface)
}

func (auth *authenticator) gatewayLoginPage(response http.ResponseWriter, request *http.Request) {
	auth.loginPageFor(response, request, gatewayAuthSurface)
}

func (auth *authenticator) loginPageFor(response http.ResponseWriter, request *http.Request, surface authSurface) {
	nextPath := safeNextPath(request.URL.Query().Get("next"))
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Header().Set("Cache-Control", "no-store")
	if err := auth.templates.ExecuteTemplate(response, "login.html", auth.loginPageData(surface, nextPath, false)); err != nil {
		http.Error(response, "Unable to render sign in", http.StatusInternalServerError)
	}
}

func (auth *authenticator) loginStyles(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/css; charset=utf-8")
	response.Header().Set("Cache-Control", "no-store")
	_, _ = response.Write(auth.loginStylesheet)
}

func (auth *authenticator) loginFavicon(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "image/svg+xml")
	response.Header().Set("Cache-Control", "private, max-age=86400")
	_, _ = response.Write(auth.favicon)
}

func (auth *authenticator) login(response http.ResponseWriter, request *http.Request) {
	auth.loginFor(response, request, dashboardAuthSurface)
}

func (auth *authenticator) gatewayLogin(response http.ResponseWriter, request *http.Request) {
	auth.loginFor(response, request, gatewayAuthSurface)
}

func (auth *authenticator) loginFor(response http.ResponseWriter, request *http.Request, surface authSurface) {
	request.Body = http.MaxBytesReader(response, request.Body, 8<<10)
	if err := request.ParseForm(); err != nil {
		http.Error(response, "Invalid form", http.StatusBadRequest)
		return
	}
	nextPath := safeNextPath(request.FormValue("next"))
	clientKey := loginClientKey(request)
	if retryAfter, limited := auth.loginRetryAfter(clientKey); limited {
		response.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Round(time.Second).Seconds())))
		http.Error(response, "Too many sign-in attempts", http.StatusTooManyRequests)
		return
	}
	usernameMatches := subtle.ConstantTimeCompare([]byte(request.FormValue("username")), []byte(auth.username)) == 1
	passwordMatches := subtle.ConstantTimeCompare([]byte(request.FormValue("password")), []byte(auth.password)) == 1
	if !usernameMatches || !passwordMatches {
		auth.recordLoginFailure(clientKey)
		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		response.Header().Set("Cache-Control", "no-store")
		response.WriteHeader(http.StatusUnauthorized)
		_ = auth.templates.ExecuteTemplate(response, "login.html", auth.loginPageData(surface, nextPath, true))
		return
	}
	auth.clearLoginFailures(clientKey)
	auth.setSessionCookieFor(response, request, surface)
	if auth.trustedIPs != nil && !auth.hasTrustedIPAddress(request) {
		http.Redirect(response, request, surface.TrustPath+"?next="+url.QueryEscape(nextPath), http.StatusSeeOther)
		return
	}
	http.Redirect(response, request, nextPath, http.StatusSeeOther)
}

func (auth *authenticator) logout(response http.ResponseWriter, request *http.Request) {
	auth.logoutFor(response, request, dashboardAuthSurface)
}

func (auth *authenticator) gatewayLogout(response http.ResponseWriter, request *http.Request) {
	auth.logoutFor(response, request, gatewayAuthSurface)
}

func (auth *authenticator) logoutFor(response http.ResponseWriter, request *http.Request, surface authSurface) {
	http.SetCookie(response, &http.Cookie{
		Name: surface.CookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: requestIsHTTPS(request), SameSite: surface.SameSite,
	})
	http.Redirect(response, request, surface.LoginPath, http.StatusSeeOther)
}

func (auth *authenticator) hasValidMediaBypass(request *http.Request) bool {
	if auth.bypassKey == "" || (request.Method != http.MethodGet && request.Method != http.MethodHead) {
		return false
	}
	if !strings.HasPrefix(request.URL.Path, "/media/") {
		return false
	}
	provided := request.URL.Query().Get("bp")
	return subtle.ConstantTimeCompare([]byte(provided), []byte(auth.bypassKey)) == 1
}

func (auth *authenticator) setSessionCookie(response http.ResponseWriter, request *http.Request) {
	auth.setSessionCookieFor(response, request, dashboardAuthSurface)
}

func (auth *authenticator) setSessionCookieFor(response http.ResponseWriter, request *http.Request, surface authSurface) {
	expiresAt := time.Now().Add(sessionDuration)
	payload := auth.username + "\n" + strconv.FormatInt(expiresAt.Unix(), 10)
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	signature := auth.sign(encodedPayload)
	http.SetCookie(response, &http.Cookie{
		Name: surface.CookieName, Value: encodedPayload + "." + signature, Path: "/",
		Expires: expiresAt, MaxAge: int(sessionDuration.Seconds()), HttpOnly: true,
		Secure: requestIsHTTPS(request), SameSite: surface.SameSite,
	})
}

func (auth *authenticator) hasValidSession(request *http.Request) bool {
	return auth.hasValidSessionFor(request, sessionCookieName)
}

func (auth *authenticator) hasValidSessionFor(request *http.Request, cookieName string) bool {
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		return false
	}
	encodedPayload, signature, ok := strings.Cut(cookie.Value, ".")
	if !ok || subtle.ConstantTimeCompare([]byte(signature), []byte(auth.sign(encodedPayload))) != 1 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return false
	}
	username, expiration, ok := strings.Cut(string(payload), "\n")
	if !ok || subtle.ConstantTimeCompare([]byte(username), []byte(auth.username)) != 1 {
		return false
	}
	expiresAt, err := strconv.ParseInt(expiration, 10, 64)
	return err == nil && time.Now().Unix() < expiresAt
}

func (auth *authenticator) loginPageData(surface authSurface, nextPath string, invalid bool) loginPageData {
	return loginPageData{
		InvalidCredentials: invalid, Next: nextPath, StylesheetPath: surface.StylesheetPath, FaviconPath: surface.FaviconPath,
		FormAction: surface.LoginAction, PageTitle: surface.PageTitle, Heading: surface.Heading,
		Description: surface.Description, Context: surface.Context,
	}
}

func (auth *authenticator) loginRetryAfter(clientKey string) (time.Duration, bool) {
	now := time.Now()
	auth.attemptsMutex.Lock()
	defer auth.attemptsMutex.Unlock()
	attempt, exists := auth.attempts[clientKey]
	if !exists || now.Sub(attempt.WindowStart) >= loginAttemptWindow {
		if exists {
			delete(auth.attempts, clientKey)
		}
		return 0, false
	}
	if attempt.Count < maxLoginAttempts {
		return 0, false
	}
	return loginAttemptWindow - now.Sub(attempt.WindowStart), true
}

func (auth *authenticator) recordLoginFailure(clientKey string) {
	now := time.Now()
	auth.attemptsMutex.Lock()
	defer auth.attemptsMutex.Unlock()
	attempt := auth.attempts[clientKey]
	if attempt.WindowStart.IsZero() || now.Sub(attempt.WindowStart) >= loginAttemptWindow {
		attempt = loginAttempt{WindowStart: now}
	}
	attempt.Count++
	auth.attempts[clientKey] = attempt
}

func (auth *authenticator) clearLoginFailures(clientKey string) {
	auth.attemptsMutex.Lock()
	delete(auth.attempts, clientKey)
	auth.attemptsMutex.Unlock()
}

func loginClientKey(request *http.Request) string {
	if address, err := clientIPAddress(request); err == nil {
		return address.String()
	}
	return request.RemoteAddr
}

func (auth *authenticator) hasTrustedIPAddress(request *http.Request) bool {
	address, err := clientIPAddress(request)
	return err == nil && auth.trustedIPs.Contains(address)
}

func (auth *authenticator) sign(value string) string {
	mac := hmac.New(sha256.New, auth.sessionKey)
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func safeNextPath(candidate string) string {
	if candidate == "" {
		return "/app/"
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return "/app/"
	}
	return parsed.RequestURI()
}

func requestIsHTTPS(request *http.Request) bool {
	if request.TLS != nil {
		return true
	}
	forwardedProtocol := strings.TrimSpace(strings.Split(request.Header.Get("X-Forwarded-Proto"), ",")[0])
	return strings.EqualFold(forwardedProtocol, "https")
}

func isMediaBypass(request *http.Request) bool {
	active, _ := request.Context().Value(mediaBypassContextKey{}).(bool)
	return active
}
