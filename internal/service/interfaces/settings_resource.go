package interfaces

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &interfaceSettingsResource{}
var _ resource.ResourceWithConfigure = &interfaceSettingsResource{}
var _ resource.ResourceWithImportState = &interfaceSettingsResource{}

type interfaceSettingsResource struct{ client opnsense.Client }

func newInterfaceSettingsResource() resource.Resource { return &interfaceSettingsResource{} }

func (r *interfaceSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_settings"
}
func (r *interfaceSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = interfaceSettingsResourceSchema()
}
func (r *interfaceSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(client)
}

func (r *interfaceSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data interfaceSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure Interface Settings", err.Error())
		return
	}
	state, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Interface Settings Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *interfaceSettingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	state, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Interface Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *interfaceSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data interfaceSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure Interface Settings", err.Error())
		return
	}
	state, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Interface Settings Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *interfaceSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Singleton Resource Removed From State Only", "Global interface settings remain unchanged in OPNsense. Re-import with ID `interfaces_settings` to manage them again.")
}

func (r *interfaceSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "interfaces_settings" {
		resp.Diagnostics.AddError("Invalid Import ID", "The interfaces settings singleton must be imported with ID `interfaces_settings`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *interfaceSettingsResource) apply(ctx context.Context, data *interfaceSettingsResourceModel) error {
	result, err := r.client.Interfaces().SettingsSet(ctx, convertInterfaceSettingsSchemaToStruct(data))
	if err != nil {
		return err
	}
	if result.Result != "saved" {
		return fmt.Errorf("unexpected settings result %q", result.Result)
	}
	reconfigured, err := r.client.Interfaces().SettingsReconfigure(ctx)
	if err != nil {
		return err
	}
	if reconfigured.Status != "ok" {
		return fmt.Errorf("unexpected reconfigure status %q", reconfigured.Status)
	}
	tflog.Trace(ctx, "configured global interface settings")
	return nil
}

func (r *interfaceSettingsResource) read(ctx context.Context) (*interfaceSettingsResourceModel, error) {
	result, err := r.client.Interfaces().SettingsGet(ctx)
	if err != nil {
		return nil, err
	}
	return convertInterfaceSettingsStructToSchema(&result.Settings), nil
}
