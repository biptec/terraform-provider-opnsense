package quagga

import (
	"context"
	"errors"
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/opnsense-go/pkg/quagga"
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

type ospfInterfaceModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	InterfaceName      types.String `tfsdk:"interface_name"`
	AuthType           types.String `tfsdk:"auth_type"`
	AuthKey            types.String `tfsdk:"auth_key"`
	AuthKeyID          types.Int64  `tfsdk:"auth_key_id"`
	Area               types.String `tfsdk:"area"`
	Cost               types.Int64  `tfsdk:"cost"`
	CostDemoted        types.Int64  `tfsdk:"cost_demoted"`
	CarpDependOn       types.String `tfsdk:"carp_depend_on"`
	HelloInterval      types.Int64  `tfsdk:"hello_interval"`
	DeadInterval       types.Int64  `tfsdk:"dead_interval"`
	RetransmitInterval types.Int64  `tfsdk:"retransmit_interval"`
	TransmitDelay      types.Int64  `tfsdk:"transmit_delay"`
	Priority           types.Int64  `tfsdk:"priority"`
	BFD                types.Bool   `tfsdk:"bfd"`
	NetworkType        types.String `tfsdk:"network_type"`
	P2MPOptions        types.String `tfsdk:"p2mp_options"`
	ID                 types.String `tfsdk:"id"`
}

func ospfInterfaceResourceSchema() rschema.Schema {
	return rschema.Schema{MarkdownDescription: "Manages one OPNsense OSPF object.", Attributes: map[string]rschema.Attribute{
		"enabled":             rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"interface_name":      rschema.StringAttribute{Optional: true, Computed: true},
		"auth_type":           rschema.StringAttribute{Optional: true, Computed: true},
		"auth_key":            rschema.StringAttribute{Optional: true, Computed: true},
		"auth_key_id":         rschema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(1), Validators: []validator.Int64{int64validator.Between(1, 255)}},
		"area":                rschema.StringAttribute{Required: true},
		"cost":                rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(1, 65535)}},
		"cost_demoted":        rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(1, 65535)}},
		"carp_depend_on":      rschema.StringAttribute{Optional: true, Computed: true},
		"hello_interval":      rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"dead_interval":       rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"retransmit_interval": rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"transmit_delay":      rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"priority":            rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"bfd":                 rschema.BoolAttribute{Optional: true, Computed: true},
		"network_type":        rschema.StringAttribute{Optional: true, Computed: true},
		"p2mp_options":        rschema.StringAttribute{Optional: true, Computed: true},
		"id":                  rschema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func ospfInterfaceDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads one OPNsense OSPF object by UUID.", Attributes: map[string]dsschema.Attribute{
		"enabled":             dsschema.BoolAttribute{Computed: true},
		"interface_name":      dsschema.StringAttribute{Computed: true},
		"auth_type":           dsschema.StringAttribute{Computed: true},
		"auth_key":            dsschema.StringAttribute{Computed: true},
		"auth_key_id":         dsschema.Int64Attribute{Computed: true},
		"area":                dsschema.StringAttribute{Computed: true},
		"cost":                dsschema.Int64Attribute{Computed: true},
		"cost_demoted":        dsschema.Int64Attribute{Computed: true},
		"carp_depend_on":      dsschema.StringAttribute{Computed: true},
		"hello_interval":      dsschema.Int64Attribute{Computed: true},
		"dead_interval":       dsschema.Int64Attribute{Computed: true},
		"retransmit_interval": dsschema.Int64Attribute{Computed: true},
		"transmit_delay":      dsschema.Int64Attribute{Computed: true},
		"priority":            dsschema.Int64Attribute{Computed: true},
		"bfd":                 dsschema.BoolAttribute{Computed: true},
		"network_type":        dsschema.StringAttribute{Computed: true},
		"p2mp_options":        dsschema.StringAttribute{Computed: true},
		"id":                  dsschema.StringAttribute{Required: true},
	}}
}

