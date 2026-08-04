package routing

import (
	"context"
	"fmt"
	"strings"
	"time"

	apirouting "github.com/biptec/opnsense-go/pkg/routing"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	gatewayGroupRetryInterval = 2 * time.Second
	gatewayGroupRetryTimeout  = 30 * time.Second
)

func gatewayGroupDependencyPending(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "gateway_group.") &&
		strings.Contains(message, "Option [") &&
		strings.Contains(message, "not in list")
}

func addGatewayGroupWhenDependenciesReady(
	ctx context.Context,
	add func(context.Context, *apirouting.GatewayGroup) (string, error),
	group *apirouting.GatewayGroup,
	interval time.Duration,
	timeout time.Duration,
) (string, error) {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var lastErr error
	for attempt := 1; ; attempt++ {
		id, err := add(waitCtx, group)
		if err == nil || id != "" || !gatewayGroupDependencyPending(err) {
			return id, err
		}
		lastErr = err
		tflog.Debug(ctx, "routing gateway group dependency is not ready", map[string]any{
			"attempt": attempt,
			"error":   err.Error(),
		})

		timer := time.NewTimer(interval)
		select {
		case <-waitCtx.Done():
			timer.Stop()
			return "", fmt.Errorf(
				"gateway group dependencies did not become ready: %v: %w",
				lastErr,
				waitCtx.Err(),
			)
		case <-timer.C:
		}
	}
}

func retryGatewayGroupAdd(
	ctx context.Context,
	add func(context.Context, *apirouting.GatewayGroup) (string, error),
	group *apirouting.GatewayGroup,
) (string, error) {
	return addGatewayGroupWhenDependenciesReady(
		ctx, add, group, gatewayGroupRetryInterval, gatewayGroupRetryTimeout,
	)
}
