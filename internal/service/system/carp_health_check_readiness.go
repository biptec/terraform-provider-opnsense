package system

import (
	"context"
	"fmt"
	"strconv"
	"time"

	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
)

const (
	carpHealthHealthyPollInterval = 200 * time.Millisecond
	carpHealthHealthyMinTimeout   = 30 * time.Second
)

type carpHealthStatusClient interface {
	CarpHealthStatus(context.Context) (*apiextensions.CarpHealthStatusResponse, error)
}

type carpHealthReadinessClient interface {
	carpHealthStatusClient
	CarpHealthGet(context.Context) (*apiextensions.CarpHealthResponse, error)
}

func waitForCarpHealthCheckHealthy(
	ctx context.Context,
	client carpHealthStatusClient,
	id string,
	pollInterval time.Duration,
	timeout time.Duration,
) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	lastState := "runtime status unavailable"
	var lastErr error
	for {
		status, err := client.CarpHealthStatus(ctx)
		if err != nil {
			lastErr = err
			lastState = "status request failed"
		} else {
			lastErr = nil
			lastState = "check not present in current runtime state"
			if status != nil && !status.Running {
				lastState = "health monitor is not running"
			}
			if status != nil && status.Running {
				for _, check := range status.Checks {
					if check.UUID != id {
						continue
					}
					if check.Healthy {
						return nil
					}
					lastState = "probe is not healthy"
					break
				}
			}
		}

		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-deadline.C:
			timer.Stop()
			if lastErr != nil {
				return fmt.Errorf("CARP health check %s did not become probe-healthy: %w", id, lastErr)
			}
			return fmt.Errorf("CARP health check %s did not become probe-healthy within %s (%s)", id, timeout, lastState)
		case <-timer.C:
		}
	}
}

func carpHealthReadinessTimeout(settings *apiextensions.CarpHealthSettings) time.Duration {
	interval := int64(1)
	recoveryThreshold := int64(2)
	if settings != nil {
		if parsed, err := strconv.ParseInt(settings.Interval, 10, 64); err == nil && parsed > 0 {
			interval = parsed
		}
		if parsed, err := strconv.ParseInt(settings.RecoveryThreshold, 10, 64); err == nil && parsed > 0 {
			recoveryThreshold = parsed
		}
	}

	// One interval may elapse before the first probe after a monitor restart.
	// Two additional intervals give the monitor room to absorb a concurrent
	// check reconfigure while still honoring deliberately slow health settings.
	timeout := time.Duration(interval*(recoveryThreshold+2)) * time.Second
	if timeout < carpHealthHealthyMinTimeout {
		return carpHealthHealthyMinTimeout
	}
	return timeout
}

func waitForCarpHealthCheckHealthyDefault(ctx context.Context, client carpHealthReadinessClient, id string) error {
	settings, err := client.CarpHealthGet(ctx)
	if err != nil {
		return fmt.Errorf("read CARP health settings before readiness wait: %w", err)
	}
	return waitForCarpHealthCheckHealthy(
		ctx,
		client,
		id,
		carpHealthHealthyPollInterval,
		carpHealthReadinessTimeout(&settings.CarpHealth),
	)
}

func carpHealthCheckStateFromRemote(data *carpHealthCheckResourceModel, remote *apiextensions.CarpHealthCheck, id string) *carpHealthCheckResourceModel {
	state := carpHealthCheckFromAPI(remote, id)
	state.WaitForHealthy = data.WaitForHealthy
	return state
}

func (r *carpHealthCheckResource) waitForHealthyIfRequested(ctx context.Context, data *carpHealthCheckResourceModel, id string) error {
	if !data.WaitForHealthy.ValueBool() {
		return nil
	}
	return waitForCarpHealthCheckHealthyDefault(ctx, r.client.ApiExtensions(), id)
}
