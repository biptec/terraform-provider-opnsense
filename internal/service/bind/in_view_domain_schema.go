package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type inViewDomainResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ViewID       types.String `tfsdk:"view_id"`
	DomainName   types.String `tfsdk:"domain_name"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	SourceViewID types.String `tfsdk:"source_view_id"`
}

func inViewDomainResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Shares an existing primary or secondary BIND zone from an earlier view using BIND in-view.", Attributes: map[string]schema.Attribute{
		"id":             schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"view_id":        schema.StringAttribute{Required: true, Validators: []validator.String{validators.IsUUIDv4()}, MarkdownDescription: "Target view UUID. The target view must be ordered after source_view_id."},
		"domain_name":    schema.StringAttribute{Required: true, MarkdownDescription: "Zone name shared from the source view."},
		"enabled":        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
		"source_view_id": schema.StringAttribute{Required: true, Validators: []validator.String{validators.IsUUIDv4()}, MarkdownDescription: "Source view UUID containing the primary or secondary zone."},
	}}
}

func inViewDomainDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads a BIND in-view zone.", Attributes: map[string]dschema.Attribute{
		"id":             dschema.StringAttribute{Required: true},
		"view_id":        dschema.StringAttribute{Computed: true},
		"domain_name":    dschema.StringAttribute{Computed: true},
		"enabled":        dschema.BoolAttribute{Computed: true},
		"source_view_id": dschema.StringAttribute{Computed: true},
	}}
}

func inViewDomainModelToAPI(d *inViewDomainResourceModel) (*apibind.InViewDomain, error) {
	return &apibind.InViewDomain{
		View: api.SelectedMap(d.ViewID.ValueString()), DomainName: d.DomainName.ValueString(),
		Enabled: tools.BoolToString(d.Enabled.ValueBool()), SourceView: api.SelectedMap(d.SourceViewID.ValueString()),
	}, nil
}

func inViewDomainAPIToModel(d *apibind.InViewDomain) (*inViewDomainResourceModel, error) {
	return &inViewDomainResourceModel{
		ViewID: types.StringValue(d.View.String()), DomainName: types.StringValue(d.DomainName),
		Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), SourceViewID: types.StringValue(d.SourceView.String()),
	}, nil
}
