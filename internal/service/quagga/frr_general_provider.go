package quagga

import (
	"context"
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/opnsense-go/pkg/quagga"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type frrGeneralModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	Profile       types.String `tfsdk:"profile"`
	EnableCARP    types.Bool   `tfsdk:"enable_carp"`
	EnableSyslog  types.Bool   `tfsdk:"enable_syslog"`
	EnableSNMP    types.Bool   `tfsdk:"enable_snmp"`
	SyslogLevel   types.String `tfsdk:"syslog_level"`
	FirewallRules types.Bool   `tfsdk:"firewall_rules"`
	ID            types.String `tfsdk:"id"`
}

func frrGeneralResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages OPNsense FRR/OSPF singleton settings.", Attributes: map[string]schema.Attribute{
		"enabled":        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"profile":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("traditional")},
		"enable_carp":    schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"enable_syslog":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"enable_snmp":    schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"syslog_level":   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("notifications")},
		"firewall_rules": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"id":             schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func frrGeneralDataSourceSchema() dsschema.Schema {
	return dsschema.Schema{MarkdownDescription: "Reads OPNsense FRR/OSPF singleton settings.", Attributes: map[string]dsschema.Attribute{
		"enabled":        dsschema.BoolAttribute{Computed: true},
		"profile":        dsschema.StringAttribute{Computed: true},
		"enable_carp":    dsschema.BoolAttribute{Computed: true},
		"enable_syslog":  dsschema.BoolAttribute{Computed: true},
		"enable_snmp":    dsschema.BoolAttribute{Computed: true},
		"syslog_level":   dsschema.StringAttribute{Computed: true},
		"firewall_rules": dsschema.BoolAttribute{Computed: true},
		"id":             dsschema.StringAttribute{Computed: true},
	}}
}

func frrGeneralToAPI(ctx context.Context, d *frrGeneralModel) *quagga.FRRGeneralSettings {
	return &quagga.FRRGeneralSettings{
		Enabled:       optionalBoolToAPI(d.Enabled),
		Profile:       api.SelectedMap(d.Profile.ValueString()),
		EnableCARP:    optionalBoolToAPI(d.EnableCARP),
		EnableSyslog:  optionalBoolToAPI(d.EnableSyslog),
		EnableSNMP:    optionalBoolToAPI(d.EnableSNMP),
		SyslogLevel:   api.SelectedMap(d.SyslogLevel.ValueString()),
		FirewallRules: optionalBoolToAPI(d.FirewallRules),
	}
}

func frrGeneralFromAPI(d *quagga.FRRGeneralSettings) *frrGeneralModel {
	return &frrGeneralModel{
		Enabled:       optionalBoolFromAPI(d.Enabled),
		Profile:       types.StringValue(d.Profile.String()),
		EnableCARP:    optionalBoolFromAPI(d.EnableCARP),
		EnableSyslog:  optionalBoolFromAPI(d.EnableSyslog),
		EnableSNMP:    optionalBoolFromAPI(d.EnableSNMP),
		SyslogLevel:   types.StringValue(d.SyslogLevel.String()),
		FirewallRules: optionalBoolFromAPI(d.FirewallRules),
		ID:            types.StringValue("frr_general"),
	}
}

var _ resource.Resource = &frrGeneralResource{}
var _ resource.ResourceWithConfigure = &frrGeneralResource{}
var _ resource.ResourceWithImportState = &frrGeneralResource{}

type frrGeneralResource struct{ client opnsense.Client }

func newFrrGeneralResource() resource.Resource { return &frrGeneralResource{} }

func (r *frrGeneralResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_frr_general"
}
func (r *frrGeneralResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = frrGeneralResourceSchema()
}
func (r *frrGeneralResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureQuaggaResource(req, resp)
}

func (r *frrGeneralResource) apply(ctx context.Context, data *frrGeneralModel) error {
	setResult, err := r.client.Quagga().FRRGeneralSet(ctx, frrGeneralToAPI(ctx, data))
	if err != nil {
		return err
	}
	if err := validateRoutingSet(setResult); err != nil {
		return err
	}
	reconfigureResult, err := r.client.Quagga().FRRGeneralReconfigure(ctx)
	if err != nil {
		return err
	}
	return validateRoutingReconfigure(reconfigureResult)
}

func (r *frrGeneralResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data frrGeneralModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Apply OPNsense Routing Settings", err.Error())
		return
	}
	remote, err := r.client.Quagga().FRRGeneralGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Routing Settings Applied but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, frrGeneralFromAPI(&remote.General))...)
}
func (r *frrGeneralResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	remote, err := r.client.Quagga().FRRGeneralGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Routing Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, frrGeneralFromAPI(&remote.General))...)
}
func (r *frrGeneralResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data frrGeneralModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Apply OPNsense Routing Settings", err.Error())
		return
	}
	remote, err := r.client.Quagga().FRRGeneralGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Routing Settings Applied but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, frrGeneralFromAPI(&remote.General))...)
}
func (r *frrGeneralResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Routing singleton removed from Terraform state", "OPNsense singleton settings are not reset on destroy.")
}
func (r *frrGeneralResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "frr_general" {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected frr_general.")
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &frrGeneralDataSource{}
var _ datasource.DataSourceWithConfigure = &frrGeneralDataSource{}

type frrGeneralDataSource struct{ client opnsense.Client }

func newFrrGeneralDataSource() datasource.DataSource { return &frrGeneralDataSource{} }

func (d *frrGeneralDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_frr_general"
}
func (d *frrGeneralDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = frrGeneralDataSourceSchema()
}
func (d *frrGeneralDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureQuaggaDataSource(req, resp)
}

func (d *frrGeneralDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	remote, err := d.client.Quagga().FRRGeneralGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read OPNsense Routing Settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, frrGeneralFromAPI(&remote.General))...)
}
