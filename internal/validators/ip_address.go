package validators

import (
	"context"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type ipAddressValidator struct{}

func (v ipAddressValidator) Description(_ context.Context) string {
	return "must be a valid IPv4 or IPv6 address"
}
func (v ipAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v ipAddressValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() || request.ConfigValue.ValueString() == "" {
		return
	}
	if _, err := netip.ParseAddr(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(request.Path, v.Description(ctx), request.ConfigValue.ValueString()))
	}
}
func IPAddress() validator.String { return ipAddressValidator{} }

type ipVersionValidator struct{ version int }

func (v ipVersionValidator) Description(_ context.Context) string {
	if v.version == 4 {
		return "must be a valid IPv4 address"
	}
	return "must be a valid IPv6 address"
}
func (v ipVersionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v ipVersionValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() || request.ConfigValue.ValueString() == "" {
		return
	}
	addr, err := netip.ParseAddr(request.ConfigValue.ValueString())
	valid := err == nil && ((v.version == 4 && addr.Is4()) || (v.version == 6 && addr.Is6()))
	if !valid {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(request.Path, v.Description(ctx), request.ConfigValue.ValueString()))
	}
}

func IPv4Address() validator.String { return ipVersionValidator{version: 4} }
func IPv6Address() validator.String { return ipVersionValidator{version: 6} }
