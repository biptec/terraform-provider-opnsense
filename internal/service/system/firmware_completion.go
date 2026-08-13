package system

import (
	"strings"

	apicore "github.com/biptec/opnsense-go/pkg/core"
)

func firmwareOperationComplete(running *apicore.FirmwareRunningResponse, runningErr error, status *apicore.FirmwareUpgradeStatusResponse, statusErr error) bool {
	if runningErr == nil && running != nil && strings.EqualFold(strings.TrimSpace(running.Status), "ready") {
		return true
	}
	return statusErr == nil && status != nil && strings.EqualFold(strings.TrimSpace(status.Status), "done")
}
