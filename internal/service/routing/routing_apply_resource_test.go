package routing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
)

func TestApplyRoutingConfigReconfiguresRoutes(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/api/routing/settings/reconfigure" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		calls++
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := api.NewClient(api.Options{Uri: server.URL, MaxRetries: -1})
	if err := applyRoutingConfig(context.Background(), client); err != nil {
		t.Fatalf("applyRoutingConfig() error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("reconfigure calls = %d, want 1", calls)
	}
}

func TestApplyRoutingConfigReportsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/api/routing/settings/reconfigure" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "failed"})
	}))
	defer server.Close()

	client := api.NewClient(api.Options{Uri: server.URL, MaxRetries: -1})
	err := applyRoutingConfig(context.Background(), client)
	if err == nil || !strings.Contains(err.Error(), "reconfigure routing") {
		t.Fatalf("applyRoutingConfig() error = %v, want routing reconfigure failure", err)
	}
}
