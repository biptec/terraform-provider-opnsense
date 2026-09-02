package system

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &pluginStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &pluginStatusDataSource{}

type pluginStatusDataSource struct {
	client opnsense.Client
}

type pluginStatusDataSourceModel struct {
	Name       types.String `tfsdk:"name"`
	Installed  types.Bool   `tfsdk:"installed"`
	Provided   types.Bool   `tfsdk:"provided"`
	Version    types.String `tfsdk:"version"`
	Locked     types.Bool   `tfsdk:"locked"`
	Repository types.String `tfsdk:"repository"`
	Origin     types.String `tfsdk:"origin"`
}

func newPluginStatusDataSource() datasource.DataSource {
	return &pluginStatusDataSource{}
}

func (d *pluginStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin_status"
}

func (d *pluginStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads cached local OPNsense package/plugin state through api-extensions without refreshing remote firmware repositories or mutating the appliance.",
		Attributes: map[string]schema.Attribute{
			"name":       schema.StringAttribute{Required: true, MarkdownDescription: "Exact OPNsense package/plugin name."},
			"installed":  schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the package is currently installed."},
			"provided":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the package is available in cached local package metadata."},
			"version":    schema.StringAttribute{Computed: true, MarkdownDescription: "Installed/provided package version reported by the local package database."},
			"locked":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the package is locked against firmware changes."},
			"repository": schema.StringAttribute{Computed: true, MarkdownDescription: "Repository recorded by the local package database."},
			"origin":     schema.StringAttribute{Computed: true, MarkdownDescription: "Package origin recorded by the local package database."},
		},
	}
}

func (d *pluginStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *pluginStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config pluginStatusDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	pkg, err := findLocalPlugin(ctx, d.client, name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Local OPNsense Plugin Status", err.Error())
		return
	}

	state := pluginStatusDataSourceModel{
		Name:       config.Name,
		Installed:  types.BoolValue(pkg.Installed),
		Provided:   types.BoolValue(pkg.Provided),
		Version:    types.StringValue(pkg.Version),
		Locked:     types.BoolValue(pkg.Locked),
		Repository: types.StringValue(pkg.Repository),
		Origin:     types.StringValue(pkg.Origin),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
