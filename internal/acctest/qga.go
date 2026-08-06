package acctest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
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

// QGASocket returns the configured QEMU Guest Agent socket path.
func QGASocket() string {
	return os.Getenv("QEMU_GA_SOCKET")
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

// QGAGuestExec executes a command in the disposable OPNsense guest.
func QGAGuestExec(ctx context.Context, executable string, arguments ...string) (string, error) {
	socketPath := QGASocket()
	if socketPath == "" {
		return "", fmt.Errorf("QEMU_GA_SOCKET is not configured")
	}
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
			return string(stdout), fmt.Errorf("guest command exited %d: %s", status.ExitCode, string(stderr))
		}
		return string(stdout), nil
	}
}
