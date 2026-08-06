package system_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type qgaError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

type qgaEnvelope struct {
	Return json.RawMessage `json:"return"`
	Error  *qgaError       `json:"error"`
	Event  string          `json:"event"`
}

func qgaCall(ctx context.Context, socketPath string, request any, result any) error {
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
		var envelope qgaEnvelope
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

func qgaExec(ctx context.Context, socketPath, executable string, arguments ...string) (string, error) {
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
	if err := qgaCall(ctx, socketPath, request, &started); err != nil {
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
		if err := qgaCall(ctx, socketPath, statusRequest, &status); err != nil {
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
			return string(stdout), fmt.Errorf("guest command exited %d: %s", status.ExitCode, strings.TrimSpace(string(stderr)))
		}
		return string(stdout), nil
	}
}

func checkNtpRuntimeListener(address string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		socketPath := os.Getenv("QEMU_GA_SOCKET")
		if socketPath == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var lastOutput string
		for ctx.Err() == nil {
			output, err := qgaExec(ctx, socketPath, "/usr/bin/sockstat", "-4", "-l", "-P", "udp")
			if err == nil {
				lastOutput = output
				expected := address + ":123"
				if strings.Contains(output, expected) &&
					!strings.Contains(output, "0.0.0.0:123") &&
					!strings.Contains(output, "*:123") {
					return nil
				}
			}
			time.Sleep(time.Second)
		}
		return fmt.Errorf("NTP did not bind exclusively to %s: %s", address, strings.TrimSpace(lastOutput))
	}
}
