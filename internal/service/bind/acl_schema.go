package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type aclResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Name     types.String `tfsdk:"name"`
	Networks types.Set    `tfsdk:"networks"`
}

func aclResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages a BIND ACL used by views, query policy, recursion and zone transfers.", Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the ACL.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"enabled":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether the ACL is enabled."},
		"name":     schema.StringAttribute{Required: true, MarkdownDescription: "Unique ACL name."},
		"networks": schema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "IPv4 or IPv6 CIDR networks in this ACL.", Validators: []validator.Set{setvalidator.ValueStringsAre(validators.CIDR())}},
	}}
}

func aclDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a BIND ACL.", Attributes: map[string]dschema.Attribute{
		"id":       dschema.StringAttribute{Required: true, MarkdownDescription: "UUID of the ACL."},
		"enabled":  dschema.BoolAttribute{Computed: true},
		"name":     dschema.StringAttribute{Computed: true},
		"networks": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
	}}
}

func aclModelToAPI(d *aclResourceModel) (*apibind.Acl, error) {
	return &apibind.Acl{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Networks: api.SelectedMapList(tools.SetToStringSlice(d.Networks))}, nil
}

func aclAPIToModel(d *apibind.Acl) (*aclResourceModel, error) {
	return &aclResourceModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), Name: types.StringValue(d.Name), Networks: tools.StringSliceToSet([]string(d.Networks))}, nil
}
