package system_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/biptec/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func checkNtpRuntimeListener(address string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if acctest.QGASocket() == "" {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var lastOutput string
		for ctx.Err() == nil {
			output, err := acctest.QGAGuestExec(ctx, "/usr/bin/sockstat", "-4", "-l", "-P", "udp")
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
