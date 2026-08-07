package bind

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apibind "github.com/biptec/opnsense-go/pkg/bind"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/biptec/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type settingsResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	DisableIPv6        types.Bool   `tfsdk:"disable_ipv6"`
	EnableRPZ          types.Bool   `tfsdk:"enable_rpz"`
	ListenIPv4         types.Set    `tfsdk:"listen_ipv4"`
	ListenIPv6         types.Set    `tfsdk:"listen_ipv6"`
	QuerySource        types.String `tfsdk:"query_source"`
	QuerySourceIPv6    types.String `tfsdk:"query_source_ipv6"`
	TransferSource     types.String `tfsdk:"transfer_source"`
	TransferSourceIPv6 types.String `tfsdk:"transfer_source_ipv6"`
	Port               types.Int64  `tfsdk:"port"`
	Forwarders         types.Set    `tfsdk:"forwarders"`
	FilterAAAAIPv4     types.Bool   `tfsdk:"filter_aaaa_ipv4"`
	FilterAAAAIPv6     types.Bool   `tfsdk:"filter_aaaa_ipv6"`
	FilterAAAAACL      types.Set    `tfsdk:"filter_aaaa_acl"`
	LogSize            types.Int64  `tfsdk:"log_size_mb"`
	LogLevel           types.String `tfsdk:"log_level"`
	MaxCacheSize       types.Int64  `tfsdk:"max_cache_size_percent"`
	RecursionACLIDs    types.Set    `tfsdk:"legacy_recursion_acl_ids"`
	AllowTransferACLs  types.Set    `tfsdk:"legacy_allow_transfer_acl_ids"`
	AllowQueryACLs     types.Set    `tfsdk:"legacy_allow_query_acl_ids"`
	DNSSECValidation   types.String `tfsdk:"dnssec_validation"`
	HideHostname       types.Bool   `tfsdk:"hide_hostname"`
	HideVersion        types.Bool   `tfsdk:"hide_version"`
	DisablePrefetch    types.Bool   `tfsdk:"disable_prefetch"`
	EnableRateLimiting types.Bool   `tfsdk:"enable_rate_limiting"`
	RateLimitCount     types.Int64  `tfsdk:"rate_limit_count"`
	RateLimitExcept    types.Set    `tfsdk:"rate_limit_exceptions"`
}

