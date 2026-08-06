package bind_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
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

func TestWaitForConsecutiveSuccessesResetsAfterFailure(t *testing.T) {
	transient := errors.New("transient API restart")
	results := []error{nil, nil, transient, nil, nil, nil}
	calls := 0

	err := waitForConsecutiveSuccesses(context.Background(), 0, 3, func(context.Context) error {
		result := results[calls]
		calls++
		return result
	})
	if err != nil {
		t.Fatalf("waitForConsecutiveSuccesses() error = %v", err)
	}
	if calls != len(results) {
		t.Fatalf("expected %d probes after reset, got %d", len(results), calls)
	}
}

func TestWaitForConsecutiveSuccessesReportsLastFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	lastErr := errors.New("connection reset by peer")

	err := waitForConsecutiveSuccesses(ctx, 0, 2, func(context.Context) error {
		cancel()
		return lastErr
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if !strings.Contains(err.Error(), lastErr.Error()) {
		t.Fatalf("expected last probe error in result, got %v", err)
	}
}

func TestWaitForConsecutiveSuccessesRejectsInvalidCount(t *testing.T) {
	err := waitForConsecutiveSuccesses(context.Background(), 0, 0, func(context.Context) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "at least one") {
		t.Fatalf("expected invalid required-success count error, got %v", err)
	}
}
