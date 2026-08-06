package firewall

import (
	"testing"

	"github.com/biptec/opnsense-go/pkg/api"
	opnfirewall "github.com/biptec/opnsense-go/pkg/firewall"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func baseNATResourceModel() *natResourceModel {
	return &natResourceModel{
		Enabled:     types.BoolValue(true),
		DisableNAT:  types.BoolValue(false),
		Sequence:    types.Int64Value(100),
		Interface:   types.StringValue("wan"),
		IPProtocol:  types.StringValue("inet"),
		Protocol:    types.StringValue("any"),
		Source:      &firewallLocation{Net: types.StringValue("198.51.100.112/29"), Port: types.StringValue(""), Invert: types.BoolValue(false)},
		Destination: &firewallLocation{Net: types.StringValue("any"), Port: types.StringValue(""), Invert: types.BoolValue(false)},
		Log:         types.BoolValue(false),
		Description: types.StringValue("Routed public subnet"),
	}
}

func TestConvertNATSchemaToStructNoNATWithoutTarget(t *testing.T) {
	model := baseNATResourceModel()
	model.DisableNAT = types.BoolValue(true)

	result, err := convertNATSchemaToStruct(model)

	require.NoError(t, err)
	require.Equal(t, "1", result.DisableNAT)
	require.Empty(t, result.Target)
	require.Empty(t, result.TargetPort)
}

func TestConvertNATSchemaToStructUsesInterfaceAddressWithoutTarget(t *testing.T) {
	result, err := convertNATSchemaToStruct(baseNATResourceModel())

	require.NoError(t, err)
	require.Equal(t, "0", result.DisableNAT)
	require.Empty(t, result.Target)
	require.Empty(t, result.TargetPort)
}

func TestConvertNATSchemaToStructExplicitTarget(t *testing.T) {
	model := baseNATResourceModel()
	model.Target = &firewallTarget{IP: types.StringValue("wanip"), Port: types.StringValue("443")}

	result, err := convertNATSchemaToStruct(model)

	require.NoError(t, err)
	require.Equal(t, "wanip", result.Target)
	require.Equal(t, "443", result.TargetPort)
}

func TestValidateNATTargetConfiguration(t *testing.T) {
	t.Run("no NAT rejects target", func(t *testing.T) {
		model := baseNATResourceModel()
		model.DisableNAT = types.BoolValue(true)
		model.Target = &firewallTarget{IP: types.StringValue("wanip"), Port: types.StringValue("")}
		require.EqualError(t, validateNATTargetConfiguration(model), "target must be omitted when disable_nat is true")
	})

	t.Run("port requires IP", func(t *testing.T) {
		model := baseNATResourceModel()
		model.Target = &firewallTarget{IP: types.StringValue(""), Port: types.StringValue("443")}
		require.EqualError(t, validateNATTargetConfiguration(model), "target.ip must be set when target.port is configured")
	})

	t.Run("empty block must be omitted", func(t *testing.T) {
		model := baseNATResourceModel()
		model.Target = &firewallTarget{IP: types.StringValue(""), Port: types.StringValue("")}
		require.EqualError(t, validateNATTargetConfiguration(model), "empty target block must be omitted")
	})
}

func TestConvertNATStructToSchemaNormalizesEmptyTarget(t *testing.T) {
	model, err := convertNATStructToSchema(&opnfirewall.NAT{
		Enabled:        "1",
		DisableNAT:     "1",
		Sequence:       "100",
		Interface:      api.SelectedMap("wan"),
		IPProtocol:     api.SelectedMap("inet"),
		Protocol:       api.SelectedMap("any"),
		SourceNet:      "198.51.100.112/29",
		SourceInvert:   "0",
		DestinationNet: "any",
		Description:    "Routed public subnet",
	})

	require.NoError(t, err)
	require.True(t, model.DisableNAT.ValueBool())
	require.Nil(t, model.Target)
}

func TestConvertNATStructToSchemaPreservesTarget(t *testing.T) {
	model, err := convertNATStructToSchema(&opnfirewall.NAT{
		Enabled:        "1",
		DisableNAT:     "0",
		Sequence:       "100",
		Interface:      api.SelectedMap("wan"),
		IPProtocol:     api.SelectedMap("inet"),
		Protocol:       api.SelectedMap("tcp"),
		SourceNet:      "lan",
		DestinationNet: "any",
		Target:         "wanip",
		TargetPort:     "443",
	})

	require.NoError(t, err)
	require.NotNil(t, model.Target)
	require.Equal(t, "wanip", model.Target.IP.ValueString())
	require.Equal(t, "443", model.Target.Port.ValueString())
}

func TestValidateSourceNATSettingsResults(t *testing.T) {
	require.EqualError(t, validateSourceNATSettingsSetResult(nil), "source NAT settings API returned an empty response")
	require.EqualError(t, validateSourceNATSettingsSetResult(&api.ActionResult{Result: "failed"}), `source NAT settings API returned result "failed" instead of "saved"`)
	require.NoError(t, validateSourceNATSettingsSetResult(&api.ActionResult{Result: "saved"}))

	require.EqualError(t, validateSourceNATSettingsApplyResult(nil), "source NAT apply API returned an empty response")
	require.EqualError(t, validateSourceNATSettingsApplyResult(&opnfirewall.SourceNATApplyResult{Status: "failed"}), `source NAT apply API returned status "failed" instead of OK`)
	require.NoError(t, validateSourceNATSettingsApplyResult(&opnfirewall.SourceNATApplyResult{Status: "OK\n\n"}))
}
