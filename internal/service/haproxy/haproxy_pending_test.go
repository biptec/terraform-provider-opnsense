package haproxy_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type haproxyDiffResponse struct {
	Response string `json:"response"`
}

func checkHAProxyNoPendingDiff() resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		req, err := http.NewRequest(http.MethodPost, os.Getenv("OPNSENSE_URI")+"/api/haproxy/export/diff/", nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(os.Getenv("OPNSENSE_API_KEY"), os.Getenv("OPNSENSE_API_SECRET"))
		client := &http.Client{
			Timeout:   15 * time.Second,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true"}},
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("read HAProxy config diff: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("read HAProxy config diff: HTTP %d", resp.StatusCode)
		}
		var payload haproxyDiffResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return fmt.Errorf("decode HAProxy config diff: %w", err)
		}
		if diff := strings.TrimSpace(payload.Response); diff != "" {
			return fmt.Errorf("HAProxy still has pending configuration changes: %s", diff)
		}
		return nil
	}
}
