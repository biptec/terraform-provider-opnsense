package bind_test

import (
	"reflect"
	"testing"
)

func TestDNSPort53Tokens(t *testing.T) {
	output := `
USER     COMMAND PID FD PROTO  LOCAL ADDRESS         FOREIGN ADDRESS
root     named   100 21 tcp4   192.0.2.2:53         *:*
root     named   100 22 tcp4   198.51.100.231:53    *:*
root     named   100 23 tcp4   192.0.2.2:53         *:*
root     named   100 24 udp4   192.0.2.2:53         *:*
root     named   100 25 udp4   198.51.100.231:53    *:*
root     unbound 101 26 tcp6   *:53                  *:*
root     unbound 101 27 udp6   *:53                  *:*
root     unbound 101 28 tcp4   *:953                 *:*
root     dnsmasq 102 27 tcp4   *:53053               *:*
`
	want := []string{
		"named@192.0.2.2",
		"named@198.51.100.231",
	}
	for _, protocol := range []string{"tcp4", "udp4"} {
		if got := dnsPort53Tokens(output, protocol); !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected %s port 53 tokens: want %v, got %v", protocol, want, got)
		}
	}

	want6 := []string{"unbound@*"}
	for _, protocol := range []string{"tcp6", "udp6"} {
		if got := dnsPort53Tokens(output, protocol); !reflect.DeepEqual(got, want6) {
			t.Fatalf("unexpected %s port 53 tokens: want %v, got %v", protocol, want6, got)
		}
	}
}
