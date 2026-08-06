package system

import (
	"context"
	"fmt"
	"sort"

	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ntpSettingsResource{}
var _ resource.ResourceWithConfigure = &ntpSettingsResource{}
var _ resource.ResourceWithImportState = &ntpSettingsResource{}

type ntpSettingsResource struct{ client opnsense.Client }

func newNtpSettingsResource() resource.Resource { return &ntpSettingsResource{} }

func (r *ntpSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ntp_settings"
}

func (r *ntpSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ntpSettingsResourceSchema()
}

func (r *ntpSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *ntpSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ntpSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure NTP", err.Error())
		return
	}
	data.ID = types.StringValue("ntp_settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ntpSettingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	result, err := r.client.ApiExtensions().NtpGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read NTP Settings", err.Error())
		return
	}
	state, err := ntpFromAPI(ctx, &result.NTP)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Convert NTP Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ntpSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ntpSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure NTP", err.Error())
		return
	}
	data.ID = types.StringValue("ntp_settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ntpSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Singleton Resource Removed From State Only",
		"NTP settings remain unchanged in OPNsense. Re-import with ID `ntp_settings` to manage them again.",
	)
}

func (r *ntpSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "ntp_settings" {
		resp.Diagnostics.AddError("Invalid Import ID", "The NTP settings singleton must be imported with ID `ntp_settings`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *ntpSettingsResource) apply(ctx context.Context, data *ntpSettingsResourceModel) error {
	desired, err := ntpToAPI(ctx, data)
	if err != nil {
		return err
	}
	currentResponse, err := r.client.ApiExtensions().NtpGet(ctx)
	if err != nil {
		return err
	}
	if currentResponse == nil {
		return fmt.Errorf("NTP get API returned an empty response")
	}
	if ntpEqual(&currentResponse.NTP, desired) {
		return nil
	}

	result, err := r.client.ApiExtensions().NtpSet(ctx, desired)
	if err != nil {
		return err
	}
	if err = validateNTPAction("settings update", result); err != nil {
		return err
	}
	reconfigured, err := r.client.ApiExtensions().NtpReconfigure(ctx)
	if err != nil {
		return err
	}
	if err = validateNTPAction("reconfigure", reconfigured); err != nil {
		return err
	}
	tflog.Trace(ctx, "configured NTP settings")
	return nil
}

func ntpToAPI(ctx context.Context, data *ntpSettingsResourceModel) (*apiextensions.NtpSettings, error) {
	interfaces, err := stringSet(ctx, data.Interfaces)
	if err != nil {
		return nil, err
	}
	var serverModels []ntpServerModel
	diagnostics := data.Servers.ElementsAs(ctx, &serverModels, false)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("unable to decode NTP servers: %v", diagnostics.Errors())
	}
	servers := make([]apiextensions.NtpServer, 0, len(serverModels))
	for _, server := range serverModels {
		servers = append(servers, apiextensions.NtpServer{
			Host:     server.Host.ValueString(),
			NoSelect: server.NoSelect.ValueBool(),
			Prefer:   server.Prefer.ValueBool(),
			IBurst:   server.IBurst.ValueBool(),
			Pool:     server.Pool.ValueBool(),
		})
	}
	sort.Slice(servers, func(i, j int) bool { return servers[i].Host < servers[j].Host })
	return &apiextensions.NtpSettings{
		Enabled:              data.Enabled.ValueBool(),
		Servers:              servers,
		Interfaces:           interfaces,
		Orphan:               int(data.Orphan.ValueInt64()),
		MaxClock:             int(data.MaxClock.ValueInt64()),
		ClientMode:           data.ClientMode.ValueBool(),
		KissOfDeath:          data.KissOfDeath.ValueBool(),
		RateLimiting:         data.RateLimiting.ValueBool(),
		DenyModifications:    data.DenyModifications.ValueBool(),
		DisableQueries:       data.DisableQueries.ValueBool(),
		DisableServing:       data.DisableServing.ValueBool(),
		DenyPeerAssociations: data.DenyPeerAssociations.ValueBool(),
		DenyTrapService:      data.DenyTrapService.ValueBool(),
	}, nil
}

func ntpFromAPI(ctx context.Context, data *apiextensions.NtpSettings) (*ntpSettingsResourceModel, error) {
	servers := make([]ntpServerModel, 0, len(data.Servers))
	for _, server := range data.Servers {
		servers = append(servers, ntpServerModel{
			Host:     types.StringValue(server.Host),
			NoSelect: types.BoolValue(server.NoSelect),
			Prefer:   types.BoolValue(server.Prefer),
			IBurst:   types.BoolValue(server.IBurst),
			Pool:     types.BoolValue(server.Pool),
		})
	}
	serverSet, diagnostics := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ntpServerAttributeTypes}, servers)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("unable to encode NTP servers: %v", diagnostics.Errors())
	}
	return &ntpSettingsResourceModel{
		Enabled:              types.BoolValue(data.Enabled),
		Servers:              serverSet,
		Interfaces:           stringSetValue(data.Interfaces),
		Orphan:               types.Int64Value(int64(data.Orphan)),
		MaxClock:             types.Int64Value(int64(data.MaxClock)),
		ClientMode:           types.BoolValue(data.ClientMode),
		KissOfDeath:          types.BoolValue(data.KissOfDeath),
		RateLimiting:         types.BoolValue(data.RateLimiting),
		DenyModifications:    types.BoolValue(data.DenyModifications),
		DisableQueries:       types.BoolValue(data.DisableQueries),
		DisableServing:       types.BoolValue(data.DisableServing),
		DenyPeerAssociations: types.BoolValue(data.DenyPeerAssociations),
		DenyTrapService:      types.BoolValue(data.DenyTrapService),
		ID:                   types.StringValue("ntp_settings"),
	}, nil
}

func ntpEqual(current, desired *apiextensions.NtpSettings) bool {
	if current.Enabled != desired.Enabled ||
		current.Orphan != desired.Orphan ||
		current.MaxClock != desired.MaxClock ||
		current.ClientMode != desired.ClientMode ||
		current.KissOfDeath != desired.KissOfDeath ||
		current.RateLimiting != desired.RateLimiting ||
		current.DenyModifications != desired.DenyModifications ||
		current.DisableQueries != desired.DisableQueries ||
		current.DisableServing != desired.DisableServing ||
		current.DenyPeerAssociations != desired.DenyPeerAssociations ||
		current.DenyTrapService != desired.DenyTrapService ||
		!sameStrings(current.Interfaces, desired.Interfaces) ||
		len(current.Servers) != len(desired.Servers) {
		return false
	}
	currentServers := make(map[string]apiextensions.NtpServer, len(current.Servers))
	for _, server := range current.Servers {
		currentServers[server.Host] = server
	}
	for _, server := range desired.Servers {
		if currentServers[server.Host] != server {
			return false
		}
	}
	return true
}
