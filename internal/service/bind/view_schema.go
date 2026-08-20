package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type viewResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	Sequence             types.Int64  `tfsdk:"sequence"`
	Name                 types.String `tfsdk:"name"`
	MatchAny             types.Bool   `tfsdk:"match_any"`
	MatchClientACLs      types.Set    `tfsdk:"match_client_acl_ids"`
	MatchDestinationACLs types.Set    `tfsdk:"match_destination_acl_ids"`
	Recursion            types.Bool   `tfsdk:"recursion"`
	AllowRecursion       types.Set    `tfsdk:"allow_recursion_acl_ids"`
	AllowQueryAny        types.Bool   `tfsdk:"allow_query_any"`
	AllowQuery           types.Set    `tfsdk:"allow_query_acl_ids"`
	AllowTransfer        types.Set    `tfsdk:"allow_transfer_acl_ids"`
	Forwarders           types.Set    `tfsdk:"forwarders"`
	DNSSECValidation     types.String `tfsdk:"dnssec_validation"`
}

func emptyStringSet() types.Set {
	return types.SetValueMust(types.StringType, []attr.Value{})
}

func viewResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages a BIND view. Views are evaluated in ascending sequence order; a match-any view must be last.", Attributes: map[string]schema.Attribute{
		"id":                        schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the view.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"enabled":                   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether the view is enabled."},
		"sequence":                  schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(100), MarkdownDescription: "Evaluation order from 1 to 9999.", Validators: []validator.Int64{int64validator.Between(1, 9999)}},
		"name":                      schema.StringAttribute{Required: true, MarkdownDescription: "Unique BIND view name.", Validators: []validator.String{stringvalidator.LengthBetween(1, 32)}},
		"match_any":                 schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Match every client. Only the final catch-all view should enable this."},
		"match_client_acl_ids":      schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, MarkdownDescription: "ACL UUIDs used by match-clients.", Validators: []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}},
		"match_destination_acl_ids": schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, MarkdownDescription: "ACL UUIDs used by match-destinations. Use separate LAN and WAN destination ACLs for split DNS.", Validators: []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}},
		"recursion":                 schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Enable recursive resolution in this view."},
		"allow_recursion_acl_ids":   schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, MarkdownDescription: "ACL UUIDs allowed to use recursion. If empty, match_client_acl_ids are used.", Validators: []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}},
		"allow_query_any":           schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Allow any client to query zones in this view."},
		"allow_query_acl_ids":       schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, MarkdownDescription: "ACL UUIDs allowed to query this view. If empty, match_client_acl_ids are used.", Validators: []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}},
		"allow_transfer_acl_ids":    schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, MarkdownDescription: "Default ACL UUIDs allowed to transfer zones in this view.", Validators: []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}},
		"forwarders":                schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, MarkdownDescription: "Optional upstream resolvers used by recursive queries."},
		"dnssec_validation":         schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("auto"), MarkdownDescription: "DNSSEC validation mode for recursive queries: auto or no.", Validators: []validator.String{stringvalidator.OneOf("auto", "no")}},
	}}
}

func viewDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a BIND view by UUID or semantic view name.", Attributes: map[string]dschema.Attribute{
		"id":                        dschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "UUID selector or resolved UUID of the view.", Validators: []validator.String{validators.IsUUIDv4()}},
		"enabled":                   dschema.BoolAttribute{Computed: true},
		"sequence":                  dschema.Int64Attribute{Computed: true},
		"name":                      dschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Unique semantic BIND view name. Lookup is trimmed and case-insensitive.", Validators: []validator.String{stringvalidator.LengthBetween(1, 32)}},
		"match_any":                 dschema.BoolAttribute{Computed: true},
		"match_client_acl_ids":      dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"match_destination_acl_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"recursion":                 dschema.BoolAttribute{Computed: true},
		"allow_recursion_acl_ids":   dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"allow_query_any":           dschema.BoolAttribute{Computed: true},
		"allow_query_acl_ids":       dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"allow_transfer_acl_ids":    dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"forwarders":                dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"dnssec_validation":         dschema.StringAttribute{Computed: true},
	}}
}

func viewModelToAPI(d *viewResourceModel) (*apibind.View, error) {
	return &apibind.View{
		Enabled: tools.BoolToString(d.Enabled.ValueBool()), Sequence: tools.Int64ToString(d.Sequence.ValueInt64()), Name: d.Name.ValueString(),
		MatchAny: tools.BoolToString(d.MatchAny.ValueBool()), MatchClients: api.SelectedMapList(tools.SetToStringSlice(d.MatchClientACLs)),
		MatchDestinations: api.SelectedMapList(tools.SetToStringSlice(d.MatchDestinationACLs)),
		Recursion:         tools.BoolToString(d.Recursion.ValueBool()), AllowRecursion: api.SelectedMapList(tools.SetToStringSlice(d.AllowRecursion)),
		AllowQueryAny: tools.BoolToString(d.AllowQueryAny.ValueBool()), AllowQuery: api.SelectedMapList(tools.SetToStringSlice(d.AllowQuery)),
		AllowTransfer: api.SelectedMapList(tools.SetToStringSlice(d.AllowTransfer)), Forwarders: api.SelectedMapList(tools.SetToStringSlice(d.Forwarders)),
		DNSSECValidation: api.SelectedMap(d.DNSSECValidation.ValueString()),
	}, nil
}

func viewAPIToModel(d *apibind.View) (*viewResourceModel, error) {
	return &viewResourceModel{
		Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Sequence: types.Int64Value(tools.StringToInt64(d.Sequence)), Name: types.StringValue(d.Name),
		MatchAny: types.BoolValue(tools.StringToBool(d.MatchAny)), MatchClientACLs: tools.StringSliceToSet([]string(d.MatchClients)),
		MatchDestinationACLs: tools.StringSliceToSet([]string(d.MatchDestinations)),
		Recursion:            types.BoolValue(tools.StringToBool(d.Recursion)), AllowRecursion: tools.StringSliceToSet([]string(d.AllowRecursion)),
		AllowQueryAny: types.BoolValue(tools.StringToBool(d.AllowQueryAny)), AllowQuery: tools.StringSliceToSet([]string(d.AllowQuery)),
		AllowTransfer: tools.StringSliceToSet([]string(d.AllowTransfer)), Forwarders: tools.StringSliceToSet([]string(d.Forwarders)),
		DNSSECValidation: types.StringValue(d.DNSSECValidation.String()),
	}, nil
}
