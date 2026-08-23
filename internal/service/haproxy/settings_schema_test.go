package haproxy

import (
	"testing"

	apihaproxy "github.com/biptec/opnsense-go/pkg/haproxy"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCompleteSettingsModelPreservesExistingSingletonDefaults(t *testing.T) {
	t.Parallel()
	plan := &settingsModel{
		Enabled:         types.BoolUnknown(),
		ShowIntro:       types.BoolUnknown(),
		GracefulStop:    types.BoolUnknown(),
		HardStopAfter:   types.StringUnknown(),
		CloseSpreadTime: types.StringUnknown(),
		SeamlessReload:  types.BoolUnknown(),
	}
	current := settingsAPIToModel(&apihaproxy.SettingsResponse{HAProxy: apihaproxy.Settings{General: apihaproxy.GeneralSettings{
		Enabled: "0", ShowIntro: "1", GracefulStop: "1", HardStopAfter: "45s", CloseSpreadTime: "7s", SeamlessReload: "1",
	}}})

	completeSettingsModel(plan, current)

	if plan.Enabled.ValueBool() || !plan.ShowIntro.ValueBool() || !plan.GracefulStop.ValueBool() || !plan.SeamlessReload.ValueBool() {
		t.Fatalf("existing boolean settings were not preserved: %+v", plan)
	}
	if plan.HardStopAfter.ValueString() != "45s" || plan.CloseSpreadTime.ValueString() != "7s" {
		t.Fatalf("existing string settings were not preserved: %+v", plan)
	}
}

func TestCompleteSettingsModelKeepsExplicitOverrides(t *testing.T) {
	t.Parallel()
	plan := &settingsModel{
		Enabled:         types.BoolValue(true),
		ShowIntro:       types.BoolValue(false),
		GracefulStop:    types.BoolValue(false),
		HardStopAfter:   types.StringValue("60s"),
		CloseSpreadTime: types.StringValue("10s"),
		SeamlessReload:  types.BoolValue(false),
	}
	current := settingsAPIToModel(&apihaproxy.SettingsResponse{HAProxy: apihaproxy.Settings{General: apihaproxy.GeneralSettings{
		Enabled: "0", ShowIntro: "1", GracefulStop: "1", HardStopAfter: "45s", CloseSpreadTime: "7s", SeamlessReload: "1",
	}}})

	completeSettingsModel(plan, current)

	if !plan.Enabled.ValueBool() || plan.ShowIntro.ValueBool() || plan.GracefulStop.ValueBool() || plan.SeamlessReload.ValueBool() {
		t.Fatalf("explicit boolean overrides were replaced: %+v", plan)
	}
	if plan.HardStopAfter.ValueString() != "60s" || plan.CloseSpreadTime.ValueString() != "10s" {
		t.Fatalf("explicit string overrides were replaced: %+v", plan)
	}
}

func TestSettingsResourceSchemaSupportsAutomaticAdoption(t *testing.T) {
	t.Parallel()
	schema := settingsResourceSchema()
	for _, name := range []string{"enabled", "show_intro", "graceful_stop", "hard_stop_after", "close_spread_time", "seamless_reload"} {
		attribute := schema.Attributes[name]
		if !attribute.IsOptional() || !attribute.IsComputed() {
			t.Fatalf("%s must remain Optional+Computed for safe singleton adoption", name)
		}
	}
}
