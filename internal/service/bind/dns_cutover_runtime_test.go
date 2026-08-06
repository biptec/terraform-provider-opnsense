package bind_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDNSServiceCutover(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { bindPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccDNSCutoverConfig("legacy"),
				ResourceName:       "opnsense_unbound_service.test",
				ImportState:        true,
				ImportStateId:      "unbound_service",
				ImportStatePersist: true,
			},
			{
				Config:             testAccDNSCutoverConfig("legacy"),
				ResourceName:       "opnsense_dnsmasq_settings.test",
				ImportState:        true,
				ImportStateId:      "dnsmasq_settings",
				ImportStatePersist: true,
			},
			{
				Config:             testAccDNSCutoverConfig("legacy"),
				ResourceName:       "opnsense_bind_settings.test",
				ImportState:        true,
				ImportStateId:      "bind_settings",
				ImportStatePersist: true,
			},
			{
				Config: testAccDNSCutoverConfig("prepared"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_service.test", "enabled", "false"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.test", "dns_port", "0"),
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "enabled", "false"),
					checkDNSPort53("", nil, nil),
					checkOPNsenseAPIStable(),
				),
			},
			{
				Config: testAccDNSCutoverConfig("active"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "port", "53"),
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "listen_ipv4.#", "2"),
					checkDNSPort53("named", []string{"192.0.2.2", "198.51.100.231"}, nil),
				),
			},
			{
				Config: testAccDNSCutoverConfig("prepared"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "enabled", "false"),
					checkDNSPort53("", nil, nil),
				),
			},
			{
				Config: testAccDNSCutoverConfig("legacy"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_service.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_dnsmasq_settings.test", "dns_port", "53053"),
					resource.TestCheckResourceAttr("opnsense_bind_settings.test", "enabled", "false"),
					checkDNSPort53("unbound", []string{"*"}, []string{"*"}),
				),
			},
			{
				Config: testAccDNSCutoverCleanupConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDNSPort53("unbound", []string{"*"}, []string{"*"}),
					checkOPNsenseAPIStable(),
				),
			},
			{
				Config: testAccDNSCutoverCleanupConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDNSPort53("unbound", []string{"*"}, []string{"*"}),
					checkOPNsenseAPIStable(),
				),
			},
		},
	})
}
func testAccDNSCutoverConfig(phase string) string {
	unboundEnabled := false
	dnsmasqPort := 0
	bindEnabled := false
	bindPort := 53
	bindAddresses := `["192.0.2.2", "198.51.100.231"]`

	switch phase {
	case "prepared":
	case "active":
		bindEnabled = true
	case "legacy":
		unboundEnabled = true
		dnsmasqPort = 53053
		bindPort = 53531
		bindAddresses = `["127.0.0.1"]`
	default:
		panic("unknown DNS cutover phase: " + phase)
	}

	return fmt.Sprintf(`
resource "opnsense_unbound_service" "test" {
  enabled = %[1]t
}

resource "opnsense_dnsmasq_settings" "test" {
  enabled  = true
  dns_port = %[2]d
}

resource "opnsense_interfaces_loopback" "dns" {
  description = "DNS internal service network"
}

resource "opnsense_interfaces_assignment" "dns" {
  device          = format("lo%%d", opnsense_interfaces_loopback.dns.device_id)
  description     = "DNS internal service network"
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

resource "opnsense_interfaces_vip" "dns_public" {
  mode        = "ipalias"
  interface   = "wan"
  network     = "198.51.100.231/32"
  description = "DNS public frontend"
}

resource "opnsense_bind_settings" "test" {
  enabled      = %[3]t
  disable_ipv6 = true
  listen_ipv4  = %[4]s
  listen_ipv6  = ["::1"]
  port         = %[5]d

  depends_on = [
    opnsense_unbound_service.test,
    opnsense_dnsmasq_settings.test,
    opnsense_interfaces_assignment.dns,
    opnsense_interfaces_vip.dns_public,
  ]
}
`, unboundEnabled, dnsmasqPort, bindEnabled, bindAddresses, bindPort)
}
func testAccDNSCutoverCleanupConfig(keepLoopback bool) string {
	loopback := ""
	if keepLoopback {
		loopback = `
resource "opnsense_interfaces_loopback" "dns" {
  description = "DNS internal service network"
}
`
	}
	return fmt.Sprintf(`
resource "opnsense_unbound_service" "test" {
  enabled = true
}

resource "opnsense_dnsmasq_settings" "test" {
  enabled  = true
  dns_port = 53053
}
%s
resource "opnsense_bind_settings" "test" {
  enabled      = false
  disable_ipv6 = true
  listen_ipv4  = ["127.0.0.1"]
  listen_ipv6  = ["::1"]
  port         = 53531
}
`, loopback)
}

