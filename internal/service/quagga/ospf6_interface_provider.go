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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ospf6InterfaceModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	InterfaceName      types.String `tfsdk:"interface_name"`
	Area               types.String `tfsdk:"area"`
	Passive            types.Bool   `tfsdk:"passive"`
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
	ID                 types.String `tfsdk:"id"`
}

func ospf6InterfaceResourceSchema() rschema.Schema {
	return rschema.Schema{MarkdownDescription: "Manages one OPNsense OSPF object.", Attributes: map[string]rschema.Attribute{
		"enabled":             rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"interface_name":      rschema.StringAttribute{Optional: true, Computed: true},
		"area":                rschema.StringAttribute{Required: true},
		"passive":             rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"cost":                rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"cost_demoted":        rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(1, 65535)}},
		"carp_depend_on":      rschema.StringAttribute{Optional: true, Computed: true},
		"hello_interval":      rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"dead_interval":       rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"retransmit_interval": rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"transmit_delay":      rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"priority":            rschema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(0, 4294967295)}},
		"bfd":                 rschema.BoolAttribute{Optional: true, Computed: true},
		"network_type":        rschema.StringAttribute{Optional: true, Computed: true},
		"id":                  rschema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func ospf6InterfaceDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads one OPNsense OSPF object by UUID.", Attributes: map[string]dsschema.Attribute{
		"enabled":             dsschema.BoolAttribute{Computed: true},
		"interface_name":      dsschema.StringAttribute{Computed: true},
		"area":                dsschema.StringAttribute{Computed: true},
		"passive":             dsschema.BoolAttribute{Computed: true},
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
		"id":                  dsschema.StringAttribute{Required: true},
	}}
}

func ospf6InterfaceToAPI(d *ospf6InterfaceModel) *quagga.OSPF6Interface {
	return &quagga.OSPF6Interface{
		Enabled:            optionalBoolToAPI(d.Enabled),
		InterfaceName:      api.SelectedMap(d.InterfaceName.ValueString()),
		Area:               d.Area.ValueString(),
		Passive:            optionalBoolToAPI(d.Passive),
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
	}
}

func ospf6InterfaceFromAPI(d *quagga.OSPF6Interface, id string) *ospf6InterfaceModel {
	return &ospf6InterfaceModel{
		Enabled:            optionalBoolFromAPI(d.Enabled),
		InterfaceName:      types.StringValue(d.InterfaceName.String()),
		Area:               types.StringValue(d.Area),
		Passive:            optionalBoolFromAPI(d.Passive),
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
		ID:                 types.StringValue(id),
	}
}

var _ resource.Resource = &ospf6InterfaceResource{}
var _ resource.ResourceWithConfigure = &ospf6InterfaceResource{}
var _ resource.ResourceWithImportState = &ospf6InterfaceResource{}

type ospf6InterfaceResource struct{ client opnsense.Client }

func newOspf6InterfaceResource() resource.Resource { return &ospf6InterfaceResource{} }

func (r *ospf6InterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_interface"
}
func (r *ospf6InterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ospf6InterfaceResourceSchema()
}
func (r *ospf6InterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}

func (r *ospf6InterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ospf6InterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Quagga().AddOSPF6Interface(ctx, ospf6InterfaceToAPI(&data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create OPNsense OSPF Object", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6Interface(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("OSPF Object Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6InterfaceFromAPI(remote, id))...)
}
func (r *ospf6InterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ospf6InterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.Quagga().GetOSPF6Interface(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read OPNsense OSPF Object", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6InterfaceFromAPI(remote, data.ID.ValueString()))...)
}
func (r *ospf6InterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ospf6InterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().UpdateOSPF6Interface(ctx, data.ID.ValueString(), ospf6InterfaceToAPI(&data)); err != nil {
		resp.Diagnostics.AddError("Unable to Update OPNsense OSPF Object", err.Error())
		return
	}
	remote, err := r.client.Quagga().GetOSPF6Interface(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("OSPF Object Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6InterfaceFromAPI(remote, data.ID.ValueString()))...)
}

func (r *ospf6InterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ospf6InterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Quagga().DeleteOSPF6Interface(ctx, data.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if !errors.As(err, &notFound) {
			resp.Diagnostics.AddError("Unable to Delete OPNsense OSPF Object", err.Error())
		}
	}
}
func (r *ospf6InterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &ospf6InterfaceDataSource{}
var _ datasource.DataSourceWithConfigure = &ospf6InterfaceDataSource{}

type ospf6InterfaceDataSource struct{ client opnsense.Client }

func newOspf6InterfaceDataSource() datasource.DataSource { return &ospf6InterfaceDataSource{} }

func (d *ospf6InterfaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_ospf6_interface"
}
func (d *ospf6InterfaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ospf6InterfaceDataSourceSchema()
}
func (d *ospf6InterfaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}

func (d *ospf6InterfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ospf6InterfaceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := d.client.Quagga().GetOSPF6Interface(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense OSPF Object", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, ospf6InterfaceFromAPI(remote, data.ID.ValueString()))...)
}
