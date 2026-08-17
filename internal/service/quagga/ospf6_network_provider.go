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

type ospf6NetworkModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	IPAddress     types.String `tfsdk:"ip_address"`
	Netmask       types.Int64  `tfsdk:"netmask"`
	Area          types.String `tfsdk:"area"`
	AreaRange     types.String `tfsdk:"area_range"`
	PrefixListIn  types.String `tfsdk:"prefix_list_in"`
	PrefixListOut types.String `tfsdk:"prefix_list_out"`
	ID            types.String `tfsdk:"id"`
}

func ospf6NetworkResourceSchema() rschema.Schema {
	return rschema.Schema{MarkdownDescription: "Manages one OPNsense OSPF object.", Attributes: map[string]rschema.Attribute{
		"enabled":         rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"ip_address":      rschema.StringAttribute{Required: true, Validators: []validator.String{validators.IPv6Address()}},
		"netmask":         rschema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(64), Validators: []validator.Int64{int64validator.Between(0, 128)}},
		"area":            rschema.StringAttribute{Required: true},
		"area_range":      rschema.StringAttribute{Optional: true, Computed: true},
		"prefix_list_in":  rschema.StringAttribute{Optional: true, Computed: true},
		"prefix_list_out": rschema.StringAttribute{Optional: true, Computed: true},
		"id":              rschema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func ospf6NetworkDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads one OPNsense OSPF object by UUID.", Attributes: map[string]dsschema.Attribute{
		"enabled":         dsschema.BoolAttribute{Computed: true},
		"ip_address":      dsschema.StringAttribute{Computed: true},
		"netmask":         dsschema.Int64Attribute{Computed: true},
		"area":            dsschema.StringAttribute{Computed: true},
		"area_range":      dsschema.StringAttribute{Computed: true},
		"prefix_list_in":  dsschema.StringAttribute{Computed: true},
		"prefix_list_out": dsschema.StringAttribute{Computed: true},
		"id":              dsschema.StringAttribute{Required: true},
	}}
}

func ospf6NetworkToAPI(d *ospf6NetworkModel) *quagga.OSPF6Network {
	return &quagga.OSPF6Network{
		Enabled:       optionalBoolToAPI(d.Enabled),
		IPAddress:     d.IPAddress.ValueString(),
		Netmask:       optionalIntToAPI(d.Netmask),
		Area:          d.Area.ValueString(),
		AreaRange:     d.AreaRange.ValueString(),
		PrefixListIn:  api.SelectedMap(d.PrefixListIn.ValueString()),
		PrefixListOut: api.SelectedMap(d.PrefixListOut.ValueString()),
	}
}

func ospf6NetworkFromAPI(d *quagga.OSPF6Network, id string) *ospf6NetworkModel {
	return &ospf6NetworkModel{
		Enabled:       optionalBoolFromAPI(d.Enabled),
		IPAddress:     types.StringValue(d.IPAddress),
		Netmask:       optionalIntFromAPI(d.Netmask),
		Area:          types.StringValue(d.Area),
		AreaRange:     types.StringValue(d.AreaRange),
		PrefixListIn:  types.StringValue(d.PrefixListIn.String()),
		PrefixListOut: types.StringValue(d.PrefixListOut.String()),
		ID:            types.StringValue(id),
	}
}

var _ resource.Resource = &ospf6NetworkResource{}
var _ resource.ResourceWithConfigure = &ospf6NetworkResource{}
var _ resource.ResourceWithImportState = &ospf6NetworkResource{}

type ospf6NetworkResource struct{ client opnsense.Client }

func newOspf6NetworkResource() resource.Resource { return &ospf6NetworkResource{} }

func (r *ospf6NetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_network"
}
func (r *ospf6NetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospf6NetworkResourceSchema()
}
func (r *ospf6NetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}

func (r *ospf6NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospf6NetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPF6Network(ctx, ospf6NetworkToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OPNsense OSPF Object", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6Network(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPF Object Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6NetworkFromAPI(remote, id))...)
}
func (r *ospf6NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospf6NetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPF6Network(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OPNsense OSPF Object", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6NetworkFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospf6NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospf6NetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPF6Network(ctx, data.ID.ValueString(), ospf6NetworkToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OPNsense OSPF Object", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6Network(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPF Object Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6NetworkFromAPI(remote, data.ID.ValueString()))...)
}

func (r *ospf6NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospf6NetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().DeleteOSPF6Network(ctx, data.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if !errors.As(err, &notFound) {
			resp.Diagnostics.AddError("Unable to Delete OPNsense OSPF Object", err.Error())
		}
	}
}
func (r *ospf6NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &ospf6NetworkDataSource{}
var _ datasource.DataSourceWithConfigure = &ospf6NetworkDataSource{}

type ospf6NetworkDataSource struct{ client opnsense.Client }

func newOspf6NetworkDataSource() datasource.DataSource { return &ospf6NetworkDataSource{} }

func (d *ospf6NetworkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_network"
}
func (d *ospf6NetworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospf6NetworkDataSourceSchema()
}
func (d *ospf6NetworkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}

func (d *ospf6NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospf6NetworkModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPF6Network(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense OSPF Object", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6NetworkFromAPI(remote, data.ID.ValueString()))...)
}