func checkOPNsenseAPIStable() resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		endpoint := strings.TrimRight(os.Getenv("OPNSENSE_URI"), "/") + "/api/bind/general/get"
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true", // #nosec G402 -- acceptance targets use self-signed certificates.
			},
		}
		defer transport.CloseIdleConnections()

		client := &http.Client{Timeout: 10 * time.Second, Transport: transport}
		probe := func(ctx context.Context) error {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if err != nil {
				return err
			}
			req.SetBasicAuth(os.Getenv("OPNSENSE_API_KEY"), os.Getenv("OPNSENSE_API_SECRET"))

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			_, _ = io.Copy(io.Discard, resp.Body)
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("BIND API readiness probe returned HTTP %d", resp.StatusCode)
			}
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := waitForConsecutiveSuccesses(ctx, 2*time.Second, 5, probe); err != nil {
			return fmt.Errorf("OPNsense API did not remain stable after interface reconfiguration: %w", err)
		}
		return nil
	}
}

func waitForConsecutiveSuccesses(
	ctx context.Context,
	interval time.Duration,
	required int,
	probe func(context.Context) error,
) error {
	if required < 1 {
		return fmt.Errorf("required successes must be at least one")
	}

	consecutive := 0
	var lastErr error
	for {
		if err := probe(ctx); err != nil {
			lastErr = err
			consecutive = 0
		} else {
			consecutive++
			if consecutive >= required {
				return nil
			}
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			if lastErr != nil {
				return fmt.Errorf("%w: last probe error: %v", ctx.Err(), lastErr)
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func checkDNSPort53(expectedProcess string, expectedIPv4, expectedIPv6 []string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		expected4 := dnsPort53ExpectedTokens(expectedProcess, expectedIPv4)
		expected6 := dnsPort53ExpectedTokens(expectedProcess, expectedIPv6)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		var lastOutput string
		for ctx.Err() == nil {
			output, err := acctest.QGAGuestExec(ctx, "/usr/bin/sockstat", "-4", "-6", "-l")
			if err == nil {
				lastOutput = output
				tcp4 := dnsPort53Tokens(output, "tcp4")
				udp4 := dnsPort53Tokens(output, "udp4")
				tcp6 := dnsPort53Tokens(output, "tcp6")
				udp6 := dnsPort53Tokens(output, "udp6")
				if equalStringSlices(tcp4, expected4) &&
					equalStringSlices(udp4, expected4) &&
					equalStringSlices(tcp6, expected6) &&
					equalStringSlices(udp6, expected6) {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf(
			"DNS listeners on port 53 did not converge to IPv4 %v and IPv6 %v: %s",
			expected4,
			expected6,
			strings.TrimSpace(lastOutput),
		)
	}
}

func dnsPort53ExpectedTokens(process string, addresses []string) []string {
	result := make([]string, 0, len(addresses))
	for _, address := range addresses {
		result = append(result, process+"@"+address)
	}
	sort.Strings(result)
	return result
}
func dnsPort53Tokens(output, protocol string) []string {
	values := make(map[string]struct{})
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 || fields[4] != protocol {
			continue
		}
		localAddress := fields[5]
		if !strings.HasSuffix(localAddress, ":53") {
			continue
		}
		address := strings.TrimSuffix(localAddress, ":53")
		values[fields[1]+"@"+address] = struct{}{}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func equalStringSlices(left, right []string) bool {
	return strings.Join(left, "\x00") == strings.Join(right, "\x00")
}
