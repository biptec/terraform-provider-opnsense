package bind_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	testViewInternal = "11111111-1111-4111-8111-111111111111"
	testViewPublic   = "22222222-2222-4222-8222-222222222222"
	testZoneInternal = "33333333-3333-4333-8333-333333333333"
	testZonePublic   = "44444444-4444-4444-8444-444444444444"
)

func semanticSelected(value string) map[string]any {
	return map[string]any{value: map[string]any{"value": value, "selected": 1}}
}

func semanticViewPayload(name string) map[string]any {
	return map[string]any{
		"enabled": "1", "sequence": "20", "name": name, "matchany": "0", "matchclients": map[string]any{},
		"matchdestinations": map[string]any{}, "recursion": "0", "allowrecursion": map[string]any{},
		"allowqueryany": "0", "allowquery": map[string]any{}, "allowtransfer": map[string]any{},
		"forwarders": map[string]any{}, "dnssecvalidation": semanticSelected("auto"),
	}
}
func semanticDomainPayload(viewID string) map[string]any {
	return map[string]any{
		"view": semanticSelected(viewID), "domainname": "biptec.com", "enabled": "1",
		"allowtransfer": map[string]any{}, "allowrndctransfer": "0", "primarytransferkey": map[string]any{},
		"alsonotify": map[string]any{}, "allowquery": map[string]any{}, "allowrndcupdate": "0",
		"updatekeys": map[string]any{}, "updatepolicy": semanticSelected("self_txt"), "dnssec": "0",
		"serial": "2026082001", "ttl": "300", "refresh": "300", "retry": "300", "expire": "86400",
		"negative": "300", "mailadmin": "hostmaster@biptec.com", "dnsserver": "ns.biptec.com",
	}
}

func newSemanticLookupServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/bind/view/search_view":
			_ = json.NewEncoder(w).Encode(map[string]any{"total": 2, "rowCount": 2, "current": 1, "rows": []map[string]any{
				{"uuid": testViewInternal, "name": "internal", "sequence": "20"},
				{"uuid": testViewPublic, "name": "public", "sequence": "30"},
			}})
		case "/api/bind/view/get_view/" + testViewInternal:
			_ = json.NewEncoder(w).Encode(map[string]any{"view": semanticViewPayload("internal")})
		case "/api/bind/view/get_view/" + testViewPublic:
			_ = json.NewEncoder(w).Encode(map[string]any{"view": semanticViewPayload("public")})
		case "/api/bind/domain/search_primary_domain":
			_ = json.NewEncoder(w).Encode(map[string]any{"total": 2, "rowCount": 2, "current": 1, "rows": []map[string]any{
				{"uuid": testZoneInternal, "view": testViewInternal, "domainname": "biptec.com"},
				{"uuid": testZonePublic, "view": testViewPublic, "domainname": "BIPTEC.COM."},
			}})
		case "/api/bind/domain/get_domain/" + testZoneInternal:
			_ = json.NewEncoder(w).Encode(map[string]any{"domain": semanticDomainPayload(testViewInternal)})
		case "/api/bind/domain/get_domain/" + testZonePublic:
			_ = json.NewEncoder(w).Encode(map[string]any{"domain": semanticDomainPayload(testViewPublic)})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestBindSemanticLookupFrameworkContract(t *testing.T) {
	server := newSemanticLookupServer(t)
	defer server.Close()

	config := fmt.Sprintf(`
provider "opnsense" {
  uri        = %q
  api_key    = "test"
  api_secret = "test"
  retries    = 1
}
`, server.URL)
	config += `
data "opnsense_bind_view" "internal" {
  name = " INTERNAL "
}

data "opnsense_bind_primary_domain" "semantic" {
  domain_name = " BIPTEC.COM. "
  view_name   = "INTERNAL"
}

data "opnsense_bind_primary_domain" "by_view_id" {
  domain_name = "biptec.com"
  view_id     = "11111111-1111-4111-8111-111111111111"
}

data "opnsense_bind_primary_domain" "by_id" {
  id = "33333333-3333-4333-8333-333333333333"
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{Config: config, Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.opnsense_bind_view.internal", "id", testViewInternal),
			resource.TestCheckResourceAttr("data.opnsense_bind_view.internal", "name", "internal"),
			resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.semantic", "id", testZoneInternal),
			resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.semantic", "view_id", testViewInternal),
			resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.semantic", "view_name", "internal"),
			resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.semantic", "domain_name", "biptec.com"),
			resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.by_view_id", "id", testZoneInternal),
			resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.by_view_id", "view_name", "internal"),
			resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.by_id", "domain_name", "biptec.com"),
			resource.TestCheckResourceAttr("data.opnsense_bind_primary_domain.by_id", "view_name", "internal"),
		)}},
	})
}

type semanticStoredRecord struct {
	Domain  string `json:"domain"`
	Enabled string `json:"enabled"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Value   string `json:"value"`
}

