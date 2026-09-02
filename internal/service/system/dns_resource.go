package system

import (
	"context"
	"fmt"

	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &dnsSettingsResource{}
var _ resource.ResourceWithConfigure = &dnsSettingsResource{}
var _ resource.ResourceWithImportState = &dnsSettingsResource{}

type dnsSettingsResource struct{ client opnsense.Client }

func newDnsSettingsResource() resource.Resource { return &dnsSettingsResource{} }

func (r *dnsSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_dns"
}

func (r *dnsSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = dnsSettingsResourceSchema()
}

func (r *dnsSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *dnsSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data dnsSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure System DNS", err.Error())
		return
	}
	state, err := r.readModel(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read System DNS Settings After Apply", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *dnsSettingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state, err := r.readModel(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read System DNS Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *dnsSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data dnsSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure System DNS", err.Error())
		return
	}
	state, err := r.readModel(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read System DNS Settings After Apply", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *dnsSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Singleton Resource Removed From State Only",
		"System DNS settings remain unchanged in OPNsense. Re-import with ID `system_dns` to manage them again.",
	)
}

func (r *dnsSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "system_dns" {
		resp.Diagnostics.AddError("Invalid Import ID", "The system DNS singleton must be imported with ID `system_dns`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *dnsSettingsResource) apply(ctx context.Context, data *dnsSettingsResourceModel) error {
	desired, err := dnsToAPI(ctx, data)
	if err != nil {
		return err
	}
	currentResponse, err := r.client.ApiExtensions().DnsGet(ctx)
	if err != nil {
		return err
	}
	if currentResponse == nil {
		return fmt.Errorf("system DNS get API returned an empty response")
	}
	if dnsEqual(&currentResponse.DNS, desired) {
		return nil
	}

	result, err := r.client.ApiExtensions().DnsSet(ctx, desired)
	if err != nil {
		return err
	}
	if err = validateDNSAction("settings update", result); err != nil {
		return err
	}
	reconfigured, err := r.client.ApiExtensions().DnsReconfigure(ctx)
	if err != nil {
		return err
	}
	if err = validateDNSAction("reconfigure", reconfigured); err != nil {
		return err
	}
	tflog.Trace(ctx, "configured system DNS resolver settings")
	return nil
}

func (r *dnsSettingsResource) readModel(ctx context.Context) (*dnsSettingsResourceModel, error) {
	result, err := r.client.ApiExtensions().DnsGet(ctx)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("system DNS get API returned an empty response")
	}
	return dnsFromAPI(ctx, &result.DNS)
}

func dnsToAPI(ctx context.Context, data *dnsSettingsResourceModel) (*apiextensions.DnsSettings, error) {
	var servers []string
	diagnostics := data.Servers.ElementsAs(ctx, &servers, false)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("unable to decode system DNS servers: %v", diagnostics.Errors())
	}
	return &apiextensions.DnsSettings{
		Servers:         servers,
		AllowOverride:   data.AllowOverride.ValueBool(),
		UseLocalService: data.UseLocalService.ValueBool(),
	}, nil
}

func dnsFromAPI(ctx context.Context, data *apiextensions.DnsSettings) (*dnsSettingsResourceModel, error) {
	servers, diagnostics := types.ListValueFrom(ctx, types.StringType, data.Servers)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("unable to encode system DNS servers: %v", diagnostics.Errors())
	}
	return &dnsSettingsResourceModel{
		Servers:         servers,
		AllowOverride:   types.BoolValue(data.AllowOverride),
		UseLocalService: types.BoolValue(data.UseLocalService),
		ID:              types.StringValue("system_dns"),
	}, nil
}

func dnsEqual(current, desired *apiextensions.DnsSettings) bool {
	if current.AllowOverride != desired.AllowOverride ||
		current.UseLocalService != desired.UseLocalService ||
		len(current.Servers) != len(desired.Servers) {
		return false
	}
	for index := range current.Servers {
		if current.Servers[index] != desired.Servers[index] {
			return false
		}
	}
	return true
}
