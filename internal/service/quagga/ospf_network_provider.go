package quagga

import (
	"context"
	"errors"
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/opnsense-go/pkg/quagga"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ospfNetworkModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	IPAddress     types.String `tfsdk:"ip_address"`
	Area          types.String `tfsdk:"area"`
	Netmask       types.Int64  `tfsdk:"netmask"`
	AreaRange     types.String `tfsdk:"area_range"`
	PrefixListIn  types.String `tfsdk:"prefix_list_in"`
	PrefixListOut types.String `tfsdk:"prefix_list_out"`
	ID            types.String `tfsdk:"id"`
}

func ospfNetworkResourceSchema() rschema.Schema {
	return rschema.Schema{MarkdownDescription: "Manages one OPNsense OSPF object.", Attributes: map[string]rschema.Attribute{
		"enabled":         rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"ip_address":      rschema.StringAttribute{Required: true, Validators: []validator.String{validators.IPv4Address()}},
		"area":            rschema.StringAttribute{Required: true},
		"netmask":         rschema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(24), Validators: []validator.Int64{int64validator.Between(0, 32)}},
		"area_range":      rschema.StringAttribute{Optional: true, Computed: true},
		"prefix_list_in":  rschema.StringAttribute{Optional: true, Computed: true},
		"prefix_list_out": rschema.StringAttribute{Optional: true, Computed: true},
		"id":              rschema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func ospfNetworkDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads one OPNsense OSPF object by UUID.", Attributes: map[string]dsschema.Attribute{
		"enabled":         dsschema.BoolAttribute{Computed: true},
		"ip_address":      dsschema.StringAttribute{Computed: true},
		"area":            dsschema.StringAttribute{Computed: true},
		"netmask":         dsschema.Int64Attribute{Computed: true},
		"area_range":      dsschema.StringAttribute{Computed: true},
		"prefix_list_in":  dsschema.StringAttribute{Computed: true},
		"prefix_list_out": dsschema.StringAttribute{Computed: true},
		"id":              dsschema.StringAttribute{Required: true},
	}}
}

func ospfNetworkToAPI(d *ospfNetworkModel) *quagga.OSPFNetwork {
	return &quagga.OSPFNetwork{
		Enabled:       optionalBoolToAPI(d.Enabled),
		IPAddress:     d.IPAddress.ValueString(),
		Area:          d.Area.ValueString(),
		Netmask:       optionalIntToAPI(d.Netmask),
		AreaRange:     d.AreaRange.ValueString(),
		PrefixListIn:  api.SelectedMap(d.PrefixListIn.ValueString()),
		PrefixListOut: api.SelectedMap(d.PrefixListOut.ValueString()),
	}
}

func ospfNetworkFromAPI(d *quagga.OSPFNetwork, id string) *ospfNetworkModel {
	return &ospfNetworkModel{
		Enabled:       optionalBoolFromAPI(d.Enabled),
		IPAddress:     types.StringValue(d.IPAddress),
		Area:          types.StringValue(d.Area),
		Netmask:       optionalIntFromAPI(d.Netmask),
		AreaRange:     types.StringValue(d.AreaRange),
		PrefixListIn:  types.StringValue(d.PrefixListIn.String()),
		PrefixListOut: types.StringValue(d.PrefixListOut.String()),
		ID:            types.StringValue(id),
	}
}

var _ resource.Resource = &ospfNetworkResource{}
var _ resource.ResourceWithConfigure = &ospfNetworkResource{}
var _ resource.ResourceWithImportState = &ospfNetworkResource{}

type ospfNetworkResource struct{ client opnsense.Client }

func newOspfNetworkResource() resource.Resource { return &ospfNetworkResource{} }

func (r *ospfNetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_network"
}
func (r *ospfNetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospfNetworkResourceSchema()
}
func (r *ospfNetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}

func (r *ospfNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospfNetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPFNetwork(ctx, ospfNetworkToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OPNsense OSPF Object", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFNetwork(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPF Object Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfNetworkFromAPI(remote, id))...)
}
func (r *ospfNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospfNetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPFNetwork(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OPNsense OSPF Object", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfNetworkFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospfNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospfNetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPFNetwork(ctx, data.ID.ValueString(), ospfNetworkToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OPNsense OSPF Object", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFNetwork(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPF Object Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfNetworkFromAPI(remote, data.ID.ValueString()))...)
}

func (r *ospfNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospfNetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().DeleteOSPFNetwork(ctx, data.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if !errors.As(err, &notFound) {
			resp.Diagnostics.AddError("Unable to Delete OPNsense OSPF Object", err.Error())
		}
	}
}
func (r *ospfNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &ospfNetworkDataSource{}
var _ datasource.DataSourceWithConfigure = &ospfNetworkDataSource{}

type ospfNetworkDataSource struct{ client opnsense.Client }

func newOspfNetworkDataSource() datasource.DataSource { return &ospfNetworkDataSource{} }

func (d *ospfNetworkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_network"
}
func (d *ospfNetworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospfNetworkDataSourceSchema()
}
func (d *ospfNetworkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}

func (d *ospfNetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospfNetworkModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPFNetwork(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense OSPF Object", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfNetworkFromAPI(remote, data.ID.ValueString()))...)
}
