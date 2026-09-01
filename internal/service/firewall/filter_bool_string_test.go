package firewall

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/biptec/opnsense-go/pkg/api"
	apifirewall "github.com/biptec/opnsense-go/pkg/firewall"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFilterLogBoolStringConversion(t *testing.T) {
	model := &filterResourceModel{
		Enabled: types.BoolValue(true), Sequence: types.Int64Value(1), NoXMLRPCSync: types.BoolValue(false),
		Filter:    &filterFilterBlock{Quick: types.BoolValue(true), Action: types.StringValue("pass"), AllowOptions: types.BoolValue(false), Direction: types.StringValue("in"), IPProtocol: types.StringValue("inet"), Protocol: types.StringValue("CARP"), Log: types.BoolValue(true), TCPFlags: types.SetNull(types.StringType), TCPFlagsOutOf: types.SetNull(types.StringType), Schedule: types.StringValue("")},
		Interface: &filterInterfaceBlock{Invert: types.BoolValue(false), Interface: types.SetNull(types.StringType)},
	}
	remote, err := convertFilterSchemaToStruct(model)
	if err != nil {
		t.Fatalf("convertFilterSchemaToStruct(): %v", err)
	}
	if remote.Log != api.BoolString("1") {
		t.Fatalf("remote log=%q", remote.Log)
	}

	read := &apifirewall.Filter{Log: api.BoolString("1")}
	state, err := convertFilterStructToSchema(read)
	if err != nil {
		t.Fatalf("convertFilterStructToSchema(): %v", err)
	}
	if state.Filter == nil || !state.Filter.Log.ValueBool() {
		t.Fatalf("unexpected log state: %#v", state.Filter)
	}
}

func TestIsStaleReplyToGatewayError(t *testing.T) {
	stale := errors.New("resource not changed. result: failed. errors: map[rule.replyto:Specify a valid gateway from the list matching the networks ip protocol.]")
	if !isStaleReplyToGatewayError(stale) {
		t.Fatal("expected reply-to gateway validation error to be retryable")
	}
	if isStaleReplyToGatewayError(errors.New("resource not changed. result: failed. errors: map[rule.gateway:invalid]")) {
		t.Fatal("unrelated gateway validation error must not be retryable")
	}
}

func TestAddFilterResolvedRetriesExistingReplyToGateway(t *testing.T) {
	var addCalls, searchCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/api/firewall/filter/addRule":
			addCalls++
			if addCalls == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"result":      "failed",
					"validations": map[string]any{"rule.replyto": "Specify a valid gateway from the list matching the networks ip protocol."},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "saved", "uuid": "filter-id"})
		case "/api/routing/settings/searchGateway":
			searchCalls++
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total": 1, "rowCount": 1, "current": 1,
				"rows": []map[string]any{{
					"name":       "NC_TRANSIT_PEER_V4",
					"ipprotocol": map[string]any{"inet": map[string]any{"selected": 1, "value": "IPv4"}},
				}},
			})
		case "/api/firewall/filter/apply":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		default:
			http.NotFound(w, req)
		}
	}))
	defer server.Close()

	resource := &filterResource{client: opnsense.NewClient(api.NewClient(api.Options{Uri: server.URL}))}
	filter := &apifirewall.Filter{IPProtocol: api.SelectedMap("inet"), ReplyTo: api.SelectedMap("NC_TRANSIT_PEER_V4")}
	id, err := resource.addFilterResolvedWithTiming(context.Background(), filter, time.Millisecond, time.Second)
	if err != nil {
		t.Fatalf("addFilterResolvedWithTiming(): %v", err)
	}
	if id != "filter-id" || addCalls != 2 || searchCalls != 1 {
		t.Fatalf("id=%q add=%d search=%d, want filter-id/2/1", id, addCalls, searchCalls)
	}
}

func TestAddFilterResolvedDoesNotRetryMissingOrWrongFamilyGateway(t *testing.T) {
	for _, protocol := range []string{"inet6", ""} {
		t.Run(protocol, func(t *testing.T) {
			var addCalls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.URL.Path {
				case "/api/firewall/filter/addRule":
					addCalls++
					_ = json.NewEncoder(w).Encode(map[string]any{
						"result":      "failed",
						"validations": map[string]any{"rule.replyto": "Specify a valid gateway from the list matching the networks ip protocol."},
					})
				case "/api/routing/settings/searchGateway":
					rows := []map[string]any{}
					if protocol != "" {
						rows = append(rows, map[string]any{
							"name":       "NC_TRANSIT_PEER_V4",
							"ipprotocol": map[string]any{protocol: map[string]any{"selected": 1, "value": protocol}},
						})
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"total": len(rows), "rowCount": len(rows), "current": 1, "rows": rows})
				case "/api/firewall/filter/apply":
					_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
				default:
					http.NotFound(w, req)
				}
			}))
			defer server.Close()

			resource := &filterResource{client: opnsense.NewClient(api.NewClient(api.Options{Uri: server.URL}))}
			filter := &apifirewall.Filter{IPProtocol: api.SelectedMap("inet"), ReplyTo: api.SelectedMap("NC_TRANSIT_PEER_V4")}
			_, err := resource.addFilterResolvedWithTiming(context.Background(), filter, 0, time.Second)
			if err == nil {
				t.Fatal("expected original validation error")
			}
			if addCalls != 1 {
				t.Fatalf("add calls = %d, want 1", addCalls)
			}
		})
	}
}
