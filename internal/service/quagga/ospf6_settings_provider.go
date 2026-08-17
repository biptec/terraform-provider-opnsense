package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/opnsense-go/pkg/quagga"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ospf6SettingsModel struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	CarpDemote      types.Bool   `tfsdk:"carp_demote"`
	RouterID        types.String `tfsdk:"router_id"`
	Originate       types.Bool   `tfsdk:"originate"`
	OriginateAlways types.Bool   `tfsdk:"originate_always"`
	OriginateMetric types.Int64  `tfsdk:"originate_metric"`
	ID              types.String `tfsdk:"id"`
}

func ospf6SettingsResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages OPNsense FRR/OSPF singleton settings.", Attributes: map[string]schema.Attribute{
		"enabled":          schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"carp_demote":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"router_id":        schema.StringAttribute{Optional: true, Computed: true},
		"originate":        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"originate_always": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"originate_metric": schema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 16777214)}},
		"id":               schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func ospf6SettingsDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads OPNsense FRR/OSPF singleton settings.", Attributes: map[string]dsschema.Attribute{
		"enabled":          dsschema.BoolAttribute{Computed: true},
		"carp_demote":      dsschema.BoolAttribute{Computed: true},
		"router_id":        dsschema.StringAttribute{Computed: true},
		"originate":        dsschema.BoolAttribute{Computed: true},
		"originate_always": dsschema.BoolAttribute{Computed: true},
		"originate_metric": dsschema.Int64Attribute{Computed: true},
		"id":               dsschema.StringAttribute{Computed: true},
	}}
}

func ospf6SettingsToAPI(ctx context.Context, d *ospf6SettingsModel) *quagga.OSPF6SettingsData {
	return &quagga.OSPF6SettingsData{
		Enabled:         optionalBoolToAPI(d.Enabled),
		CarpDemote:      optionalBoolToAPI(d.CarpDemote),
		RouterID:        d.RouterID.ValueString(),
		Originate:       optionalBoolToAPI(d.Originate),
		OriginateAlways: optionalBoolToAPI(d.OriginateAlways),
		OriginateMetric: optionalIntToAPI(d.OriginateMetric),
	}
}

func ospf6SettingsFromAPI(d *quagga.OSPF6SettingsData) *ospf6SettingsModel {
	return &ospf6SettingsModel{
		Enabled:         optionalBoolFromAPI(d.Enabled),
		CarpDemote:      optionalBoolFromAPI(d.CarpDemote),
		RouterID:        types.StringValue(d.RouterID),
		Originate:       optionalBoolFromAPI(d.Originate),
		OriginateAlways: optionalBoolFromAPI(d.OriginateAlways),
		OriginateMetric: optionalIntFromAPI(d.OriginateMetric),
		ID:              types.StringValue("ospf6_settings"),
	}
}

var _ resource.Resource = &ospf6SettingsResource{}
var _ resource.ResourceWithConfigure = &ospf6SettingsResource{}
var _ resource.ResourceWithImportState = &ospf6SettingsResource{}

type ospf6SettingsResource struct{ client opnsense.Client }

func newOspf6SettingsResource() resource.Resource { return &ospf6SettingsResource{} }

func (r *ospf6SettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_settings"
}
func (r *ospf6SettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospf6SettingsResourceSchema()
}
func (r *ospf6SettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}

func (r *ospf6SettingsResource) apply(ctx context.Context, data *ospf6SettingsModel) error {
	setResult, err := r.client.Quagga().OSPF6SettingsSet(ctx, ospf6SettingsToAPI(ctx, data))
	if err != nil {
		return err
	}
	if err := validateRoutingSet(setResult); err != nil {
		return err
	}
	reconfigureResult, err := r.client.Quagga().OSPF6SettingsReconfigure(ctx)
	if err != nil {
		return err
	}
	return validateRoutingReconfigure(reconfigureResult)
}

func (r *ospf6SettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospf6SettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Apply OPNsense Routing Settings", err.Error())
		return
	}
	remote, err := r.client.Quagga().OSPF6SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Routing Settings Applied but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6SettingsFromAPI(&remote.OSPF6))...)
}
func (r *ospf6SettingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	remote, err := r.client.Quagga().OSPF6SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Routing Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6SettingsFromAPI(&remote.OSPF6))...)
}
func (r *ospf6SettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospf6SettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Apply OPNsense Routing Settings", err.Error())
		return
	}
	remote, err := r.client.Quagga().OSPF6SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Routing Settings Applied but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6SettingsFromAPI(&remote.OSPF6))...)
}
func (r *ospf6SettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Routing singleton removed from Terraform state", "OPNsense singleton settings are not reset on destroy.")
}
func (r *ospf6SettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "ospf6_settings" {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected ospf6_settings.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &ospf6SettingsDataSource{}
var _ datasource.DataSourceWithConfigure = &ospf6SettingsDataSource{}

type ospf6SettingsDataSource struct{ client opnsense.Client }

func newOspf6SettingsDataSource() datasource.DataSource { return &ospf6SettingsDataSource{} }

func (d *ospf6SettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_settings"
}
func (d *ospf6SettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospf6SettingsDataSourceSchema()
}
func (d *ospf6SettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}

func (d *ospf6SettingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	remote, err := d.client.Quagga().OSPF6SettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Routing Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6SettingsFromAPI(&remote.OSPF6))...)
}
