package caddy

import (
	"fmt"
	"github.com/biptec/opnsense-go/pkg/api"
	apicaddy "github.com/biptec/opnsense-go/pkg/caddy"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

type handlerResourceModel struct {
	ID                                 types.String `tfsdk:"id"`
	Enabled                            types.Bool   `tfsdk:"enabled"`
	DomainID                           types.String `tfsdk:"domain_id"`
	SubdomainID                        types.String `tfsdk:"subdomain_id"`
	HandlerType                        types.String `tfsdk:"handler_type"`
	Path                               types.String `tfsdk:"path"`
	AccessListID                       types.String `tfsdk:"access_list_id"`
	BasicAuthIDs                       types.Set    `tfsdk:"basic_auth_ids"`
	HeaderIDs                          types.Set    `tfsdk:"header_ids"`
	Directive                          types.String `tfsdk:"directive"`
	UpstreamDomains                    types.Set    `tfsdk:"upstream_domains"`
	UpstreamPort                       types.Int64  `tfsdk:"upstream_port"`
	UpstreamPath                       types.String `tfsdk:"upstream_path"`
	ForwardAuth                        types.Bool   `tfsdk:"forward_auth"`
	UpstreamProtocol                   types.String `tfsdk:"upstream_protocol"`
	HTTPVersion                        types.String `tfsdk:"http_version"`
	HTTPKeepalive                      types.Int64  `tfsdk:"http_keepalive"`
	TLSSkipVerify                      types.Bool   `tfsdk:"tls_insecure_skip_verify"`
	TLSTrustCARefID                    types.String `tfsdk:"tls_trust_ca_ref_id"`
	TLSServerName                      types.String `tfsdk:"tls_server_name"`
	LoadBalancingPolicy                types.String `tfsdk:"load_balancing_policy"`
	LoadBalancingRetries               types.Int64  `tfsdk:"load_balancing_retries"`
	LoadBalancingTryDuration           types.Int64  `tfsdk:"load_balancing_try_duration"`
	LoadBalancingTryInterval           types.Int64  `tfsdk:"load_balancing_try_interval"`
	PassiveHealthFailDuration          types.Int64  `tfsdk:"passive_health_fail_duration"`
	PassiveHealthMaxFails              types.Int64  `tfsdk:"passive_health_max_fails"`
	PassiveHealthUnhealthyStatus       types.String `tfsdk:"passive_health_unhealthy_status"`
	PassiveHealthUnhealthyLatency      types.Int64  `tfsdk:"passive_health_unhealthy_latency"`
	PassiveHealthUnhealthyRequestCount types.Int64  `tfsdk:"passive_health_unhealthy_request_count"`
	HealthURI                          types.String `tfsdk:"health_uri"`
	HealthUpstream                     types.String `tfsdk:"health_upstream"`
	HealthPort                         types.Int64  `tfsdk:"health_port"`
	HealthInterval                     types.Int64  `tfsdk:"health_interval"`
	HealthPasses                       types.Int64  `tfsdk:"health_passes"`
	HealthFails                        types.Int64  `tfsdk:"health_fails"`
	HealthTimeout                      types.Int64  `tfsdk:"health_timeout"`
	HealthStatus                       types.String `tfsdk:"health_status"`
	HealthBody                         types.String `tfsdk:"health_body"`
	HealthFollowRedirects              types.Bool   `tfsdk:"health_follow_redirects"`
	Description                        types.String `tfsdk:"description"`
}

func optionalPositiveInt(description string, min int64) schema.Int64Attribute {
	return schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(-1), MarkdownDescription: description + " `-1` uses the Caddy default.", Validators: []validator.Int64{int64validator.Any(int64validator.OneOf(-1), int64validator.AtLeast(min))}}
}
func handlerResourceSchema() schema.Schema {
	return schema.Schema{MarkdownDescription: "Manages a Caddy HTTP handler and its upstream reverse-proxy configuration.", Version: 1, Attributes: map[string]schema.Attribute{
		"id":                       schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the handler.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"enabled":                  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether to enable this handler. Defaults to `true`."},
		"domain_id":                schema.StringAttribute{Required: true, MarkdownDescription: "UUID of the Caddy domain containing this handler."},
		"subdomain_id":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional UUID of a Caddy subdomain. Defaults to `\"\"`."},
		"handler_type":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("handle"), MarkdownDescription: "Handler type: `handle` or `handle_path`. Defaults to `handle`.", Validators: []validator.String{stringvalidator.OneOf("handle", "handle_path")}},
		"path":                     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional request path matcher starting with `/`. Defaults to `\"\"`."},
		"access_list_id":           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional Caddy access-list UUID. Defaults to `\"\"`."},
		"basic_auth_ids":           schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), MarkdownDescription: "Caddy basic-auth entry UUIDs. Defaults to an empty set."},
		"header_ids":               schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)), MarkdownDescription: "Caddy header-manipulation UUIDs. Defaults to an empty set."},
		"directive":                schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("reverse_proxy"), MarkdownDescription: "Handler directive: `reverse_proxy` or `redir`. Defaults to `reverse_proxy`.", Validators: []validator.String{stringvalidator.OneOf("reverse_proxy", "redir")}},
		"upstream_domains":         schema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "One or more upstream IP addresses, hostnames, or FQDNs."},
		"upstream_port":            schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(-1), MarkdownDescription: "Optional upstream port. `-1` leaves it unset.", Validators: []validator.Int64{int64validator.Any(int64validator.OneOf(-1), int64validator.Between(1, 65535))}},
		"upstream_path":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional path prepended to upstream requests. Defaults to `\"\"`."},
		"forward_auth":             schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether this handler is used for forward authentication. Defaults to `false`."},
		"upstream_protocol":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("http"), MarkdownDescription: "Upstream protocol: `http`, `https`, or `h2c`. Defaults to `http`.", Validators: []validator.String{stringvalidator.OneOf("http", "https", "h2c")}},
		"http_version":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional HTTP version restriction: `http1`, `http2`, or `http3`. Empty uses Caddy defaults.", Validators: []validator.String{stringvalidator.OneOf("", "http1", "http2", "http3")}},
		"http_keepalive":           optionalPositiveInt("HTTP keepalive duration in seconds.", 0),
		"tls_insecure_skip_verify": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether to skip upstream TLS certificate verification. Defaults to `false`; prefer `tls_trust_ca_ref_id`."},
		"tls_trust_ca_ref_id":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Reference ID of an OPNsense CA trusted for the HTTPS upstream. Defaults to `\"\"`."},
		"tls_server_name":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "TLS server name used to verify the upstream certificate. Defaults to `\"\"`."},
		"load_balancing_policy":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Load-balancing policy. Empty means random; alternatives include `first`, `round_robin`, `least_conn`, `ip_hash`, `client_ip_hash`, and `uri_hash`.", Validators: []validator.String{stringvalidator.OneOf("", "first", "round_robin", "least_conn", "ip_hash", "client_ip_hash", "uri_hash")}},
		"load_balancing_retries":   optionalPositiveInt("Number of load-balancing retries.", 1), "load_balancing_try_duration": optionalPositiveInt("Load-balancing try duration in seconds.", 1), "load_balancing_try_interval": optionalPositiveInt("Load-balancing try interval in seconds.", 0),
		"passive_health_fail_duration": optionalPositiveInt("Passive health failure duration in seconds.", 1), "passive_health_max_fails": optionalPositiveInt("Passive health maximum failures.", 2),
		"passive_health_unhealthy_status":  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "HTTP status or status class considered unhealthy, for example `404` or `5xx`. Defaults to `\"\"`."},
		"passive_health_unhealthy_latency": optionalPositiveInt("Unhealthy latency threshold in seconds.", 1), "passive_health_unhealthy_request_count": optionalPositiveInt("Unhealthy request-count threshold.", 1),
		"health_uri":      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Active health-check URI. Defaults to `\"\"`."},
		"health_upstream": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Dedicated active health-check upstream. Defaults to `\"\"`."},
		"health_port":     schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(-1), MarkdownDescription: "Active health-check port. `-1` leaves it unset.", Validators: []validator.Int64{int64validator.Any(int64validator.OneOf(-1), int64validator.Between(1, 65535))}},
		"health_interval": optionalPositiveInt("Active health-check interval in seconds.", 1), "health_passes": optionalPositiveInt("Consecutive successful checks required.", 1), "health_fails": optionalPositiveInt("Consecutive failed checks required.", 1), "health_timeout": optionalPositiveInt("Active health-check timeout in seconds.", 1),
		"health_status":           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Expected active health-check HTTP status, for example `200` or `2xx`. Defaults to `\"\"`."},
		"health_body":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Expected response-body substring for active health checks. Defaults to `\"\"`."},
		"health_follow_redirects": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether active health checks follow redirects. Defaults to `false`."},
		"description":             schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Optional description. Defaults to `\"\"`."},
	}}
}

