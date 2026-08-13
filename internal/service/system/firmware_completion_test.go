package system

import (
	"errors"
	"testing"

	apicore "github.com/biptec/opnsense-go/pkg/core"
)

func TestFirmwareOperationComplete(t *testing.T) {
	ready := &apicore.FirmwareRunningResponse{Status: "ready"}
	busy := &apicore.FirmwareRunningResponse{Status: "busy"}
	done := &apicore.FirmwareUpgradeStatusResponse{Status: "done"}
	running := &apicore.FirmwareUpgradeStatusResponse{Status: "running"}

	if !firmwareOperationComplete(ready, nil, running, nil) {
		t.Fatal("ready firmware state was not accepted")
	}
	if !firmwareOperationComplete(busy, nil, done, nil) {
		t.Fatal("completed firmware action with busy global lock was not accepted")
	}
	if firmwareOperationComplete(busy, nil, running, nil) {
		t.Fatal("running firmware action was accepted")
	}
	if firmwareOperationComplete(nil, errors.New("running failed"), nil, errors.New("status failed")) {
		t.Fatal("failed firmware endpoints were accepted")
	}
}
