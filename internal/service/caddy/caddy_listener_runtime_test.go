package caddy_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCaddyMultipleListenerAddresses(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { caddyPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccCaddySettingsConfig(80, 443, "root"),
				ResourceName:       "opnsense_caddy_settings.test",
				ImportState:        true,
				ImportStateId:      "caddy_settings",
				ImportStateVerify:  false,
				ImportStatePersist: true,
			},
			{
				Config: testAccCaddyMultipleListenerConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "http_port", "18080"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "listen_addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr("opnsense_caddy_settings.test", "listen_addresses.*", "192.0.2.2"),
					resource.TestCheckTypeSetElemAttr("opnsense_caddy_settings.test", "listen_addresses.*", "198.51.100.2"),
					checkCaddyTCPListeners(18080, []string{"192.0.2.2", "198.51.100.2"}),
				),
			},
			{
				Config: testAccCaddyMultipleListenerConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_caddy_settings.test", "listen_addresses.#", "1"),
					resource.TestCheckTypeSetElemAttr("opnsense_caddy_settings.test", "listen_addresses.*", "127.0.0.1"),
					checkCaddyTCPListeners(18080, nil),
				),
			},
		},
	})
}

func testAccCaddyMultipleListenerConfig(enabled bool) string {
	addresses := `["127.0.0.1"]`
	if enabled {
		addresses = `["192.0.2.2", "198.51.100.2"]`
	}
	return fmt.Sprintf(`
resource "opnsense_interfaces_loopback" "frontend_a" {
  description = "Caddy frontend A"
}

resource "opnsense_interfaces_assignment" "frontend_a" {
  device          = format("lo%%d", opnsense_interfaces_loopback.frontend_a.device_id)
  description     = "Caddy frontend A"
  enabled         = true
  allow_readdress = true

  ipv4 = {
    mode    = "static"
    address = "192.0.2.2"
    prefix  = 30
  }

  ipv6 = {
    mode = "none"
  }
}

resource "opnsense_interfaces_loopback" "frontend_b" {
  description = "Caddy frontend B"
}

resource "opnsense_interfaces_assignment" "frontend_b" {
  device          = format("lo%%d", opnsense_interfaces_loopback.frontend_b.device_id)
  description     = "Caddy frontend B"
  enabled         = true
  allow_readdress = true

  ipv4 = {
    mode    = "static"
    address = "198.51.100.2"
    prefix  = 30
  }

  ipv6 = {
    mode = "none"
  }
}

resource "opnsense_caddy_settings" "test" {
  enabled               = %[1]t
  enable_layer4         = false
  http_port             = 18080
  https_port            = 18443
  listen_addresses      = %[2]s
  acme_email            = ""
  auto_https            = "off"
  run_as_user            = "root"
  grace_period           = 10
  http_versions          = ["h1", "h2"]
  log_level              = ""
  plain_access_log       = false
  plain_access_log_keep  = 10

  depends_on = [
    opnsense_interfaces_assignment.frontend_a,
    opnsense_interfaces_assignment.frontend_b,
  ]
}
`, enabled, addresses)
}

func checkCaddyTCPListeners(port int, expected []string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		socketPath := caddyQGASocket()
		if socketPath == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		expectedSet := make(map[string]struct{}, len(expected))
		for _, address := range expected {
			expectedSet[address] = struct{}{}
		}
		pattern := regexp.MustCompile(`(?:^|\s)(\*|[0-9.]+):` + fmt.Sprint(port) + `(?:\s|$)`)
		var lastOutput string
		for ctx.Err() == nil {
			output, err := caddyQGAGuestExec(ctx, socketPath, "/usr/bin/sockstat", "-4", "-l", "-P", "tcp")
			if err == nil {
				lastOutput = output
				actualSet := make(map[string]struct{})
				for _, line := range strings.Split(output, "\n") {
					match := pattern.FindStringSubmatch(line)
					if len(match) == 2 {
						actualSet[match[1]] = struct{}{}
					}
				}
				if sameStringSet(actualSet, expectedSet) {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("Caddy TCP listeners on port %d are %v, expected %v; sockstat: %s", port, sortedKeysFromSet(lastOutput, pattern), expected, strings.TrimSpace(lastOutput))
	}
}

func sameStringSet(actual, expected map[string]struct{}) bool {
	if len(actual) != len(expected) {
		return false
	}
	for value := range expected {
		if _, ok := actual[value]; !ok {
			return false
		}
	}
	return true
}

func sortedKeysFromSet(output string, pattern *regexp.Regexp) []string {
	values := make(map[string]struct{})
	for _, line := range strings.Split(output, "\n") {
		match := pattern.FindStringSubmatch(line)
		if len(match) == 2 {
			values[match[1]] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func caddyQGASocket() string {
	path := os.Getenv("QEMU_GA_SOCKET")
	if path == "" {
		path = "/tmp/qemu-virtserialport.sock"
	}
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

type caddyQGAError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

type caddyQGAEnvelope struct {
	Return json.RawMessage `json:"return"`
	Error  *caddyQGAError  `json:"error"`
	Event  string          `json:"event"`
}

func caddyQGACall(ctx context.Context, socketPath string, request any, result any) error {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	connection, err := dialer.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return err
	}
	defer connection.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = connection.SetDeadline(deadline)
	}
	if err = json.NewEncoder(connection).Encode(request); err != nil {
		return err
	}
	decoder := json.NewDecoder(connection)
	for {
		var envelope caddyQGAEnvelope
		if err = decoder.Decode(&envelope); err != nil {
			return err
		}
		if envelope.Event != "" {
			continue
		}
		if envelope.Error != nil {
			return fmt.Errorf("QEMU guest agent %s: %s", envelope.Error.Class, envelope.Error.Desc)
		}
		if envelope.Return == nil {
			continue
		}
		if result == nil {
			return nil
		}
		return json.Unmarshal(envelope.Return, result)
	}
}

func caddyQGAGuestExec(ctx context.Context, socketPath, executable string, arguments ...string) (string, error) {
	var started struct {
		PID int `json:"pid"`
	}
	request := map[string]any{
		"execute": "guest-exec",
		"arguments": map[string]any{
			"path":           executable,
			"arg":            arguments,
			"capture-output": true,
		},
	}
	if err := caddyQGACall(ctx, socketPath, request, &started); err != nil {
		return "", err
	}
	for {
		var status struct {
			Exited   bool   `json:"exited"`
			ExitCode int    `json:"exitcode"`
			OutData  string `json:"out-data"`
			ErrData  string `json:"err-data"`
		}
		statusRequest := map[string]any{
			"execute":   "guest-exec-status",
			"arguments": map[string]any{"pid": started.PID},
		}
		if err := caddyQGACall(ctx, socketPath, statusRequest, &status); err != nil {
			return "", err
		}
		if !status.Exited {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(250 * time.Millisecond):
				continue
			}
		}
		stdout, err := base64.StdEncoding.DecodeString(status.OutData)
		if err != nil {
			return "", fmt.Errorf("decode guest stdout: %w", err)
		}
		stderr, err := base64.StdEncoding.DecodeString(status.ErrData)
		if err != nil {
			return "", fmt.Errorf("decode guest stderr: %w", err)
		}
		if status.ExitCode != 0 {
			return string(stdout), fmt.Errorf("guest command exited %d: %s", status.ExitCode, string(stderr))
		}
		return string(stdout), nil
	}
}
