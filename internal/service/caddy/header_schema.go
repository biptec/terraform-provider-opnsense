package caddy

import (
	"regexp"

	"github.com/biptec/opnsense-go/pkg/api"
	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var headerTextPattern = regexp.MustCompile(`^[^"]*$`)

type headerResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Direction   types.String `tfsdk:"direction"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Replace     types.String `tfsdk:"replace"`
	Description types.String `tfsdk:"description"`
}

func headerResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a Caddy request or response header operation.",
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the header operation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"direction": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("header_up"),
				MarkdownDescription: "Apply the operation to an upstream request (`header_up`) or downstream response (`header_down`).",
				Validators: []validator.String{
					stringvalidator.OneOf("header_up", "header_down"),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "HTTP header name or Caddy header operation expression.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
					stringvalidator.RegexMatches(headerTextPattern, "must not contain quotation marks"),
				},
			},
			"value": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Header value. Caddy placeholders such as `{host}` are accepted.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
					stringvalidator.RegexMatches(headerTextPattern, "must not contain quotation marks"),
				},
			},
			"replace": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Optional replacement expression.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
					stringvalidator.RegexMatches(headerTextPattern, "must not contain quotation marks"),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Optional description. Defaults to `\"\"`.",
			},
		},
	}
}

func headerDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads a Caddy header operation by UUID.",
		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{Required: true, MarkdownDescription: "UUID of the header operation."},
			"direction":   dschema.StringAttribute{Computed: true},
			"name":        dschema.StringAttribute{Computed: true},
			"value":       dschema.StringAttribute{Computed: true},
			"replace":     dschema.StringAttribute{Computed: true},
			"description": dschema.StringAttribute{Computed: true},
		},
	}
}

func convertHeaderSchemaToStruct(d *headerResourceModel) (*apicaddy.Header, error) {
	return &apicaddy.Header{
		Direction:   api.SelectedMap(d.Direction.ValueString()),
		Name:        d.Name.ValueString(),
		Value:       d.Value.ValueString(),
		Replace:     d.Replace.ValueString(),
		Description: d.Description.ValueString(),
	}, nil
}

func convertHeaderStructToSchema(d *apicaddy.Header) (*headerResourceModel, error) {
	return &headerResourceModel{
		Direction:   types.StringValue(d.Direction.String()),
		Name:        types.StringValue(d.Name),
		Value:       types.StringValue(d.Value),
		Replace:     types.StringValue(d.Replace),
		Description: types.StringValue(d.Description),
	}, nil
}