type semanticRecordStore struct {
	mu      sync.Mutex
	records map[string]semanticStoredRecord
	adds    map[string]int
	deletes map[string]int
}

func semanticRecordID(name string) string {
	switch name {
	case "alias1":
		return "55555555-5555-4555-8555-555555555555"
	case "alias2":
		return "66666666-6666-4666-8666-666666666666"
	case "alias3":
		return "77777777-7777-4777-8777-777777777777"
	default:
		return "88888888-8888-4888-8888-888888888888"
	}
}
func newSemanticRecordServer(t *testing.T) (*httptest.Server, *semanticRecordStore) {
	t.Helper()
	store := &semanticRecordStore{records: map[string]semanticStoredRecord{}, adds: map[string]int{}, deletes: map[string]int{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/bind/view/search_view":
			_ = json.NewEncoder(w).Encode(map[string]any{"total": 1, "rowCount": 1, "current": 1, "rows": []map[string]any{
				{"uuid": testViewInternal, "name": "internal", "sequence": "20"},
			}})
		case "/api/bind/view/get_view/" + testViewInternal:
			_ = json.NewEncoder(w).Encode(map[string]any{"view": semanticViewPayload("internal")})
		case "/api/bind/domain/search_primary_domain":
			_ = json.NewEncoder(w).Encode(map[string]any{"total": 1, "rowCount": 1, "current": 1, "rows": []map[string]any{
				{"uuid": testZoneInternal, "view": testViewInternal, "domainname": "example.test"},
			}})
		case "/api/bind/domain/get_domain/" + testZoneInternal:
			payload := semanticDomainPayload(testViewInternal)
			payload["domainname"] = "example.test"
			_ = json.NewEncoder(w).Encode(map[string]any{"domain": payload})
		case "/api/bind/record/add_record":
			var body struct {
				Record semanticStoredRecord `json:"record"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode add record: %v", err)
			}
			id := semanticRecordID(body.Record.Name)
			store.mu.Lock()
			store.records[id] = body.Record
			store.adds[body.Record.Name]++
			store.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "saved", "uuid": id})
		case "/api/bind/service/reload":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		default:
			switch {
			case strings.HasPrefix(r.URL.Path, "/api/bind/record/get_record/"):
				id := strings.TrimPrefix(r.URL.Path, "/api/bind/record/get_record/")
				store.mu.Lock()
				record, ok := store.records[id]
				store.mu.Unlock()
				if !ok {
					_ = json.NewEncoder(w).Encode(map[string]any{})
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"record": record})
			case strings.HasPrefix(r.URL.Path, "/api/bind/record/del_record/"):
				id := strings.TrimPrefix(r.URL.Path, "/api/bind/record/del_record/")
				store.mu.Lock()
				record, ok := store.records[id]
				if ok {
					store.deletes[record.Name]++
				}
				delete(store.records, id)
				store.mu.Unlock()
				if !ok {
					_ = json.NewEncoder(w).Encode(map[string]any{"result": "not found"})
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "deleted"})
			default:
				http.NotFound(w, r)
			}
		}
	}))
	return server, store
}

func semanticRecordsConfig(serverURL string, includeAlias2 bool) string {
	alias2 := ""
	if includeAlias2 {
		alias2 = `
    alias2 = {
      type  = "CNAME"
      value = "service2.example."
    }`
	}
	return fmt.Sprintf(`
terraform {
  required_providers {
    opnsense = {
      source = "biptec/opnsense"
    }
  }
}

provider "opnsense" {
  uri        = %q
  api_key    = "test"
  api_secret = "test"
  retries    = 1
}

data "opnsense_bind_primary_domain" "zone" {
  domain_name = "EXAMPLE.TEST."
  view_name   = "INTERNAL"
}

locals {
  aliases = {
    alias1 = {
      type  = "CNAME"
      value = "service1.example."
    }%s
    alias3 = {
      type  = "CNAME"
      value = "service3.example."
    }
  }
}

resource "opnsense_bind_record" "alias" {
  for_each = local.aliases

  domain_id = data.opnsense_bind_primary_domain.zone.id
  name      = each.key
  type      = each.value.type
  value     = each.value.value
}
`, serverURL, alias2)
}

func (s *semanticRecordStore) stats(name string) (present bool, adds, deletes int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, present = s.records[semanticRecordID(name)]
	return present, s.adds[name], s.deletes[name]
}

func semanticTerraformCLI(t *testing.T) string {
	t.Helper()
	for _, name := range []string{"tofu", "terraform"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	t.Skip("OpenTofu or Terraform CLI is required for for_each lifecycle coverage")
	return ""
}

func runSemanticTerraform(t *testing.T, ctx context.Context, cli, dir, cliConfig string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.CommandContext(ctx, cli, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "TF_CLI_CONFIG_FILE="+cliConfig, "TF_IN_AUTOMATION=1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return string(out), exitErr.ExitCode()
	}
	t.Fatalf("%s %v failed to start: %v", filepath.Base(cli), args, err)
	return "", -1
}

func semanticProviderDevOverride(t *testing.T, workDir string) string {
	t.Helper()
	gomod, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("go env GOMOD: %v", err)
	}
	repoRoot := filepath.Dir(strings.TrimSpace(string(gomod)))
	devDir := filepath.Join(workDir, "provider-dev")
	if err := os.MkdirAll(devDir, 0o755); err != nil {
		t.Fatalf("create provider dev dir: %v", err)
	}
	build := exec.Command("go", "build", "-o", filepath.Join(devDir, "terraform-provider-opnsense"), ".")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build provider: %v\n%s", err, out)
	}
	cliConfig := filepath.Join(workDir, "tofurc")
	contents := fmt.Sprintf("provider_installation {\n  dev_overrides {\n    \"registry.opentofu.org/biptec/opnsense\" = %q\n    \"registry.terraform.io/biptec/opnsense\" = %q\n  }\n  direct {}\n}\n", devDir, devDir)
	if err := os.WriteFile(cliConfig, []byte(contents), 0o600); err != nil {
		t.Fatalf("write tofu CLI config: %v", err)
	}
	return cliConfig
}

func TestBindRecordForEachWithSemanticZone(t *testing.T) {
	cli := semanticTerraformCLI(t)
	server, store := newSemanticRecordServer(t)
	defer server.Close()

	workDir := t.TempDir()
	cliConfig := semanticProviderDevOverride(t, workDir)
	configPath := filepath.Join(workDir, "main.tf")
	writeConfig := func(config string) {
		t.Helper()
		if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
			t.Fatalf("write OpenTofu config: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	writeConfig(semanticRecordsConfig(server.URL, true))
	if out, code := runSemanticTerraform(t, ctx, cli, workDir, cliConfig, "init", "-backend=false", "-input=false"); code != 0 {
		t.Fatalf("init exit=%d\n%s", code, out)
	}
	if out, code := runSemanticTerraform(t, ctx, cli, workDir, cliConfig, "apply", "-auto-approve", "-input=false"); code != 0 {
		t.Fatalf("initial apply exit=%d\n%s", code, out)
	}
	for _, name := range []string{"alias1", "alias2", "alias3"} {
		present, adds, deletes := store.stats(name)
		if !present || adds != 1 || deletes != 0 {
			t.Fatalf("after initial apply %s: present=%t adds=%d deletes=%d", name, present, adds, deletes)
		}
	}

	writeConfig(semanticRecordsConfig(server.URL, false))
	if out, code := runSemanticTerraform(t, ctx, cli, workDir, cliConfig, "apply", "-auto-approve", "-input=false"); code != 0 {
		t.Fatalf("second apply exit=%d\n%s", code, out)
	}
	for _, name := range []string{"alias1", "alias3"} {
		present, adds, deletes := store.stats(name)
		if !present || adds != 1 || deletes != 0 {
			t.Fatalf("unchanged %s was touched: present=%t adds=%d deletes=%d", name, present, adds, deletes)
		}
	}
	if present, adds, deletes := store.stats("alias2"); present || adds != 1 || deletes != 1 {
		t.Fatalf("alias2 lifecycle: present=%t adds=%d deletes=%d", present, adds, deletes)
	}

	if out, code := runSemanticTerraform(t, ctx, cli, workDir, cliConfig, "plan", "-detailed-exitcode", "-input=false"); code != 0 {
		t.Fatalf("final plan is not empty (exit=%d)\n%s", code, out)
	}
}
