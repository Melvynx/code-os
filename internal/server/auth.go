package server

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
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
	"time"
)

const (
	sessionCookieName = "stackenv_session"
	sessionDuration   = 12 * time.Hour
)

type mediaBypassContextKey struct{}

type authenticator struct {
	username        string
	password        string
	bypassKey       string
	sessionKey      []byte
	loginTemplate   *template.Template
	loginStylesheet []byte
}

type loginPageData struct {
	InvalidCredentials bool
	Next               string
}

func newAuthenticator(assets fs.FS, username, password, bypassKey string) (*authenticator, error) {
	loginTemplate, err := template.ParseFS(assets, "login.html")
	if err != nil {
		return nil, fmt.Errorf("parse login page: %w", err)
	}
	loginStylesheet, err := fs.ReadFile(assets, "login.css")
	if err != nil {
		return nil, fmt.Errorf("read login stylesheet: %w", err)
	}
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		return nil, fmt.Errorf("generate session key: %w", err)
	}
	return &authenticator{
		username: username, password: password, bypassKey: bypassKey,
		sessionKey: sessionKey, loginTemplate: loginTemplate, loginStylesheet: loginStylesheet,
	}, nil
}

func (auth *authenticator) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if auth.hasValidSession(request) {
			next.ServeHTTP(response, request)
			return
		}
		if auth.hasValidMediaBypass(request) {
			ctx := context.WithValue(request.Context(), mediaBypassContextKey{}, true)
			next.ServeHTTP(response, request.WithContext(ctx))
			return
		}
		if strings.HasPrefix(request.URL.Path, "/api/") || strings.HasPrefix(request.URL.Path, "/media/") {
			writeJSON(response, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			return
		}
		nextPath := request.URL.RequestURI()
		if nextPath == "" {
			nextPath = "/"
		}
		http.Redirect(response, request, "/login?next="+url.QueryEscape(nextPath), http.StatusSeeOther)
	})
}

func (auth *authenticator) loginPage(response http.ResponseWriter, request *http.Request) {
	nextPath := safeNextPath(request.URL.Query().Get("next"))
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Header().Set("Cache-Control", "no-store")
	if err := auth.loginTemplate.ExecuteTemplate(response, "login.html", loginPageData{Next: nextPath}); err != nil {
		http.Error(response, "Unable to render sign in", http.StatusInternalServerError)
	}
}

func (auth *authenticator) loginStyles(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/css; charset=utf-8")
	response.Header().Set("Cache-Control", "no-store")
	_, _ = response.Write(auth.loginStylesheet)
}

func (auth *authenticator) login(response http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(response, request.Body, 8<<10)
	if err := request.ParseForm(); err != nil {
		http.Error(response, "Invalid form", http.StatusBadRequest)
		return
	}
	nextPath := safeNextPath(request.FormValue("next"))
	usernameMatches := subtle.ConstantTimeCompare([]byte(request.FormValue("username")), []byte(auth.username)) == 1
	passwordMatches := subtle.ConstantTimeCompare([]byte(request.FormValue("password")), []byte(auth.password)) == 1
	if !usernameMatches || !passwordMatches {
		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		response.Header().Set("Cache-Control", "no-store")
		response.WriteHeader(http.StatusUnauthorized)
		_ = auth.loginTemplate.ExecuteTemplate(response, "login.html", loginPageData{
			InvalidCredentials: true,
			Next:               nextPath,
		})
		return
	}
	auth.setSessionCookie(response, request)
	http.Redirect(response, request, nextPath, http.StatusSeeOther)
}

func (auth *authenticator) logout(response http.ResponseWriter, request *http.Request) {
	http.SetCookie(response, &http.Cookie{
		Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: requestIsHTTPS(request), SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(response, request, "/login", http.StatusSeeOther)
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
	expiresAt := time.Now().Add(sessionDuration)
	payload := auth.username + "\n" + strconv.FormatInt(expiresAt.Unix(), 10)
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	signature := auth.sign(encodedPayload)
	http.SetCookie(response, &http.Cookie{
		Name: sessionCookieName, Value: encodedPayload + "." + signature, Path: "/",
		Expires: expiresAt, MaxAge: int(sessionDuration.Seconds()), HttpOnly: true,
		Secure: requestIsHTTPS(request), SameSite: http.SameSiteStrictMode,
	})
}

func (auth *authenticator) hasValidSession(request *http.Request) bool {
	cookie, err := request.Cookie(sessionCookieName)
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

func (auth *authenticator) sign(value string) string {
	mac := hmac.New(sha256.New, auth.sessionKey)
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func safeNextPath(candidate string) string {
	if candidate == "" {
		return "/"
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return "/"
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
