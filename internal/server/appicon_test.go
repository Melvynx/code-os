package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/melvynx/code-os/internal/model"
)

func TestIconCandidatesPreferLocalThenDomain(t *testing.T) {
	t.Parallel()
	candidates := iconCandidates(model.Application{Port: 3013, Domain: "lumail.io"}, "http://127.0.0.1:3013")
	if len(candidates) < 6 || candidates[0] != "http://127.0.0.1:3013/favicon.svg" {
		t.Fatalf("candidates = %#v", candidates)
	}
	joined := strings.Join(candidates, "\n")
	if !strings.Contains(joined, "https://lumail.io/favicon.ico") || !strings.Contains(joined, "icons.duckduckgo.com/ip3/lumail.io.ico") {
		t.Fatalf("missing domain fallbacks: %#v", candidates)
	}
}

func TestIconTargetRejectsOffPortLoopback(t *testing.T) {
	t.Parallel()
	app := model.Application{Port: 3013, Domain: "lumail.io"}
	if iconTargetAllowed(app, "http://127.0.0.1:80/favicon.ico", "http://127.0.0.1:3013") {
		t.Fatal("allowed off-port loopback")
	}
	if iconTargetAllowed(app, "http://example.com/favicon.ico", "http://127.0.0.1:3013") {
		t.Fatal("allowed arbitrary host")
	}
	if !iconTargetAllowed(app, "https://lumail.io/favicon.ico", "http://127.0.0.1:3013") {
		t.Fatal("rejected application domain")
	}
	if !iconTargetAllowed(app, "http://127.0.0.1:3013/favicon.ico", "http://127.0.0.1:3013") {
		t.Fatal("rejected application origin")
	}
}

func TestParseHTMLIconHrefsRanksSVG(t *testing.T) {
	t.Parallel()
	html := []byte(`<html><head>
<link rel="icon" href="/favicon.ico">
<link rel="icon" type="image/svg+xml" href="/favicon.svg">
<link rel="apple-touch-icon" href="/apple-icon.png" sizes="180x180">
</head></html>`)
	hrefs := parseHTMLIconHrefs(html, "http://127.0.0.1:3013/")
	if len(hrefs) == 0 || hrefs[0] != "http://127.0.0.1:3013/favicon.svg" {
		t.Fatalf("hrefs = %#v", hrefs)
	}
}

func TestApplicationIconServesLocalFavicon(t *testing.T) {
	t.Parallel()
	upstream := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/favicon.ico" {
			http.NotFound(response, request)
			return
		}
		response.Header().Set("Content-Type", "image/x-icon")
		_, _ = io.WriteString(response, "\x00\x00\x01\x00icon")
	}))
	t.Cleanup(upstream.Close)

	server := HTTPServer{
		Service: &Service{snapshot: model.Snapshot{Apps: []model.Application{{
			ID: "srv_web", Name: "start", ProjectName: "lumail.io", State: "running", Port: 3013, Domain: "lumail.io",
		}}}},
		IconClient: upstream.Client(),
		IconOrigin: func(model.Application) string { return upstream.URL },
		Icons:      newIconCache(),
	}

	request := httptest.NewRequest(http.MethodGet, "/api/applications/srv_web/icon", nil)
	request.SetPathValue("id", "srv_web")
	response := httptest.NewRecorder()
	server.applicationIcon(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Type") != "image/x-icon" {
		t.Fatalf("content-type = %q", response.Header().Get("Content-Type"))
	}
	if !strings.Contains(response.Body.String(), "icon") {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestApplicationIconNotFoundForUnknownID(t *testing.T) {
	t.Parallel()
	server := HTTPServer{Service: &Service{}, Icons: newIconCache()}
	request := httptest.NewRequest(http.MethodGet, "/api/applications/missing/icon", nil)
	request.SetPathValue("id", "missing")
	response := httptest.NewRecorder()
	server.applicationIcon(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d", response.Code)
	}
}
