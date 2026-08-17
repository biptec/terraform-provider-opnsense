package system

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/biptec/opnsense-go/pkg/api"
	apiextensions "github.com/biptec/opnsense-go/pkg/api_extensions"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &carpHealthCheckResource{}
var _ resource.ResourceWithConfigure = &carpHealthCheckResource{}
var _ resource.ResourceWithImportState = &carpHealthCheckResource{}

type carpHealthCheckResource struct{ client opnsense.Client }

func newCarpHealthCheckResource() resource.Resource { return &carpHealthCheckResource{} }
func (r *carpHealthCheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_carp_health_check"
}
func (r *carpHealthCheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = carpHealthCheckResourceSchema()
}
func (r *carpHealthCheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req, resp)
}
func (r *carpHealthCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data carpHealthCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	desired, err := carpHealthCheckToAPI(&data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid CARP Health Check", err.Error())
		return
	}
	id, err := r.client.ApiExtensions().AddCarpHealthCheck(ctx, desired)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create CARP Health Check", err.Error())
		return
	}
	remote, err := r.client.ApiExtensions().GetCarpHealthCheck(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("CARP Health Check Created but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, carpHealthCheckFromAPI(remote, id))...)
}
func (r *carpHealthCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data carpHealthCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	remote, err := r.client.ApiExtensions().GetCarpHealthCheck(ctx, data.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read CARP Health Check", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, carpHealthCheckFromAPI(remote, data.ID.ValueString()))...)
}
func (r *carpHealthCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data carpHealthCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	desired, err := carpHealthCheckToAPI(&data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid CARP Health Check", err.Error())
		return
	}
	if err := r.client.ApiExtensions().UpdateCarpHealthCheck(ctx, data.ID.ValueString(), desired); err != nil {
		resp.Diagnostics.AddError("Unable to Update CARP Health Check", err.Error())
		return
	}
	remote, err := r.client.ApiExtensions().GetCarpHealthCheck(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("CARP Health Check Updated but Read Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, carpHealthCheckFromAPI(remote, data.ID.ValueString()))...)
}
func (r *carpHealthCheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data carpHealthCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.ApiExtensions().DeleteCarpHealthCheck(ctx, data.ID.ValueString()); err != nil {
		var notFound *errs.NotFoundError
		if !errors.As(err, &notFound) {
			resp.Diagnostics.AddError("Unable to Delete CARP Health Check", err.Error())
		}
	}
}
func (r *carpHealthCheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func carpHealthCheckToAPI(data *carpHealthCheckResourceModel) (*apiextensions.CarpHealthCheck, error) {
	target, err := netip.ParseAddr(data.Target.ValueString())
	if err != nil || !target.Is4() {
		return nil, fmt.Errorf("target must be an IPv4 address")
	}
	return &apiextensions.CarpHealthCheck{Enabled: api.BoolString(tools.BoolToString(data.Enabled.ValueBool())), Name: data.Name.ValueString(), Interface: api.SelectedMap(data.Interface.ValueString()), Target: target.String()}, nil
}
func carpHealthCheckFromAPI(data *apiextensions.CarpHealthCheck, id string) *carpHealthCheckResourceModel {
	return &carpHealthCheckResourceModel{Enabled: types.BoolValue(data.Enabled.Bool()), Name: types.StringValue(data.Name), Interface: types.StringValue(data.Interface.String()), Target: types.StringValue(data.Target), ID: types.StringValue(id)}
}
