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

var _ resource.Resource = &sshResource{}
var _ resource.ResourceWithConfigure = &sshResource{}
var _ resource.ResourceWithImportState = &sshResource{}

type sshResource struct{ client opnsense.Client }

func newSshResource() resource.Resource { return &sshResource{} }

func (r *sshResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_ssh"
}

func (r *sshResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = sshResourceSchema()
}

func (r *sshResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *sshResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data sshResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure SSH", err.Error())
		return
	}
	result, err := r.client.ApiExtensions().SshGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SSH Settings After Apply", err.Error())
		return
	}
	state := sshFromAPI(&result.SSH, data.AllowReaddress.ValueBool())
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *sshResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var prior sshResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.ApiExtensions().SshGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SSH Settings", err.Error())
		return
	}
	state := sshFromAPI(&result.SSH, prior.AllowReaddress.ValueBool())
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *sshResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data sshResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure SSH", err.Error())
		return
	}
	result, err := r.client.ApiExtensions().SshGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SSH Settings After Apply", err.Error())
		return
	}
	state := sshFromAPI(&result.SSH, data.AllowReaddress.ValueBool())
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *sshResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Singleton Resource Removed From State Only",
		"SSH settings remain unchanged in OPNsense. Re-import with ID `system_ssh` to manage them again.",
	)
}

func (r *sshResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "system_ssh" {
		resp.Diagnostics.AddError("Invalid Import ID", "The SSH singleton must be imported with ID `system_ssh`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("allow_readdress"), false)...)
}

func (r *sshResource) apply(ctx context.Context, data *sshResourceModel) error {
	currentResponse, err := r.client.ApiExtensions().SshGet(ctx)
	if err != nil {
		return err
	}
	if currentResponse == nil {
		return fmt.Errorf("SSH get API returned an empty response")
	}
	current := &currentResponse.SSH
	desired, err := sshToAPI(ctx, data, current)
	if err != nil {
		return err
	}
	if sshEqual(current, desired) {
		return nil
	}
	if sshDisruptiveChange(current, desired) && !data.AllowReaddress.ValueBool() {
		return fmt.Errorf("SSH enablement, port, or listener interface change requires allow_readdress = true")
	}

	result, err := r.client.ApiExtensions().SshSet(ctx, desired)
	if err != nil {
		return err
	}
	if err = validateSSHAction("settings update", result); err != nil {
		return err
	}
	reconfigured, err := r.client.ApiExtensions().SshReconfigure(ctx)
	if err != nil {
		return err
	}
	if err = validateSSHAction("reconfigure", reconfigured); err != nil {
		return err
	}
	tflog.Trace(ctx, "configured SSH listener settings")
	return nil
}

func sshToAPI(ctx context.Context, data *sshResourceModel, current *apiextensions.SshSettings) (*apiextensions.SshSettings, error) {
	interfaces, err := stringSet(ctx, data.Interfaces)
	if err != nil {
		return nil, err
	}
	result := *current
	result.Interfaces = interfaces
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		result.Enabled = data.Enabled.ValueBool()
	}
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		result.Port = int(data.Port.ValueInt64())
	}
	if !data.PasswordAuthentication.IsNull() && !data.PasswordAuthentication.IsUnknown() {
		result.PasswordAuthentication = data.PasswordAuthentication.ValueBool()
	}
	if !data.PermitRootLogin.IsNull() && !data.PermitRootLogin.IsUnknown() {
		result.PermitRootLogin = data.PermitRootLogin.ValueBool()
	}
	return &result, nil
}

func sshFromAPI(data *apiextensions.SshSettings, allowReaddress bool) *sshResourceModel {
	return &sshResourceModel{
		Enabled:                types.BoolValue(data.Enabled),
		Port:                   types.Int64Value(int64(data.Port)),
		Interfaces:             stringSetValue(data.Interfaces),
		PasswordAuthentication: types.BoolValue(data.PasswordAuthentication),
		PermitRootLogin:        types.BoolValue(data.PermitRootLogin),
		AllowReaddress:         types.BoolValue(allowReaddress),
		ID:                     types.StringValue("system_ssh"),
	}
}

func sshDisruptiveChange(current, desired *apiextensions.SshSettings) bool {
	return current.Enabled != desired.Enabled ||
		current.Port != desired.Port ||
		!sameStrings(current.Interfaces, desired.Interfaces)
}

func sshEqual(current, desired *apiextensions.SshSettings) bool {
	return !sshDisruptiveChange(current, desired) &&
		current.PasswordAuthentication == desired.PasswordAuthentication &&
		current.PermitRootLogin == desired.PermitRootLogin
}
