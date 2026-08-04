package routing

import (
	"context"
	"errors"
	"testing"
	"time"

	apirouting "github.com/biptec/opnsense-go/pkg/routing"
)

func gatewayGroupPendingError() error {
	return errors.New(
		"resource not changed. result: failed. errors: " +
			"map[gateway_group.item:Option [TFACC_GROUP_GW] not in list.]",
	)
}

func TestGatewayGroupDependencyPending(t *testing.T) {
	if !gatewayGroupDependencyPending(gatewayGroupPendingError()) {
		t.Fatal("expected gateway option validation error to be retryable")
	}
	if gatewayGroupDependencyPending(errors.New("permission denied")) {
		t.Fatal("unrelated errors must not be retryable")
	}
}

func TestAddGatewayGroupRetriesPendingDependency(t *testing.T) {
	calls := 0
	add := func(context.Context, *apirouting.GatewayGroup) (string, error) {
		calls++
		if calls < 3 {
			return "", gatewayGroupPendingError()
		}
		return "group-id", nil
	}

	id, err := addGatewayGroupWhenDependenciesReady(
		context.Background(), add, &apirouting.GatewayGroup{},
		time.Millisecond, 100*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("addGatewayGroupWhenDependenciesReady() error = %v", err)
	}
	if id != "group-id" || calls != 3 {
		t.Fatalf("result = %q after %d calls", id, calls)
	}
}

func TestAddGatewayGroupDoesNotRetryUnrelatedError(t *testing.T) {
	calls := 0
	wantErr := errors.New("permission denied")
	add := func(context.Context, *apirouting.GatewayGroup) (string, error) {
		calls++
		return "", wantErr
	}

	_, err := addGatewayGroupWhenDependenciesReady(
		context.Background(), add, &apirouting.GatewayGroup{},
		time.Millisecond, 100*time.Millisecond,
	)
	if !errors.Is(err, wantErr) || calls != 1 {
		t.Fatalf("error = %v after %d calls", err, calls)
	}
}

func TestAddGatewayGroupDoesNotRetryAfterCreatedID(t *testing.T) {
	calls := 0
	pendingErr := gatewayGroupPendingError()
	add := func(context.Context, *apirouting.GatewayGroup) (string, error) {
		calls++
		return "created-id", pendingErr
	}

	id, err := addGatewayGroupWhenDependenciesReady(
		context.Background(), add, &apirouting.GatewayGroup{},
		time.Millisecond, 100*time.Millisecond,
	)
	if id != "created-id" || !errors.Is(err, pendingErr) || calls != 1 {
		t.Fatalf("result = %q, error = %v, calls = %d", id, err, calls)
	}
}

func TestAddGatewayGroupStopsAtTimeout(t *testing.T) {
	calls := 0
	add := func(context.Context, *apirouting.GatewayGroup) (string, error) {
		calls++
		return "", gatewayGroupPendingError()
	}

	_, err := addGatewayGroupWhenDependenciesReady(
		context.Background(), add, &apirouting.GatewayGroup{},
		50*time.Millisecond, 5*time.Millisecond,
	)
	if !errors.Is(err, context.DeadlineExceeded) || calls != 1 {
		t.Fatalf("error = %v after %d calls", err, calls)
	}
}
