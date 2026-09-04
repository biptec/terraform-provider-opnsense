package quagga

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
	"github.com/biptec/opnsense-go/pkg/opnsense"
)

func TestUnlinkOSPFPrefixListFromRouteMaps(t *testing.T) {
	testUnlinkOSPFPrefixListFromRouteMaps(t, false)
}

func TestUnlinkOSPF6PrefixListFromRouteMaps(t *testing.T) {
	testUnlinkOSPFPrefixListFromRouteMaps(t, true)
}

func testUnlinkOSPFPrefixListFromRouteMaps(t *testing.T, ipv6 bool) {
	t.Helper()

	family := "ospfsettings"
	if ipv6 {
		family = "ospf6settings"
	}

	const (
		removeID = "prefix-remove"
		keepID   = "prefix-keep"
		routeID  = "route-map-1"
	)

	var updated []string
	searchPath := "/api/quagga/" + family + "/searchRoutemap"
	getPath := "/api/quagga/" + family + "/getRoutemap/" + routeID
	setPath := "/api/quagga/" + family + "/setRoutemap/" + routeID

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case searchPath:
			if r.Method != http.MethodPost {
				t.Fatalf("search method = %s, want POST", r.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"rows": []map[string]any{
					{"uuid": routeID, "match2": removeID + "," + keepID},
					{"uuid": "unrelated", "match2": keepID},
				},
				"rowCount": 2,
				"total":    2,
				"current":  1,
			})
		case getPath:
			if r.Method != http.MethodGet {
				t.Fatalf("get method = %s, want GET", r.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"routemap": map[string]any{
					"enabled": "1",
					"name":    "nc-connected",
					"action":  "permit",
					"id":      "10",
					"match2":  removeID + "," + keepID,
					"set":     "",
				},
			})
		case setPath:
			if r.Method != http.MethodPost {
				t.Fatalf("set method = %s, want POST", r.Method)
			}
			var body map[string]map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode route-map update: %v", err)
			}
			got, _ := body["routemap"]["match2"].(string)
			updated = strings.FieldsFunc(got, func(r rune) bool { return r == ',' || r == '\n' })
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "saved"})
		case "/api/quagga/service/reconfigure":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := opnsense.NewClient(api.NewClient(api.Options{Uri: server.URL, MaxRetries: -1}))
	var err error
	if ipv6 {
		err = unlinkOSPF6PrefixListFromRouteMaps(context.Background(), client, removeID)
	} else {
		err = unlinkOSPFPrefixListFromRouteMaps(context.Background(), client, removeID)
	}
	if err != nil {
		t.Fatalf("unlink prefix list: %v", err)
	}
	if len(updated) != 1 || updated[0] != keepID {
		t.Fatalf("updated prefix list references = %#v, want [%q]", updated, keepID)
	}
}

func TestUnlinkOSPFPrefixListFromRouteMapsSerializesConcurrentMutations(t *testing.T) {
	for _, tc := range []struct {
		name string
		ipv6 bool
	}{
		{name: "ospfv2", ipv6: false},
		{name: "ospfv3", ipv6: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			family := "ospfsettings"
			if tc.ipv6 {
				family = "ospf6settings"
			}

			searchPath := "/api/quagga/" + family + "/searchRoutemap"
			firstSearchEntered := make(chan struct{})
			releaseFirstSearch := make(chan struct{})
			secondSearchEntered := make(chan struct{})
			var searches atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path != searchPath {
					http.NotFound(w, r)
					return
				}

				switch searches.Add(1) {
				case 1:
					close(firstSearchEntered)
					<-releaseFirstSearch
				case 2:
					close(secondSearchEntered)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"rows":     []map[string]any{},
					"rowCount": 0,
					"total":    0,
					"current":  1,
				})
			}))
			defer server.Close()

			client := opnsense.NewClient(api.NewClient(api.Options{Uri: server.URL, MaxRetries: -1}))
			unlink := func(id string) error {
				if tc.ipv6 {
					return unlinkOSPF6PrefixListFromRouteMaps(context.Background(), client, id)
				}
				return unlinkOSPFPrefixListFromRouteMaps(context.Background(), client, id)
			}

			errCh := make(chan error, 2)
			go func() { errCh <- unlink("prefix-a") }()
			<-firstSearchEntered
			go func() { errCh <- unlink("prefix-b") }()

			select {
			case <-secondSearchEntered:
				close(releaseFirstSearch)
				t.Fatal("second prefix-list unlink reached the shared route-map before the first unlink completed")
			case <-time.After(100 * time.Millisecond):
				// Expected: the family-specific mutex keeps the second mutation out.
			}

			close(releaseFirstSearch)
			for range 2 {
				if err := <-errCh; err != nil {
					t.Fatalf("unlink prefix list: %v", err)
				}
			}

			select {
			case <-secondSearchEntered:
			case <-time.After(time.Second):
				t.Fatal("second prefix-list unlink did not proceed after the first unlink completed")
			}
		})
	}
}

func TestWithoutSelected(t *testing.T) {
	got, changed := withoutSelected(api.SelectedMapList{"a", "b", "c"}, "b")
	if !changed {
		t.Fatal("withoutSelected() changed = false, want true")
	}
	if strings.Join(got, ",") != "a,c" {
		t.Fatalf("withoutSelected() = %v, want [a c]", got)
	}

	got, changed = withoutSelected(api.SelectedMapList{"a", "c"}, "b")
	if changed {
		t.Fatal("withoutSelected() changed = true for missing id")
	}
	if strings.Join(got, ",") != "a,c" {
		t.Fatalf("withoutSelected() missing id = %v, want [a c]", got)
	}
}
