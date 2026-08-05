package bind

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &dnssecStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &dnssecStatusDataSource{}

type dnssecStatusDataSource struct{ client opnsense.Client }

type dnssecStatusDataSourceModel struct {
	DomainID      types.String `tfsdk:"domain_id"`
	Zone          types.String `tfsdk:"zone"`
	View          types.String `tfsdk:"view"`
	Secure        types.Bool   `tfsdk:"secure"`
	InlineSigning types.Bool   `tfsdk:"inline_signing"`
	DSRecords     types.Set    `tfsdk:"ds_records"`
	Keys          types.List   `tfsdk:"keys"`
	RNDCStatus    types.Map    `tfsdk:"rndc_status"`
	Error         types.String `tfsdk:"error"`
}

type dnssecKeyModel struct {
	File      types.String `tfsdk:"file"`
	KeyTag    types.String `tfsdk:"key_tag"`
	Algorithm types.String `tfsdk:"algorithm"`
	Flags     types.String `tfsdk:"flags"`
	Role      types.String `tfsdk:"role"`
}

var dnssecKeyAttrTypes = map[string]attr.Type{
	"file":      types.StringType,
	"key_tag":   types.StringType,
	"algorithm": types.StringType,
	"flags":     types.StringType,
	"role":      types.StringType,
}

func newDNSSECStatusDataSource() datasource.DataSource { return &dnssecStatusDataSource{} }
func (d *dnssecStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_dnssec_status"
}
func (d *dnssecStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads live DNSSEC signing status and registrar-ready SHA-256 DS records for a BIND primary zone.",
		Attributes: map[string]schema.Attribute{
			"domain_id":      schema.StringAttribute{Required: true, MarkdownDescription: "UUID of the primary BIND zone.", Validators: []validator.String{validators.IsUUIDv4()}},
			"zone":           schema.StringAttribute{Required: true, MarkdownDescription: "Zone name used to validate the requested UUID."},
			"view":           schema.StringAttribute{Computed: true, MarkdownDescription: "BIND view containing the zone."},
			"secure":         schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether BIND reports the zone as secure."},
			"inline_signing": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether inline signing is active."},
			"ds_records":     schema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "SHA-256 DS records suitable for publication at the registrar."},
			"keys": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"file":      schema.StringAttribute{Computed: true},
				"key_tag":   schema.StringAttribute{Computed: true},
				"algorithm": schema.StringAttribute{Computed: true},
				"flags":     schema.StringAttribute{Computed: true},
				"role":      schema.StringAttribute{Computed: true},
			}}},
			"rndc_status": schema.MapAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Parsed `rndc zonestatus` fields."},
			"error":       schema.StringAttribute{Computed: true, MarkdownDescription: "Backend error, if any."},
		},
	}
}
func (d *dnssecStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *dnssecStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dnssecStatusDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := d.client.Bind().DNSSECStatus(ctx, config.Zone.ValueString(), config.DomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read BIND DNSSEC Status", err.Error())
		return
	}
	keys := make([]dnssecKeyModel, 0, len(status.Keys))
	for _, key := range status.Keys {
		keys = append(keys, dnssecKeyModel{File: types.StringValue(key.File), KeyTag: types.StringValue(key.KeyTag), Algorithm: types.StringValue(key.Algorithm), Flags: types.StringValue(key.Flags), Role: types.StringValue(key.Role)})
	}
	keyList, diagnostics := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: dnssecKeyAttrTypes}, keys)
	resp.Diagnostics.Append(diagnostics...)
	statusMap, diagnostics := types.MapValueFrom(ctx, types.StringType, status.RNDCStatus)
	resp.Diagnostics.Append(diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := dnssecStatusDataSourceModel{
		DomainID: config.DomainID, Zone: types.StringValue(status.Zone), View: types.StringValue(status.View),
		Secure: types.BoolValue(status.Secure), InlineSigning: types.BoolValue(status.InlineSigning),
		DSRecords: tools.StringSliceToSet(status.DSRecords), Keys: keyList, RNDCStatus: statusMap,
		Error: types.StringValue(status.Error),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
