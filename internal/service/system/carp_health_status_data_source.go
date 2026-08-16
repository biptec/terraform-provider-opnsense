package system

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type carpHealthStatusCheckModel struct {
	UUID      types.String `tfsdk:"uuid"`
	Name      types.String `tfsdk:"name"`
	Interface types.String `tfsdk:"interface"`
	Device    types.String `tfsdk:"device"`
	Target    types.String `tfsdk:"target"`
	Healthy   types.Bool   `tfsdk:"healthy"`
	Failures  types.Int64  `tfsdk:"failures"`
	Successes types.Int64  `tfsdk:"successes"`
}

type carpHealthStatusModel struct {
	ID              types.String  `tfsdk:"id"`
	Status          types.String  `tfsdk:"status"`
	Enabled         types.Bool    `tfsdk:"enabled"`
	Ready           types.Bool    `tfsdk:"ready"`
	Healthy         types.Bool    `tfsdk:"healthy"`
	Running         types.Bool    `tfsdk:"running"`
	Timestamp       types.Float64 `tfsdk:"timestamp"`
	ConfigSignature types.String  `tfsdk:"config_signature"`
	Checks          types.List    `tfsdk:"checks"`
}

var carpHealthStatusCheckTypes = map[string]attr.Type{
	"uuid": types.StringType, "name": types.StringType, "interface": types.StringType,
	"device": types.StringType, "target": types.StringType, "healthy": types.BoolType,
	"failures": types.Int64Type, "successes": types.Int64Type,
}

var _ datasource.DataSource = &carpHealthStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &carpHealthStatusDataSource{}

type carpHealthStatusDataSource struct{ client opnsense.Client }

func newCarpHealthStatusDataSource() datasource.DataSource { return &carpHealthStatusDataSource{} }
func (d *carpHealthStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carp_health_status"
}
func (d *carpHealthStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Reads runtime CARP health-monitor status from `os-api-extensions`.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true}, "status": schema.StringAttribute{Computed: true},
		"enabled": schema.BoolAttribute{Computed: true}, "ready": schema.BoolAttribute{Computed: true},
		"healthy": schema.BoolAttribute{Computed: true}, "running": schema.BoolAttribute{Computed: true},
		"timestamp": schema.Float64Attribute{Computed: true}, "config_signature": schema.StringAttribute{Computed: true},
		"checks": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{Computed: true}, "name": schema.StringAttribute{Computed: true},
			"interface": schema.StringAttribute{Computed: true}, "device": schema.StringAttribute{Computed: true},
			"target": schema.StringAttribute{Computed: true}, "healthy": schema.BoolAttribute{Computed: true},
			"failures": schema.Int64Attribute{Computed: true}, "successes": schema.Int64Attribute{Computed: true},
		}}},
	}}
}
func (d *carpHealthStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(client)
}
func (d *carpHealthStatusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	remote, err := d.client.ApiExtensions().CarpHealthStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read CARP Health Status", err.Error())
		return
	}
	checks := make([]carpHealthStatusCheckModel, 0, len(remote.Checks))
	for _, item := range remote.Checks {
		checks = append(checks, carpHealthStatusCheckModel{UUID: types.StringValue(item.UUID), Name: types.StringValue(item.Name), Interface: types.StringValue(item.Interface), Device: types.StringValue(item.Device), Target: types.StringValue(item.Target), Healthy: types.BoolValue(item.Healthy), Failures: types.Int64Value(int64(item.Failures)), Successes: types.Int64Value(int64(item.Successes))})
	}
	list, diagnostics := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: carpHealthStatusCheckTypes}, checks)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &carpHealthStatusModel{ID: types.StringValue("carp_health"), Status: types.StringValue(remote.Status), Enabled: types.BoolValue(remote.Enabled), Ready: types.BoolValue(remote.Ready), Healthy: types.BoolValue(remote.Healthy), Running: types.BoolValue(remote.Running), Timestamp: types.Float64Value(remote.Timestamp), ConfigSignature: types.StringValue(remote.ConfigSignature), Checks: list}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
