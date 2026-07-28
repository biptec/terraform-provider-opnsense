package interfaces

import (
	"github.com/biptec/opnsense-go/pkg/interfaces"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type greResourceModel struct {
	Device              types.String `tfsdk:"device"`
	LocalAddress        types.String `tfsdk:"local_address"`
	RemoteAddress       types.String `tfsdk:"remote_address"`
	TunnelLocalAddress  types.String `tfsdk:"tunnel_local_address"`
	TunnelRemoteAddress types.String `tfsdk:"tunnel_remote_address"`
	TunnelRemotePrefix  types.Int64  `tfsdk:"tunnel_remote_prefix"`
	Description         types.String `tfsdk:"description"`
	Id                  types.String `tfsdk:"id"`
}

func greResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages an OPNsense GRE tunnel interface.", Attributes: map[string]schema.Attribute{
		"device":                schema.StringAttribute{Computed: true, MarkdownDescription: "Automatically allocated GRE device name."},
		"local_address":         schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("wan"), MarkdownDescription: "Outer local endpoint: logical interface, CARP/IP Alias selector, or IP address."},
		"remote_address":        schema.StringAttribute{Optional: true, MarkdownDescription: "Outer remote IP address."},
		"tunnel_local_address":  schema.StringAttribute{Optional: true, MarkdownDescription: "Inner local tunnel address."},
		"tunnel_remote_address": schema.StringAttribute{Optional: true, MarkdownDescription: "Inner remote tunnel address."},
		"tunnel_remote_prefix":  schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(32), Validators: []validator.Int64{int64validator.Between(1, 128)}, MarkdownDescription: "Inner remote network prefix length."},
		"description":           schema.StringAttribute{Optional: true, MarkdownDescription: "Optional description."},
		"id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the GRE configuration.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func greDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Reads an OPNsense GRE tunnel interface.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Required: true}, "device": dschema.StringAttribute{Computed: true},
		"local_address": dschema.StringAttribute{Computed: true}, "remote_address": dschema.StringAttribute{Computed: true},
		"tunnel_local_address": dschema.StringAttribute{Computed: true}, "tunnel_remote_address": dschema.StringAttribute{Computed: true},
		"tunnel_remote_prefix": dschema.Int64Attribute{Computed: true}, "description": dschema.StringAttribute{Computed: true},
	}}
}

func convertGreSchemaToStruct(d *greResourceModel) (*interfaces.Gre, error) {
	if err := validateTunnelAddresses(d.RemoteAddress, d.TunnelLocalAddress, d.TunnelRemoteAddress, d.TunnelRemotePrefix); err != nil {
		return nil, err
	}
	return &interfaces.Gre{LocalAddress: d.LocalAddress.ValueString(), RemoteAddress: d.RemoteAddress.ValueString(), TunnelLocalAddress: d.TunnelLocalAddress.ValueString(), TunnelRemoteAddress: d.TunnelRemoteAddress.ValueString(), TunnelRemotePrefix: tools.Int64ToString(d.TunnelRemotePrefix.ValueInt64()), Description: d.Description.ValueString()}, nil
}

func convertGreStructToSchema(d *interfaces.Gre) (*greResourceModel, error) {
	local := d.LocalAddress
	if local == "" {
		if d.IPAddress != "" {
			local = d.IPAddress
		} else if d.Interface != "" {
			local = d.Interface
		} else {
			local = "wan"
		}
	}
	prefix := tools.StringToInt64(d.TunnelRemotePrefix)
	if prefix < 1 {
		prefix = 32
	}
	return &greResourceModel{Device: types.StringValue(d.Device), LocalAddress: types.StringValue(local), RemoteAddress: tools.StringOrNull(d.RemoteAddress), TunnelLocalAddress: tools.StringOrNull(d.TunnelLocalAddress), TunnelRemoteAddress: tools.StringOrNull(d.TunnelRemoteAddress), TunnelRemotePrefix: types.Int64Value(prefix), Description: tools.StringOrNull(d.Description)}, nil
}