func settingsAttributesResource() map[string]schema.Attribute {
	uuidSet := []validator.Set{setvalidator.ValueStringsAre(validators.IsUUIDv4())}
	ipSet := []validator.Set{setvalidator.ValueStringsAre(validators.IPAddress())}
	listenerSet := []validator.Set{
		setvalidator.SizeAtLeast(1),
		setvalidator.ValueStringsAre(validators.IPAddress()),
	}
	return map[string]schema.Attribute{
		"id":                            schema.StringAttribute{Computed: true, MarkdownDescription: "Always `bind_settings`.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"enabled":                       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether BIND is enabled."},
		"disable_ipv6":                  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Disable IPv6 support in BIND."},
		"enable_rpz":                    schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Enable response policy zones."},
		"listen_ipv4":                   schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Validators: listenerSet, MarkdownDescription: "Non-empty IPv4 listener set required by os-bind. Use 0.0.0.0 for any IPv4 address."},
		"listen_ipv6":                   schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Validators: listenerSet, MarkdownDescription: "Non-empty IPv6 listener set required by os-bind. Use ::1 when BIND runs with IPv6 disabled."},
		"query_source":                  schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{validators.IPAddress()}, MarkdownDescription: "Optional IPv4 source address for recursive queries."},
		"query_source_ipv6":             schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{validators.IPAddress()}, MarkdownDescription: "Optional IPv6 source address for recursive queries."},
		"transfer_source":               schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{validators.IPAddress()}, MarkdownDescription: "Optional IPv4 source address for zone transfers."},
		"transfer_source_ipv6":          schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{validators.IPAddress()}, MarkdownDescription: "Optional IPv6 source address for zone transfers."},
		"port":                          schema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(1, 65535)}, MarkdownDescription: "DNS listen port."},
		"forwarders":                    schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Validators: ipSet, MarkdownDescription: "Legacy global forwarding servers. Per-view forwarders take precedence when views are enabled."},
		"filter_aaaa_ipv4":              schema.BoolAttribute{Optional: true, Computed: true},
		"filter_aaaa_ipv6":              schema.BoolAttribute{Optional: true, Computed: true},
		"filter_aaaa_acl":               schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, MarkdownDescription: "Addresses or networks subject to AAAA filtering."},
		"log_size_mb":                   schema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "Maximum size of each BIND log file in MiB."},
		"log_level":                     schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{stringvalidator.OneOf("critical", "error", "warning", "notice", "info", "debug", "dynamic")}},
		"max_cache_size_percent":        schema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.Between(1, 90)}, MarkdownDescription: "Maximum BIND cache size as a percentage of memory."},
		"legacy_recursion_acl_ids":      schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Validators: uuidSet, MarkdownDescription: "Legacy global recursion ACLs used only when no views are enabled."},
		"legacy_allow_transfer_acl_ids": schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Validators: uuidSet, MarkdownDescription: "Legacy global transfer ACLs used only when no views are enabled."},
		"legacy_allow_query_acl_ids":    schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Validators: uuidSet, MarkdownDescription: "Legacy global query ACLs used only when no views are enabled."},
		"dnssec_validation":             schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{stringvalidator.OneOf("auto", "no")}, MarkdownDescription: "Legacy global DNSSEC validation mode used only when no views are enabled."},
		"hide_hostname":                 schema.BoolAttribute{Optional: true, Computed: true},
		"hide_version":                  schema.BoolAttribute{Optional: true, Computed: true},
		"disable_prefetch":              schema.BoolAttribute{Optional: true, Computed: true},
		"enable_rate_limiting":          schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Enable BIND response rate limiting."},
		"rate_limit_count":              schema.Int64Attribute{Optional: true, Computed: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "Responses per second allowed before rate limiting."},
		"rate_limit_exceptions":         schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, MarkdownDescription: "Addresses or networks exempt from response rate limiting."},
	}
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages BIND global settings. This singleton must be imported with `bind_settings` before use. Attributes omitted from configuration retain their imported upstream values.", Version: 1, Attributes: settingsAttributesResource()}
}

func settingsDataSourceSchema() dschema.Schema {
	attrs := map[string]dschema.Attribute{
		"id": dschema.StringAttribute{Computed: true}, "enabled": dschema.BoolAttribute{Computed: true}, "disable_ipv6": dschema.BoolAttribute{Computed: true}, "enable_rpz": dschema.BoolAttribute{Computed: true},
		"listen_ipv4": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "listen_ipv6": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"query_source": dschema.StringAttribute{Computed: true}, "query_source_ipv6": dschema.StringAttribute{Computed: true}, "transfer_source": dschema.StringAttribute{Computed: true}, "transfer_source_ipv6": dschema.StringAttribute{Computed: true},
		"port": dschema.Int64Attribute{Computed: true}, "forwarders": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"filter_aaaa_ipv4": dschema.BoolAttribute{Computed: true}, "filter_aaaa_ipv6": dschema.BoolAttribute{Computed: true}, "filter_aaaa_acl": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"log_size_mb": dschema.Int64Attribute{Computed: true}, "log_level": dschema.StringAttribute{Computed: true}, "max_cache_size_percent": dschema.Int64Attribute{Computed: true},
		"legacy_recursion_acl_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "legacy_allow_transfer_acl_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType}, "legacy_allow_query_acl_ids": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
		"dnssec_validation": dschema.StringAttribute{Computed: true}, "hide_hostname": dschema.BoolAttribute{Computed: true}, "hide_version": dschema.BoolAttribute{Computed: true}, "disable_prefetch": dschema.BoolAttribute{Computed: true},
		"enable_rate_limiting": dschema.BoolAttribute{Computed: true}, "rate_limit_count": dschema.Int64Attribute{Computed: true}, "rate_limit_exceptions": dschema.SetAttribute{Computed: true, ElementType: types.StringType},
	}
	return dschema.Schema{MarkdownDescription: "Reads BIND global settings.", Attributes: attrs}
}

