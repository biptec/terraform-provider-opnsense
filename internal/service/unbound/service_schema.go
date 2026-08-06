package unbound

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type serviceResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func serviceResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages only the persistent Unbound service enable flag. The resource uses read/modify/write and preserves every other Unbound setting. Do not manage `general.enabled` through `opnsense_unbound_settings` at the same time.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `unbound_service`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether Unbound is persistently enabled.",
			},
		},
	}
}
func serviceDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads the persistent Unbound service enable flag.",
		Attributes: map[string]dschema.Attribute{
			"id":      dschema.StringAttribute{Computed: true},
			"enabled": dschema.BoolAttribute{Computed: true},
		},
	}
}

func serviceAPIToModel(enabled string) *serviceResourceModel {
	return &serviceResourceModel{
		ID:      types.StringValue("unbound_service"),
		Enabled: types.BoolValue(enabled == "1"),
	}
}
