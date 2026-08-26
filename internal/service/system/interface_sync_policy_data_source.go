package system

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &interfaceSyncPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &interfaceSyncPolicyDataSource{}

type interfaceSyncPolicyDataSource struct {
	client opnsense.Client
}

type interfaceSyncPolicyDataSourceModel struct {
	PolicyID    types.String `tfsdk:"policy_id"`
	Description types.String `tfsdk:"description"`
	Synchronize types.Bool   `tfsdk:"synchronize"`
}

func newInterfaceSyncPolicyDataSource() datasource.DataSource {
	return &interfaceSyncPolicyDataSource{}
}

func (d *interfaceSyncPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface_sync_policy"
}

func (d *interfaceSyncPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an interface synchronization policy by stable policy_id. This makes the router-stored policy the source of truth for consumer synchronization behavior.",
		Attributes: map[string]schema.Attribute{
			"policy_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Stable policy identifier.",
			},
			"description": schema.StringAttribute{Computed: true},
			"synchronize": schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *interfaceSyncPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *interfaceSyncPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config interfaceSyncPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := d.client.ApiExtensions().SearchInterfaceSyncPolicy(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Search Interface Sync Policies", err.Error())
		return
	}

	var matches []struct {
		Description string
		Synchronize bool
	}
	for _, item := range result.Rows {
		if item.ID != config.PolicyID.ValueString() {
			continue
		}
		matches = append(matches, struct {
			Description string
			Synchronize bool
		}{
			Description: item.Description,
			Synchronize: tools.StringToBool(string(item.Synchronize)),
		})
	}
	if len(matches) != 1 {
		resp.Diagnostics.AddError("Interface Sync Policy Lookup Failed", fmt.Sprintf("Expected exactly one policy with policy_id %q, found %d.", config.PolicyID.ValueString(), len(matches)))
		return
	}
	state := interfaceSyncPolicyDataSourceModel{
		PolicyID:    config.PolicyID,
		Description: types.StringValue(matches[0].Description),
		Synchronize: types.BoolValue(matches[0].Synchronize),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