func settingsAPIToModel(d *apibind.SettingsResponse) (*settingsResourceModel, error) {
	g := d.General
	return &settingsResourceModel{
		ID: types.StringValue("bind_settings"), Enabled: types.BoolValue(tools.StringToBool(g.Enabled)), DisableIPv6: types.BoolValue(tools.StringToBool(g.DisableIPv6)), EnableRPZ: types.BoolValue(tools.StringToBool(g.EnableRPZ)),
		ListenIPv4: tools.StringSliceToSet([]string(g.ListenIPv4)), ListenIPv6: tools.StringSliceToSet([]string(g.ListenIPv6)), QuerySource: types.StringValue(g.QuerySource), QuerySourceIPv6: types.StringValue(g.QuerySourceIPv6), TransferSource: types.StringValue(g.TransferSource), TransferSourceIPv6: types.StringValue(g.TransferSourceIPv6),
		Port: types.Int64Value(tools.StringToInt64(g.Port)), Forwarders: tools.StringSliceToSet([]string(g.Forwarders)), FilterAAAAIPv4: types.BoolValue(tools.StringToBool(g.FilterAAAAIPv4)), FilterAAAAIPv6: types.BoolValue(tools.StringToBool(g.FilterAAAAIPv6)), FilterAAAAACL: tools.StringSliceToSet([]string(g.FilterAAAAACL)),
		LogSize: types.Int64Value(tools.StringToInt64(g.LogSize)), LogLevel: types.StringValue(g.LogLevel.String()), MaxCacheSize: types.Int64Value(tools.StringToInt64(g.MaxCacheSize)), RecursionACLIDs: tools.StringSliceToSet([]string(g.Recursion)), AllowTransferACLs: tools.StringSliceToSet([]string(g.AllowTransfer)), AllowQueryACLs: tools.StringSliceToSet([]string(g.AllowQuery)),
		DNSSECValidation: types.StringValue(g.DNSSECValidation.String()), HideHostname: types.BoolValue(tools.StringToBool(g.HideHostname)), HideVersion: types.BoolValue(tools.StringToBool(g.HideVersion)), DisablePrefetch: types.BoolValue(tools.StringToBool(g.DisablePrefetch)), EnableRateLimiting: types.BoolValue(tools.StringToBool(g.EnableRateLimiting)), RateLimitCount: types.Int64Value(tools.StringToInt64(g.RateLimitCount)), RateLimitExcept: tools.StringSliceToSet([]string(g.RateLimitExcept)),
	}, nil
}

