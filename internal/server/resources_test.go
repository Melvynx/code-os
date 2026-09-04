package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/melvynx/code-os/internal/model"
)

func TestResourcesAPIReturnsRecordedSeries(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 9, 1, 1, 20, 0, 0, time.UTC)
	server := HTTPServer{Service: &Service{resources: []model.ResourceSample{{
		At: at,
		Rows: []model.ResourceRow{{
			ID: "cursor", Kind: "agent", Name: "Cursor agent", CPUPercent: 0.4, MemoryBytes: 209_000_000,
		}},
	}}}}
	response := httptest.NewRecorder()
	server.resources(response, httptest.NewRequest(http.MethodGet, "/api/resources", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	var history model.ResourceHistory
	if err := json.Unmarshal(response.Body.Bytes(), &history); err != nil {
		t.Fatal(err)
	}
	if history.SampleCount != 1 || history.RetentionHours != 6 || len(history.Series) != 1 || history.Series[0].Name != "Cursor agent" {
		t.Fatalf("history = %+v", history)
	}
}
