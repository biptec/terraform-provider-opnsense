package quagga

import (
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func configureQuaggaResource(req resource.ConfigureRequest, resp *resource.ConfigureResponse) opnsense.Client {
	if req.ProviderData == nil {
		return nil
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return nil
	}
	return opnsense.NewClient(c)
}

func configureQuaggaDataSource(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) opnsense.Client {
	if req.ProviderData == nil {
		return nil
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return nil
	}
	return opnsense.NewClient(c)
}

func optionalBoolToAPI(v types.Bool) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return tools.BoolToString(v.ValueBool())
}
func optionalBoolFromAPI(v string) types.Bool {
	if v == "" {
		return types.BoolNull()
	}
	return types.BoolValue(tools.StringToBool(v))
}
func optionalIntToAPI(v types.Int64) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return tools.Int64ToString(v.ValueInt64())
}
func optionalIntFromAPI(v string) types.Int64 {
	if v == "" {
		return types.Int64Null()
	}
	return types.Int64Value(tools.StringToInt64(v))
}

func validateRoutingSet(result *api.ActionResult) error {
	if result == nil || result.Result != "saved" {
		if result == nil {
			return fmt.Errorf("set returned no result")
		}
		return fmt.Errorf("set returned result %q instead of saved", result.Result)
	}
	return nil
}

func validateRoutingReconfigure(result *api.ReconfigureStatusResult) error {
	if result == nil || result.Status != "ok" {
		if result == nil {
			return fmt.Errorf("reconfigure returned no status")
		}
		return fmt.Errorf("reconfigure returned status %q instead of ok", result.Status)
	}
	return nil
}
