package system

import (
	"context"
	"fmt"

	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &webguiResource{}
var _ resource.ResourceWithConfigure = &webguiResource{}
var _ resource.ResourceWithImportState = &webguiResource{}

type webguiResource struct{ client opnsense.Client }

func newWebguiResource() resource.Resource { return &webguiResource{} }

func (r *webguiResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_webgui"
}

func (r *webguiResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = webguiResourceSchema()
}

func (r *webguiResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}

func (r *webguiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data webguiResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure Web GUI", err.Error())
		return
	}
	data.ID = types.StringValue("system_webgui")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webguiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var prior webguiResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.ApiExtensions().WebguiGet(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Web GUI Settings", err.Error())
		return
	}
	state := webguiFromAPI(&result.Webgui, prior.AllowReaddress.ValueBool())
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *webguiResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data webguiResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Configure Web GUI", err.Error())
		return
	}
	data.ID = types.StringValue("system_webgui")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webguiResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Singleton Resource Removed From State Only",
		"Web GUI settings remain unchanged in OPNsense. Re-import with ID `system_webgui` to manage them again.",
	)
}

func (r *webguiResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "system_webgui" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Web GUI singleton must be imported with ID `system_webgui`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("allow_readdress"), false)...)
}

func (r *webguiResource) apply(ctx context.Context, data *webguiResourceModel) error {
	desired, err := webguiToAPI(ctx, data)
	if err != nil {
		return err
	}
	currentResponse, err := r.client.ApiExtensions().WebguiGet(ctx)
	if err != nil {
		return err
	}
	if currentResponse == nil {
		return fmt.Errorf("web GUI get API returned an empty response")
	}
	current := &currentResponse.Webgui
	if webguiEqual(current, desired) {
		return nil
	}
	if webguiDisruptiveChange(current, desired) && !data.AllowReaddress.ValueBool() {
		return fmt.Errorf("listener interface, protocol, port, or certificate change requires allow_readdress = true")
	}

	result, err := r.client.ApiExtensions().WebguiSet(ctx, desired)
	if err != nil {
		return err
	}
	if err = validateWebguiAction("settings update", result); err != nil {
		return err
	}
	reconfigured, err := r.client.ApiExtensions().WebguiReconfigure(ctx)
	if err != nil {
		return err
	}
	if err = validateWebguiAction("reconfigure", reconfigured); err != nil {
		return err
	}
	tflog.Trace(ctx, "configured Web GUI listener settings")
	return nil
}

func webguiToAPI(ctx context.Context, data *webguiResourceModel) (*apiextensions.WebguiSettings, error) {
	interfaces, err := stringSet(ctx, data.Interfaces)
	if err != nil {
		return nil, err
	}
	alternateHostnames, err := stringSet(ctx, data.AlternateHostnames)
	if err != nil {
		return nil, err
	}
	var timeout *int
	if !data.SessionTimeout.IsNull() && !data.SessionTimeout.IsUnknown() {
		value := int(data.SessionTimeout.ValueInt64())
		timeout = &value
	}
	return &apiextensions.WebguiSettings{
		Protocol:            data.Protocol.ValueString(),
		Port:                int(data.Port.ValueInt64()),
		Interfaces:          interfaces,
		CertificateRef:      data.CertificateRef.ValueString(),
		SessionTimeout:      timeout,
		HSTS:                data.HSTS.ValueBool(),
		DisableHTTPRedirect: data.DisableHTTPRedirect.ValueBool(),
		AlternateHostnames:  alternateHostnames,
	}, nil
}

func webguiFromAPI(data *apiextensions.WebguiSettings, allowReaddress bool) *webguiResourceModel {
	result := &webguiResourceModel{
		Protocol:            types.StringValue(data.Protocol),
		Port:                types.Int64Value(int64(data.Port)),
		Interfaces:          stringSetValue(data.Interfaces),
		CertificateRef:      types.StringValue(data.CertificateRef),
		HSTS:                types.BoolValue(data.HSTS),
		DisableHTTPRedirect: types.BoolValue(data.DisableHTTPRedirect),
		AlternateHostnames:  stringSetValue(data.AlternateHostnames),
		AllowReaddress:      types.BoolValue(allowReaddress),
		ID:                  types.StringValue("system_webgui"),
	}
	if data.SessionTimeout == nil {
		result.SessionTimeout = types.Int64Null()
	} else {
		result.SessionTimeout = types.Int64Value(int64(*data.SessionTimeout))
	}
	return result
}

func webguiDisruptiveChange(current, desired *apiextensions.WebguiSettings) bool {
	return current.Protocol != desired.Protocol ||
		current.Port != desired.Port ||
		current.CertificateRef != desired.CertificateRef ||
		!sameStrings(current.Interfaces, desired.Interfaces)
}

func webguiEqual(current, desired *apiextensions.WebguiSettings) bool {
	if webguiDisruptiveChange(current, desired) ||
		current.HSTS != desired.HSTS ||
		current.DisableHTTPRedirect != desired.DisableHTTPRedirect ||
		!sameStrings(current.AlternateHostnames, desired.AlternateHostnames) {
		return false
	}
	if current.SessionTimeout == nil || desired.SessionTimeout == nil {
		return current.SessionTimeout == nil && desired.SessionTimeout == nil
	}
	return *current.SessionTimeout == *desired.SessionTimeout
}
