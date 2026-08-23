package haproxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
)

func TestApplyHAProxyConfigReconfiguresValidStaging(t *testing.T) {
	t.Parallel()
	var configtests, reconfigures int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/haproxy/service/configtest":
			configtests++
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "Configuration file is valid"})
		case "/api/haproxy/service/reconfigure":
			reconfigures++
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	r := &resourceClient{client: opnsense.NewClient(api.NewClient(api.Options{Uri: server.URL}))}
	if err := r.applyHAProxyConfig(context.Background()); err != nil {
		t.Fatalf("applyHAProxyConfig() error = %v", err)
	}
	if configtests != 1 || reconfigures != 1 {
		t.Fatalf("configtest/reconfigure calls = %d/%d, want 1/1", configtests, reconfigures)
	}
}

func TestApplyHAProxyConfigRejectsInvalidStaging(t *testing.T) {
	t.Parallel()
	reconfigures := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/haproxy/service/configtest":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "[ALERT] invalid configuration"})
		case "/api/haproxy/service/reconfigure":
			reconfigures++
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	r := &resourceClient{client: opnsense.NewClient(api.NewClient(api.Options{Uri: server.URL}))}
	err := r.applyHAProxyConfig(context.Background())
	if err == nil || !strings.Contains(err.Error(), "configuration test result") {
		t.Fatalf("applyHAProxyConfig() error = %v, want configuration test failure", err)
	}
	if reconfigures != 0 {
		t.Fatalf("reconfigure called %d times for invalid staging", reconfigures)
	}
}

func TestApplyHAProxyConfigReportsReconfigureFailure(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/haproxy/service/configtest":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "Configuration file is valid"})
		case "/api/haproxy/service/reconfigure":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "failed"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	r := &resourceClient{client: opnsense.NewClient(api.NewClient(api.Options{Uri: server.URL}))}
	err := r.applyHAProxyConfig(context.Background())
	if err == nil || !strings.Contains(err.Error(), "reconfigure HAProxy status") {
		t.Fatalf("applyHAProxyConfig() error = %v, want reconfigure failure", err)
	}
}