func handlerDataSourceSchema() dschema.Schema {
	attrs := map[string]dschema.Attribute{"id": dschema.StringAttribute{Required: true}}
	for _, n := range []string{"domain_id", "subdomain_id", "handler_type", "path", "access_list_id", "directive", "upstream_path", "upstream_protocol", "http_version", "tls_trust_ca_ref_id", "tls_server_name", "load_balancing_policy", "passive_health_unhealthy_status", "health_uri", "health_upstream", "health_status", "health_body", "description"} {
		attrs[n] = dschema.StringAttribute{Computed: true}
	}
	for _, n := range []string{"enabled", "forward_auth", "tls_insecure_skip_verify", "health_follow_redirects"} {
		attrs[n] = dschema.BoolAttribute{Computed: true}
	}
	for _, n := range []string{"basic_auth_ids", "header_ids", "upstream_domains"} {
		attrs[n] = dschema.SetAttribute{Computed: true, ElementType: types.StringType}
	}
	for _, n := range []string{"upstream_port", "http_keepalive", "load_balancing_retries", "load_balancing_try_duration", "load_balancing_try_interval", "passive_health_fail_duration", "passive_health_max_fails", "passive_health_unhealthy_latency", "passive_health_unhealthy_request_count", "health_port", "health_interval", "health_passes", "health_fails", "health_timeout"} {
		attrs[n] = dschema.Int64Attribute{Computed: true}
	}
	return dschema.Schema{MarkdownDescription: "Reads a Caddy HTTP handler by UUID.", Attributes: attrs}
}

