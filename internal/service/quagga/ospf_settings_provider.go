package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/opnsense-go/pkg/quagga"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
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

type ospfSettingsModel struct {
	Enabled             types.Bool   `tfsdk:"enabled"`
	CarpDemote          types.Bool   `tfsdk:"carp_demote"`
	RouterID            types.String `tfsdk:"router_id"`
	CostReference       types.Int64  `tfsdk:"cost_reference"`
	LogAdjacencyChanges types.Bool   `tfsdk:"log_adjacency_changes"`
	Originate           types.Bool   `tfsdk:"originate"`
	OriginateAlways     types.Bool   `tfsdk:"originate_always"`
	OriginateMetric     types.Int64  `tfsdk:"originate_metric"`
	PassiveInterfaces   types.Set    `tfsdk:"passive_interfaces"`
	ID                  types.String `tfsdk:"id"`
}

func ospfSettingsResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages OPNsense FRR/OSPF singleton settings.", Attributes: map[string]schema.Attribute{
		"enabled":               schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"carp_demote":           schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"router_id":             schema.StringAttribute{Optional: true, Computed: true},
		"cost_reference":        schema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(1, 4294967)}},
		"log_adjacency_changes": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"originate":             schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"originate_always":      schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"originate_metric":      schema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 16777214)}},
		"passive_interfaces":    schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"id":                    schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func ospfSettingsDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads OPNsense FRR/OSPF singleton settings.", Attributes: map[string]dsschema.Attribute{
		"enabled":               dsschema.BoolAttribute{Computed: true},
		"carp_demote":           dsschema.BoolAttribute{Computed: true},
		"router_id":             dsschema.StringAttribute{Computed: true},
		"cost_reference":        dsschema.Int64Attribute{Computed: true},
		"log_adjacency_changes": dsschema.BoolAttribute{Computed: true},
		"originate":             dsschema.BoolAttribute{Computed: true},
		"originate_always":      dsschema.BoolAttribute{Computed: true},
		"originate_metric":      dsschema.Int64Attribute{Computed: true},
		"passive_interfaces":    dsschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"id":                    dsschema.StringAttribute{Computed: true},
	}}
}

func ospfSettingsToAPI(ctx context.Context, d *ospfSettingsModel) *quagga.OSPFSettingsData {
	var passiveInterfaces []string
	_ = d.PassiveInterfaces.ElementsAs(ctx, &passiveInterfaces, false)
	return &quagga.OSPFSettingsData{
		Enabled:             optionalBoolToAPI(d.Enabled),
		CarpDemote:          optionalBoolToAPI(d.CarpDemote),
		RouterID:            d.RouterID.ValueString(),
		CostReference:       optionalIntToAPI(d.CostReference),
		LogAdjacencyChanges: optionalBoolToAPI(d.LogAdjacencyChanges),
		Originate:           optionalBoolToAPI(d.Originate),
		OriginateAlways:     optionalBoolToAPI(d.OriginateAlways),
		OriginateMetric:     optionalIntToAPI(d.OriginateMetric),
		PassiveInterfaces:   api.SelectedMapList(passiveInterfaces),
	}
}

func ospfSettingsFromAPI(d *quagga.OSPFSettingsData) *ospfSettingsModel {
	return &ospfSettingsModel{
		Enabled:             optionalBoolFromAPI(d.Enabled),
		CarpDemote:          optionalBoolFromAPI(d.CarpDemote),
		RouterID:            types.StringValue(d.RouterID),
		CostReference:       optionalIntFromAPI(d.CostReference),
		LogAdjacencyChanges: optionalBoolFromAPI(d.LogAdjacencyChanges),
		Originate:           optionalBoolFromAPI(d.Originate),
		OriginateAlways:     optionalBoolFromAPI(d.OriginateAlways),
		OriginateMetric:     optionalIntFromAPI(d.OriginateMetric),
		PassiveInterfaces:   tools.StringSliceToSet([]string(d.PassiveInterfaces)),
		ID:                  types.StringValue("ospf_settings"),
	}
}

var _ resource.Resource = &ospfSettingsResource{}
var _ resource.ResourceWithConfigure = &ospfSettingsResource{}
var _ resource.ResourceWithImportState = &ospfSettingsResource{}

type ospfSettingsResource struct{ client opnsense.Client }

func newOspfSettingsResource() resource.Resource { return &ospfSettingsResource{} }

func (r *ospfSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_settings"
}
func (r *ospfSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospfSettingsResourceSchema()
}
func (r *ospfSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}

func (r *ospfSettingsResource) apply(ctx context.Context, data *ospfSettingsModel) error {
	setResult, err := r.client.Quagga().OSPFSettingsSet(ctx, ospfSettingsToAPI(ctx, data))
	if err != nil {
		return err
	}
	if err := validateRoutingSet(setResult); err != nil {
		return err
	}
	reconfigureResult, err := r.client.Quagga().OSPFSettingsReconfigure(ctx)
	if err != nil {
		return err
	}
	return validateRoutingReconfigure(reconfigureResult)
}

func (r *ospfSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospfSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Apply OPNsense Routing Settings", err.Error())
		return
	}
	remote, err := r.client.Quagga().OSPFSettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Routing Settings Applied but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfSettingsFromAPI(&remote.OSPF))...)
}
func (r *ospfSettingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	remote, err := r.client.Quagga().OSPFSettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Routing Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfSettingsFromAPI(&remote.OSPF))...)
}
func (r *ospfSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospfSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Apply OPNsense Routing Settings", err.Error())
		return
	}
	remote, err := r.client.Quagga().OSPFSettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Routing Settings Applied but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfSettingsFromAPI(&remote.OSPF))...)
}
func (r *ospfSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Routing singleton removed from Terraform state", "OPNsense singleton settings are not reset on destroy.")
}
func (r *ospfSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "ospf_settings" {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected ospf_settings.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &ospfSettingsDataSource{}
var _ datasource.DataSourceWithConfigure = &ospfSettingsDataSource{}

type ospfSettingsDataSource struct{ client opnsense.Client }

func newOspfSettingsDataSource() datasource.DataSource { return &ospfSettingsDataSource{} }

func (d *ospfSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_settings"
}
func (d *ospfSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospfSettingsDataSourceSchema()
}
func (d *ospfSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}

func (d *ospfSettingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	remote, err := d.client.Quagga().OSPFSettingsGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Routing Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfSettingsFromAPI(&remote.OSPF))...)
}
