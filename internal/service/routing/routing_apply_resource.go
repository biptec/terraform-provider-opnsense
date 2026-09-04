package routing

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	apirouting "github.com/biptec/opnsense-go/pkg/routing"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &routingApplyResource{}
var _ resource.ResourceWithConfigure = &routingApplyResource{}

type routingApplyResource struct{ routingResourceClient }

type routingApplyResourceModel struct {
	Trigger types.String `tfsdk:"trigger"`
	Id      types.String `tfsdk:"id"`
}

func newRoutingApplyResource() resource.Resource { return &routingApplyResource{} }

func (r *routingApplyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_apply"
}

func (r *routingApplyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reconfigures the OPNsense routing service when a caller-controlled triggger changes.",
		Attributes: map[string]schema.Attribute{
			"trigger": schema.StringAttribute{
				MarkdownDescription: "Opaque caller-controlled value. Changing this value re-applies the routing configuration.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Stable identifier for the routing apply action.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

type routingApplyClient interface {
	ReconfigureService(context.Context, api.Endpoint) error
}

func applyRoutingConfig(ctx context.Context, client routingApplyClient) error {
	if err := client.ReconfigureService(ctx, apirouting.GatewayOpts.Reconfigure); err != nil {
		return fmt.Errorf("reconfigure routing: %w", err)
	}
	return nil
}

func (r *routingApplyResource) apply(ctx context.Context, data *routingApplyResourceModel) error {
	if err := applyRoutingConfig(ctx, r.client.Routing().Client()); err != nil {
		return err
	}
	data.Id = types.StringValue("routing")
	return nil
}

func (r *routingApplyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data routingApplyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Apply Routing Configuration", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *routingApplyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data routingApplyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *routingApplyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data routingApplyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Unable to Apply Routing Configuration", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *routingApplyResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}
