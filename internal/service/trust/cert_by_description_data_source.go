package trust

import (
	"context"
	"fmt"
	"strings"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &certByDescriptionDataSource{}
var _ datasource.DataSourceWithConfigure = &certByDescriptionDataSource{}

type certByDescriptionDataSource struct {
	client opnsense.Client
}

type certByDescriptionDataSourceModel struct {
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	RefID       types.String `tfsdk:"ref_id"`
}

func newCertByDescriptionDataSource() datasource.DataSource {
	return &certByDescriptionDataSource{}
}

func (d *certByDescriptionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trust_cert_by_description"
}

func (d *certByDescriptionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up exactly one existing OPNsense Trust certificate by its stable description and returns only local metadata needed to reference it. Certificate/key material is intentionally not exposed in Terraform state.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Exact certificate description. Leading/trailing whitespace is ignored for the lookup input, but matching is case-sensitive.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Local OPNsense certificate UUID.",
			},
			"ref_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Local OPNsense certificate refid used by HAProxy and other subsystems.",
			},
		},
	}
}

func (d *certByDescriptionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *certByDescriptionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config certByDescriptionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	description := strings.TrimSpace(config.Description.ValueString())
	if description == "" {
		resp.Diagnostics.AddError("Certificate Description Lookup Failed", "description must not be empty.")
		return
	}

	result, err := d.client.Trust().SearchCert(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Search Trust Certificates", err.Error())
		return
	}

	matches := make([]struct {
		ID    string
		RefID string
	}, 0, 1)
	for _, item := range result.Rows {
		if item.Description != description {
			continue
		}
		id := strings.TrimSpace(item.UUID)
		refID := strings.TrimSpace(item.RefId)
		if id == "" || refID == "" {
			resp.Diagnostics.AddError(
				"Certificate Description Lookup Failed",
				fmt.Sprintf("Certificate %q matched by description but is missing its local UUID/refid.", description),
			)
			return
		}
		matches = append(matches, struct {
			ID    string
			RefID string
		}{ID: id, RefID: refID})
	}

	if len(matches) != 1 {
		resp.Diagnostics.AddError(
			"Certificate Description Lookup Failed",
			fmt.Sprintf("Expected exactly one certificate with description %q, found %d.", description, len(matches)),
		)
		return
	}

	state := certByDescriptionDataSourceModel{
		Description: types.StringValue(description),
		ID:          types.StringValue(matches[0].ID),
		RefID:       types.StringValue(matches[0].RefID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