func intToAPI(v types.Int64) string {
	if v.IsNull() || v.IsUnknown() || v.ValueInt64() == -1 {
		return ""
	}
	return tools.Int64ToString(v.ValueInt64())
}
func apiInt(v string) types.Int64 {
	if v == "" {
		return types.Int64Value(-1)
	}
	return types.Int64Value(tools.StringToInt64(v))
}
func protocolToAPI(v string) (string, error) {
	switch v {
	case "http":
		return "0", nil
	case "https":
		return "1", nil
	case "h2c":
		return "2", nil
	default:
		return "", fmt.Errorf("unsupported upstream protocol %q", v)
	}
}
func apiToProtocol(v string) string {
	switch v {
	case "1":
		return "https"
	case "2":
		return "h2c"
	default:
		return "http"
	}
}
func convertHandlerSchemaToStruct(d *handlerResourceModel) (*apicaddy.Handler, error) {
	proto, err := protocolToAPI(d.UpstreamProtocol.ValueString())
	if err != nil {
		return nil, err
	}
	return &apicaddy.Handler{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Domain: api.SelectedMap(d.DomainID.ValueString()), Subdomain: api.SelectedMap(d.SubdomainID.ValueString()), Type: api.SelectedMap(d.HandlerType.ValueString()), Path: d.Path.ValueString(), AccessList: api.SelectedMap(d.AccessListID.ValueString()), BasicAuth: api.SelectedMapList(tools.SetToStringSlice(d.BasicAuthIDs)), Headers: api.SelectedMapList(tools.SetToStringSlice(d.HeaderIDs)), Directive: api.SelectedMap(d.Directive.ValueString()), UpstreamDomains: api.SelectedMapList(tools.SetToStringSlice(d.UpstreamDomains)), UpstreamPort: intToAPI(d.UpstreamPort), UpstreamPath: d.UpstreamPath.ValueString(), ForwardAuth: tools.BoolToString(d.ForwardAuth.ValueBool()), UpstreamProtocol: api.SelectedMap(proto), HTTPVersion: api.SelectedMap(d.HTTPVersion.ValueString()), HTTPKeepalive: intToAPI(d.HTTPKeepalive), TLSSkipVerify: tools.BoolToString(d.TLSSkipVerify.ValueBool()), TLSTrustPool: api.SelectedMap(d.TLSTrustCARefID.ValueString()), TLSServerName: d.TLSServerName.ValueString(), LoadBalancingPolicy: api.SelectedMap(d.LoadBalancingPolicy.ValueString()), LoadBalancingRetries: intToAPI(d.LoadBalancingRetries), LoadBalancingTryDuration: intToAPI(d.LoadBalancingTryDuration), LoadBalancingTryInterval: intToAPI(d.LoadBalancingTryInterval), PassiveHealthFailDuration: intToAPI(d.PassiveHealthFailDuration), PassiveHealthMaxFails: intToAPI(d.PassiveHealthMaxFails), PassiveHealthUnhealthyStatus: d.PassiveHealthUnhealthyStatus.ValueString(), PassiveHealthUnhealthyLatency: intToAPI(d.PassiveHealthUnhealthyLatency), PassiveHealthUnhealthyRequestCount: intToAPI(d.PassiveHealthUnhealthyRequestCount), HealthURI: d.HealthURI.ValueString(), HealthUpstream: d.HealthUpstream.ValueString(), HealthPort: intToAPI(d.HealthPort), HealthInterval: intToAPI(d.HealthInterval), HealthPasses: intToAPI(d.HealthPasses), HealthFails: intToAPI(d.HealthFails), HealthTimeout: intToAPI(d.HealthTimeout), HealthStatus: d.HealthStatus.ValueString(), HealthBody: d.HealthBody.ValueString(), HealthFollowRedirects: tools.BoolToString(d.HealthFollowRedirects.ValueBool()), Description: d.Description.ValueString()}, nil
}
func convertHandlerStructToSchema(d *apicaddy.Handler) (*handlerResourceModel, error) {
	return &handlerResourceModel{Enabled: types.BoolValue(tools.StringToBool(d.Enabled)), DomainID: types.StringValue(d.Domain.String()), SubdomainID: types.StringValue(d.Subdomain.String()), HandlerType: types.StringValue(d.Type.String()), Path: types.StringValue(d.Path), AccessListID: types.StringValue(d.AccessList.String()), BasicAuthIDs: tools.StringSliceToSet([]string(d.BasicAuth)), HeaderIDs: tools.StringSliceToSet([]string(d.Headers)), Directive: types.StringValue(d.Directive.String()), UpstreamDomains: tools.StringSliceToSet([]string(d.UpstreamDomains)), UpstreamPort: apiInt(d.UpstreamPort), UpstreamPath: types.StringValue(d.UpstreamPath), ForwardAuth: types.BoolValue(tools.StringToBool(d.ForwardAuth)), UpstreamProtocol: types.StringValue(apiToProtocol(d.UpstreamProtocol.String())), HTTPVersion: types.StringValue(d.HTTPVersion.String()), HTTPKeepalive: apiInt(d.HTTPKeepalive), TLSSkipVerify: types.BoolValue(tools.StringToBool(d.TLSSkipVerify)), TLSTrustCARefID: types.StringValue(d.TLSTrustPool.String()), TLSServerName: types.StringValue(d.TLSServerName), LoadBalancingPolicy: types.StringValue(d.LoadBalancingPolicy.String()), LoadBalancingRetries: apiInt(d.LoadBalancingRetries), LoadBalancingTryDuration: apiInt(d.LoadBalancingTryDuration), LoadBalancingTryInterval: apiInt(d.LoadBalancingTryInterval), PassiveHealthFailDuration: apiInt(d.PassiveHealthFailDuration), PassiveHealthMaxFails: apiInt(d.PassiveHealthMaxFails), PassiveHealthUnhealthyStatus: types.StringValue(d.PassiveHealthUnhealthyStatus), PassiveHealthUnhealthyLatency: apiInt(d.PassiveHealthUnhealthyLatency), PassiveHealthUnhealthyRequestCount: apiInt(d.PassiveHealthUnhealthyRequestCount), HealthURI: types.StringValue(d.HealthURI), HealthUpstream: types.StringValue(d.HealthUpstream), HealthPort: apiInt(d.HealthPort), HealthInterval: apiInt(d.HealthInterval), HealthPasses: apiInt(d.HealthPasses), HealthFails: apiInt(d.HealthFails), HealthTimeout: apiInt(d.HealthTimeout), HealthStatus: types.StringValue(d.HealthStatus), HealthBody: types.StringValue(d.HealthBody), HealthFollowRedirects: types.BoolValue(tools.StringToBool(d.HealthFollowRedirects)), Description: types.StringValue(d.Description)}, nil
}
