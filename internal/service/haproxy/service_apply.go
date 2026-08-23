package haproxy

import (
	"context"
	"fmt"
	"strings"
)

func (r *resourceClient) applyHAProxyConfig(ctx context.Context) error {
	checked, err := r.client.Haproxy().ServiceConfigtest(ctx)
	if err != nil {
		return fmt.Errorf("configuration test: %w", err)
	}
	if checked == nil {
		return fmt.Errorf("configuration test returned no result")
	}
	result := strings.ToLower(checked.Result)
	if strings.Contains(result, "error") || strings.Contains(result, "failed") || strings.Contains(result, "alert") {
		return fmt.Errorf("configuration test result: %s", checked.Result)
	}

	reconfigured, err := r.client.Haproxy().ServiceReconfigure(ctx)
	if err != nil {
		return fmt.Errorf("reconfigure HAProxy: %w", err)
	}
	if reconfigured != nil && reconfigured.Status != "" && !strings.EqualFold(reconfigured.Status, "ok") {
		return fmt.Errorf("reconfigure HAProxy status: %s", reconfigured.Status)
	}
	return nil
}
