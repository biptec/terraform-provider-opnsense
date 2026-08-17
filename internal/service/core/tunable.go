package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	coreapi "github.com/biptec/opnsense-go/pkg/core"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tunableModel struct {
	ID          types.String `tfsdk:"id"`
	Tunable     types.String `tfsdk:"tunable"`
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
}

func tunableResourceSchema() rschema.Schema {
	return rschema.Schema{MarkdownDescription: "Manages one OPNsense system tunable.", Attributes: map[string]rschema.Attribute{
		"id":          rschema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"tunable":     rschema.StringAttribute{Required: true},
		"value":       rschema.StringAttribute{Required: true},
		"description": rschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
	}}
}
func tunableDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads one OPNsense system tunable by UUID.", Attributes: map[string]dsschema.Attribute{
		"id": dsschema.StringAttribute{Required: true}, "tunable": dsschema.StringAttribute{Computed: true},
		"value": dsschema.StringAttribute{Computed: true}, "description": dsschema.StringAttribute{Computed: true},
	}}
}

func tunableToAPI(data *tunableModel) *coreapi.Tunable {
	return &coreapi.Tunable{Tunable: data.Tunable.ValueString(), Value: data.Value.ValueString(), Description: data.Description.ValueString()}
}
func tunableFromAPI(data *coreapi.Tunable, id string) *tunableModel {
	return &tunableModel{ID: types.StringValue(id), Tunable: types.StringValue(data.Tunable), Value: types.StringValue(data.Value), Description: types.StringValue(data.Description)}
}

var _ resource.Resource = &tunableResource{}
var _ resource.ResourceWithConfigure = &tunableResource{}
var _ resource.ResourceWithImportState = &tunableResource{}

type tunableResource struct{ client opnsense.Client }

func newTunableResource() resource.Resource { return &tunableResource{} }
func (r *tunableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_core_tunable"
}
func (r *tunableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tunableResourceSchema()
}
func (r *tunableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(c)
}
func (r *tunableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data tunableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Core().AddTunable(ctx, tunableToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OPNsense Tunable", err.Error())
		return
	}
	remote, err := r.client.Core().GetTunable(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OPNsense Tunable Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tunableFromAPI(remote, id))...)
}
func (r *tunableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data tunableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Core().GetTunable(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OPNsense Tunable", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tunableFromAPI(remote, data.ID.ValueString()))...)
}
func (r *tunableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data tunableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Core().UpdateTunable(ctx, data.ID.ValueString(), tunableToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OPNsense Tunable", err.Error())
		return
	}
	remote, err := r.client.Core().GetTunable(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OPNsense Tunable Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tunableFromAPI(remote, data.ID.ValueString()))...)
}
func (r *tunableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data tunableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Core().DeleteTunable(ctx, data.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if !errors.As(err, &notFound) {
			resp.Diagnostics.AddError("Unable to Delete OPNsense Tunable", err.Error())
		}
	}
}
func (r *tunableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &tunableDataSource{}
var _ datasource.DataSourceWithConfigure = &tunableDataSource{}

type tunableDataSource struct{ client opnsense.Client }

func newTunableDataSource() datasource.DataSource { return &tunableDataSource{} }
func (d *tunableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_core_tunable"
}
func (d *tunableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tunableDataSourceSchema()
}
func (d *tunableDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(c)
}
func (d *tunableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tunableModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Core().GetTunable(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Tunable", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, tunableFromAPI(remote, data.ID.ValueString()))...)
}
