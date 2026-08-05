package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type forwardDomainResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ViewID         types.String `tfsdk:"view_id"`
	DomainName     types.String `tfsdk:"domain_name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	ForwardServers types.Set    `tfsdk:"forward_servers"`
	AllowQueryACLs types.Set    `tfsdk:"allow_query_acl_ids"`
}

func forwardDomainResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages a forward-only BIND zone inside a selected view.", Attributes: map[string]schema.Attribute{
		"id":                  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"view_id":             schema.StringAttribute{Required: true, Validators: []validator.String{validators.IsUUIDv4()}},
		"domain_name":         schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(1, 255)}},
		"enabled":             schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"forward_servers":     schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.ValueStringsAre(validators.IPAddress())}},
		"allow_query_acl_ids": schema.SetAttribute{Optional: true, Computed: true, Default: setdefault.StaticValue(emptyStringSet()), ElementType: types.StringType, Validators: []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}},
	}}
}
func forwardDomainDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a forward BIND zone.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "view_id": dschema.StringAttribute{Computed: true}, "domain_name": dschema.StringAttribute{Computed: true}, "enabled": dschema.BoolAttribute{Computed: true},
		"forward_servers": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "allow_query_acl_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
	}}
}
func forwardDomainModelToAPI(d *forwardDomainResourceModel) (*apibind.ForwardDomain, error) {
	return &apibind.ForwardDomain{View: api.SelectedMap(d.ViewID.ValueString()), DomainName: d.DomainName.ValueString(), Enabled: tools.BoolToString(d.Enabled.ValueBool()), ForwardServers: api.SelectedMapList(tools.SetToStringSlice(d.ForwardServers)), AllowQuery: api.SelectedMapList(tools.SetToStringSlice(d.AllowQueryACLs))}, nil
}
func forwardDomainAPIToModel(d *apibind.ForwardDomain) (*forwardDomainResourceModel, error) {
	return &forwardDomainResourceModel{ViewID: types.StringValue(d.View.String()), DomainName: types.StringValue(d.DomainName), Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), ForwardServers: tools.StringSliceToSet([]string(d.ForwardServers)), AllowQueryACLs: tools.StringSliceToSet([]string(d.AllowQuery))}, nil
}
