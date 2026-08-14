package caddy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type settingsDNSConfigValidator struct{}

func (settingsDNSConfigValidator) Description(context.Context) string {
	return "validates provider-specific Caddy DNS credentials and RFC2136 settings"
}

func (v settingsDNSConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (settingsDNSConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var provider, apiKey, rfcKey, server, keyName types.String
	var version types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_provider"), &provider)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_api_key"), &apiKey)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_rfc2136_key"), &rfcKey)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_rfc2136_server"), &server)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_rfc2136_key_name"), &keyName)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dns_credentials_version"), &version)...)
	if resp.Diagnostics.HasError() || provider.IsNull() || provider.IsUnknown() {
		return
	}

	if provider.ValueString() == "rfc2136" {
		if !server.IsUnknown() && (server.IsNull() || server.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(path.Root("dns_rfc2136_server"), "Missing RFC2136 Server", "dns_rfc2136_server is required when dns_provider is rfc2136.")
		}
		if !keyName.IsUnknown() && (keyName.IsNull() || keyName.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(path.Root("dns_rfc2136_key_name"), "Missing RFC2136 TSIG Key Name", "dns_rfc2136_key_name is required when dns_provider is rfc2136.")
		}
	}

	if !apiKey.IsNull() && !apiKey.IsUnknown() && provider.ValueString() != "cloudflare" {
		resp.Diagnostics.AddAttributeError(path.Root("dns_api_key"), "DNS Credential Does Not Match Provider", "dns_api_key can only be configured when dns_provider is cloudflare.")
	}
	if !rfcKey.IsNull() && !rfcKey.IsUnknown() && provider.ValueString() != "rfc2136" {
		resp.Diagnostics.AddAttributeError(path.Root("dns_rfc2136_key"), "DNS Credential Does Not Match Provider", "dns_rfc2136_key can only be configured when dns_provider is rfc2136.")
	}
	credentialConfigured := (!apiKey.IsNull() && !apiKey.IsUnknown()) || (!rfcKey.IsNull() && !rfcKey.IsUnknown())
	if credentialConfigured && !version.IsUnknown() && (version.IsNull() || version.ValueInt64() < 1) {
		resp.Diagnostics.AddAttributeError(
			path.Root("dns_credentials_version"),
			"Missing DNS Credential Version",
			"Set dns_credentials_version to at least 1 when supplying a write-only DNS credential, and increment it whenever that credential changes.",
		)
	}
}