func applySettingsModel(g *apibind.GeneralSettings, d *settingsResourceModel) {
	if !d.Enabled.IsNull() && !d.Enabled.IsUnknown() {
		g.Enabled = tools.BoolToString(d.Enabled.ValueBool())
	}
	if !d.DisableIPv6.IsNull() && !d.DisableIPv6.IsUnknown() {
		g.DisableIPv6 = tools.BoolToString(d.DisableIPv6.ValueBool())
	}
	if !d.EnableRPZ.IsNull() && !d.EnableRPZ.IsUnknown() {
		g.EnableRPZ = tools.BoolToString(d.EnableRPZ.ValueBool())
	}
	if !d.ListenIPv4.IsNull() && !d.ListenIPv4.IsUnknown() {
		g.ListenIPv4 = api.SelectedMapList(tools.SetToStringSlice(d.ListenIPv4))
	}
	if !d.ListenIPv6.IsNull() && !d.ListenIPv6.IsUnknown() {
		g.ListenIPv6 = api.SelectedMapList(tools.SetToStringSlice(d.ListenIPv6))
	}
	if !d.QuerySource.IsNull() && !d.QuerySource.IsUnknown() {
		g.QuerySource = d.QuerySource.ValueString()
	}
	if !d.QuerySourceIPv6.IsNull() && !d.QuerySourceIPv6.IsUnknown() {
		g.QuerySourceIPv6 = d.QuerySourceIPv6.ValueString()
	}
	if !d.TransferSource.IsNull() && !d.TransferSource.IsUnknown() {
		g.TransferSource = d.TransferSource.ValueString()
	}
	if !d.TransferSourceIPv6.IsNull() && !d.TransferSourceIPv6.IsUnknown() {
		g.TransferSourceIPv6 = d.TransferSourceIPv6.ValueString()
	}
	if !d.Port.IsNull() && !d.Port.IsUnknown() {
		g.Port = tools.Int64ToString(d.Port.ValueInt64())
	}
	if !d.Forwarders.IsNull() && !d.Forwarders.IsUnknown() {
		g.Forwarders = api.SelectedMapList(tools.SetToStringSlice(d.Forwarders))
	}
	if !d.FilterAAAAIPv4.IsNull() && !d.FilterAAAAIPv4.IsUnknown() {
		g.FilterAAAAIPv4 = tools.BoolToString(d.FilterAAAAIPv4.ValueBool())
	}
	if !d.FilterAAAAIPv6.IsNull() && !d.FilterAAAAIPv6.IsUnknown() {
		g.FilterAAAAIPv6 = tools.BoolToString(d.FilterAAAAIPv6.ValueBool())
	}
	if !d.FilterAAAAACL.IsNull() && !d.FilterAAAAACL.IsUnknown() {
		g.FilterAAAAACL = api.SelectedMapList(tools.SetToStringSlice(d.FilterAAAAACL))
	}
	if !d.LogSize.IsNull() && !d.LogSize.IsUnknown() {
		g.LogSize = tools.Int64ToString(d.LogSize.ValueInt64())
	}
	if !d.LogLevel.IsNull() && !d.LogLevel.IsUnknown() {
		g.LogLevel = api.SelectedMap(d.LogLevel.ValueString())
	}
	if !d.MaxCacheSize.IsNull() && !d.MaxCacheSize.IsUnknown() {
		g.MaxCacheSize = tools.Int64ToString(d.MaxCacheSize.ValueInt64())
	}
	if !d.RecursionACLIDs.IsNull() && !d.RecursionACLIDs.IsUnknown() {
		g.Recursion = api.SelectedMapList(tools.SetToStringSlice(d.RecursionACLIDs))
	}
	if !d.AllowTransferACLs.IsNull() && !d.AllowTransferACLs.IsUnknown() {
		g.AllowTransfer = api.SelectedMapList(tools.SetToStringSlice(d.AllowTransferACLs))
	}
	if !d.AllowQueryACLs.IsNull() && !d.AllowQueryACLs.IsUnknown() {
		g.AllowQuery = api.SelectedMapList(tools.SetToStringSlice(d.AllowQueryACLs))
	}
	if !d.DNSSECValidation.IsNull() && !d.DNSSECValidation.IsUnknown() {
		g.DNSSECValidation = api.SelectedMap(d.DNSSECValidation.ValueString())
	}
	if !d.HideHostname.IsNull() && !d.HideHostname.IsUnknown() {
		g.HideHostname = tools.BoolToString(d.HideHostname.ValueBool())
	}
	if !d.HideVersion.IsNull() && !d.HideVersion.IsUnknown() {
		g.HideVersion = tools.BoolToString(d.HideVersion.ValueBool())
	}
	if !d.DisablePrefetch.IsNull() && !d.DisablePrefetch.IsUnknown() {
		g.DisablePrefetch = tools.BoolToString(d.DisablePrefetch.ValueBool())
	}
	if !d.EnableRateLimiting.IsNull() && !d.EnableRateLimiting.IsUnknown() {
		g.EnableRateLimiting = tools.BoolToString(d.EnableRateLimiting.ValueBool())
	}
	if !d.RateLimitCount.IsNull() && !d.RateLimitCount.IsUnknown() {
		g.RateLimitCount = tools.Int64ToString(d.RateLimitCount.ValueInt64())
	}
	if !d.RateLimitExcept.IsNull() && !d.RateLimitExcept.IsUnknown() {
		g.RateLimitExcept = api.SelectedMapList(tools.SetToStringSlice(d.RateLimitExcept))
	}
}
