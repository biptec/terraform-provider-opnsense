package interfaces

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/biptec/opnsense-go/pkg/api"
)

func TestWaitForAssignableDeviceWaitsUntilDeviceAppears(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/interfaces/assignment/get_item/" {
			t.Fatalf("request path = %q", req.URL.Path)
		}
		count := requests.Add(1)
		devices := map[string]any{
			"vtnet0": map[string]any{"selected": 0, "value": "vtnet0"},
		}
		if count >= 2 {
			devices["lo9"] = map[string]any{"selected": 0, "value": "lo9"}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"interface": map[string]any{"if": devices}})
	}))
	defer server.Close()

	client := api.NewClient(api.Options{Uri: server.URL})
	if err := waitForAssignableDevice(context.Background(), client, "lo9", time.Second, time.Millisecond); err != nil {
		t.Fatalf("waitForAssignableDevice() error = %v", err)
	}
	if got := requests.Load(); got < 2 {
		t.Fatalf("readiness requests = %d, want at least 2", got)
	}
}

func TestWaitForAssignableDeviceTimesOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"interface":{"if":{"vtnet0":{"selected":0,"value":"vtnet0"}}}}`))
	}))
	defer server.Close()

	client := api.NewClient(api.Options{Uri: server.URL})
	err := waitForAssignableDevice(context.Background(), client, "lo9", 25*time.Millisecond, 5*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), `interface device "lo9" did not become assignable`) {
		t.Fatalf("waitForAssignableDevice() error = %v", err)
	}
}
