package interfaces

import (
	apiinterfaces "github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type loopbackResourceModel struct {
	DeviceID    types.Int64  `tfsdk:"device_id"`
	Description types.String `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
}

func loopbackResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages an OPNsense loopback interface.", Attributes: map[string]schema.Attribute{
		"device_id":   schema.Int64Attribute{Computed: true, MarkdownDescription: "Automatically allocated loopback device number."},
		"description": schema.StringAttribute{Required: true, MarkdownDescription: "Loopback description."},
		"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the loopback configuration.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func loopbackDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense loopback interface.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "device_id": dschema.Int64Attribute{Computed: true}, "description": dschema.StringAttribute{Computed: true},
	}}
}

func convertLoopbackSchemaToStruct(d *loopbackResourceModel) (*apiinterfaces.Loopback, error) {
	return &apiinterfaces.Loopback{Description: d.Description.ValueString()}, nil
}
func convertLoopbackStructToSchema(d *apiinterfaces.Loopback) (*loopbackResourceModel, error) {
	return &loopbackResourceModel{DeviceID: tools.StringToInt64Null(d.DeviceID), Description: types.StringValue(d.Description)}, nil
}