func ospfInterfaceToAPI(d *ospfInterfaceModel) *quagga.OSPFInterface {
	return &quagga.OSPFInterface{
		Enabled:            optionalBoolToAPI(d.Enabled),
		InterfaceName:      api.SelectedMap(d.InterfaceName.ValueString()),
		AuthType:           api.SelectedMap(d.AuthType.ValueString()),
		AuthKey:            d.AuthKey.ValueString(),
		AuthKeyID:          optionalIntToAPI(d.AuthKeyID),
		Area:               d.Area.ValueString(),
		Cost:               optionalIntToAPI(d.Cost),
		CostDemoted:        optionalIntToAPI(d.CostDemoted),
		CarpDependOn:       api.SelectedMap(d.CarpDependOn.ValueString()),
		HelloInterval:      optionalIntToAPI(d.HelloInterval),
		DeadInterval:       optionalIntToAPI(d.DeadInterval),
		RetransmitInterval: optionalIntToAPI(d.RetransmitInterval),
		TransmitDelay:      optionalIntToAPI(d.TransmitDelay),
		Priority:           optionalIntToAPI(d.Priority),
		BFD:                optionalBoolToAPI(d.BFD),
		NetworkType:        api.SelectedMap(d.NetworkType.ValueString()),
		P2MPOptions:        api.SelectedMap(d.P2MPOptions.ValueString()),
	}
}

func ospfInterfaceFromAPI(d *quagga.OSPFInterface, id string) *ospfInterfaceModel {
	return &ospfInterfaceModel{
		Enabled:            optionalBoolFromAPI(d.Enabled),
		InterfaceName:      types.StringValue(d.InterfaceName.String()),
		AuthType:           types.StringValue(d.AuthType.String()),
		AuthKey:            types.StringValue(d.AuthKey),
		AuthKeyID:          optionalIntFromAPI(d.AuthKeyID),
		Area:               types.StringValue(d.Area),
		Cost:               optionalIntFromAPI(d.Cost),
		CostDemoted:        optionalIntFromAPI(d.CostDemoted),
		CarpDependOn:       types.StringValue(d.CarpDependOn.String()),
		HelloInterval:      optionalIntFromAPI(d.HelloInterval),
		DeadInterval:       optionalIntFromAPI(d.DeadInterval),
		RetransmitInterval: optionalIntFromAPI(d.RetransmitInterval),
		TransmitDelay:      optionalIntFromAPI(d.TransmitDelay),
		Priority:           optionalIntFromAPI(d.Priority),
		BFD:                optionalBoolFromAPI(d.BFD),
		NetworkType:        types.StringValue(d.NetworkType.String()),
		P2MPOptions:        types.StringValue(d.P2MPOptions.String()),
		ID:                 types.StringValue(id),
	}
}

var _ resource.Resource = &ospfInterfaceResource{}
var _ resource.ResourceWithConfigure = &ospfInterfaceResource{}
var _ resource.ResourceWithImportState = &ospfInterfaceResource{}

type ospfInterfaceResource struct{ client opnsense.Client }

func newOspfInterfaceResource() resource.Resource { return &ospfInterfaceResource{} }

func (r *ospfInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_interface"
}
func (r *ospfInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospfInterfaceResourceSchema()
}
func (r *ospfInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}

func (r *ospfInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospfInterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPFInterface(ctx, ospfInterfaceToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OPNsense OSPF Object", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFInterface(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPF Object Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfInterfaceFromAPI(remote, id))...)
}
func (r *ospfInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospfInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPFInterface(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OPNsense OSPF Object", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfInterfaceFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospfInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospfInterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPFInterface(ctx, data.ID.ValueString(), ospfInterfaceToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OPNsense OSPF Object", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPFInterface(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPF Object Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfInterfaceFromAPI(remote, data.ID.ValueString()))...)
}

func (r *ospfInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospfInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().DeleteOSPFInterface(ctx, data.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if !errors.As(err, &notFound) {
			resp.Diagnostics.AddError("Unable to Delete OPNsense OSPF Object", err.Error())
		}
	}
}
func (r *ospfInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &ospfInterfaceDataSource{}
var _ datasource.DataSourceWithConfigure = &ospfInterfaceDataSource{}

type ospfInterfaceDataSource struct{ client opnsense.Client }

func newOspfInterfaceDataSource() datasource.DataSource { return &ospfInterfaceDataSource{} }

func (d *ospfInterfaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf_interface"
}
func (d *ospfInterfaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospfInterfaceDataSourceSchema()
}
func (d *ospfInterfaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}

func (d *ospfInterfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospfInterfaceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPFInterface(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense OSPF Object", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospfInterfaceFromAPI(remote, data.ID.ValueString()))...)
}
