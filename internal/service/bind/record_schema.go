package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type recordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	DomainID types.String `tfsdk:"domain_id"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
}

var bindRecordTypes = []string{"A", "AAAA", "CAA", "CNAME", "DNAME", "DNSKEY", "DS", "HTTPS", "MX", "NAPTR", "NS", "PTR", "RP", "RRSIG", "SRV", "SSHFP", "SVCB", "TLSA", "TXT"}

func recordResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages a record in a BIND zone.", Attributes: map[string]schema.Attribute{
		"id":        schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the record.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"domain_id": schema.StringAttribute{Required: true, MarkdownDescription: "UUID of the BIND zone.", Validators: []validator.String{validators.IsUUIDv4()}},
		"enabled":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"name":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("@"), MarkdownDescription: "Relative record owner name. Use @ for the zone apex."},
		"type":      schema.StringAttribute{Required: true, MarkdownDescription: "DNS record type.", Validators: []validator.String{stringvalidator.OneOf(bindRecordTypes...)}},
		"value":     schema.StringAttribute{Required: true, MarkdownDescription: "Record value in BIND presentation format."},
	}}
}

func recordDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a BIND record.", Attributes: map[string]dschema.Attribute{
		"id":        dschema.StringAttribute{Required: true},
		"domain_id": dschema.StringAttribute{Computed: true},
		"enabled":   dschema.BoolAttribute{Computed: true},
		"name":      dschema.StringAttribute{Computed: true},
		"type":      dschema.StringAttribute{Computed: true},
		"value":     dschema.StringAttribute{Computed: true},
	}}
}

func recordModelToAPI(d *recordResourceModel) (*apibind.Record, error) {
	return &apibind.Record{Domain: api.SelectedMap(d.DomainID.ValueString()), Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Type: api.SelectedMap(d.Type.ValueString()), Value: d.Value.ValueString()}, nil
}

func recordAPIToModel(d *apibind.Record) (*recordResourceModel, error) {
	return &recordResourceModel{DomainID: types.StringValue(d.Domain.String()), Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), Type: types.StringValue(d.Type.String()), Value: types.StringValue(d.Value)}, nil
}
