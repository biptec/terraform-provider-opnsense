package core

import (
	"context"
	"crypto/sha1"
	"fmt"
	"sort"
	"strings"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &hasyncSyncResource{}
var _ resource.ResourceWithConfigure = &hasyncSyncResource{}

type hasyncSyncModel struct {
	ID           types.String `tfsdk:"id"`
	Items        types.Set    `tfsdk:"items"`
	SyncOnCreate types.Bool   `tfsdk:"sync_on_create"`
	SyncOnDelete types.Bool   `tfsdk:"sync_on_delete"`
}

type hasyncSyncResource struct{ client opnsense.Client }

func newHasyncSyncResource() resource.Resource { return &hasyncSyncResource{} }

func (r *hasyncSyncResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_core_hasync_sync"
}

func (r *hasyncSyncResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "State-only selective OPNsense HA configuration synchronization barrier. It does not own a remote object: Create/Update and Delete optionally invoke the native HA sync mechanism for the requested, HA-enabled item identifiers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"items": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
				MarkdownDescription: "HA synchronization item identifiers to synchronize, such as `interface_vlans`, `virtualip`, `rules`, `staticroutes`, or `haproxy_objects`. Every item must already be enabled in OPNsense High Availability settings.",
			},
			"sync_on_create": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Invoke selective HA synchronization after this barrier is created or updated. Use this for the post-apply barrier.",
			},
			"sync_on_delete": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Invoke selective HA synchronization when this barrier is deleted. Use this for the post-destroy barrier.",
			},
		},
	}
}

func (r *hasyncSyncResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(client)
}

func hasyncSyncItems(ctx context.Context, data *hasyncSyncModel) ([]string, error) {
	var items []string
	diagnostics := data.Items.ElementsAs(ctx, &items, false)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("decode HA synchronization items: %s", diagnostics.Errors()[0].Summary())
	}
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("HA synchronization items must not contain empty values")
		}
		normalized = append(normalized, item)
	}
	sort.Strings(normalized)
	return normalized, nil
}

func hasyncSyncID(items []string, onCreate, onDelete bool) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%s|create=%t|delete=%t", strings.Join(items, ","), onCreate, onDelete)))
	return fmt.Sprintf("ha_sync_%x", sum[:8])
}

func mergeHasyncSyncItems(left, right []string) []string {
	seen := make(map[string]struct{}, len(left)+len(right))
	merged := make([]string, 0, len(left)+len(right))
	for _, items := range [][]string{left, right} {
		for _, item := range items {
			if _, exists := seen[item]; exists {
				continue
			}
			seen[item] = struct{}{}
			merged = append(merged, item)
		}
	}
	sort.Strings(merged)
	return merged
}

func (r *hasyncSyncResource) syncItems(ctx context.Context, items []string) error {
	if _, err := r.client.Core().HasyncSync(ctx, items); err != nil {
		return err
	}
	return nil
}

func (r *hasyncSyncResource) sync(ctx context.Context, data *hasyncSyncModel) error {
	items, err := hasyncSyncItems(ctx, data)
	if err != nil {
		return err
	}
	return r.syncItems(ctx, items)
}

func (r *hasyncSyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data hasyncSyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.SyncOnCreate.ValueBool() && !data.SyncOnDelete.ValueBool() {
		resp.Diagnostics.AddError("Invalid HA Synchronization Barrier", "At least one of sync_on_create or sync_on_delete must be true.")
		return
	}
	items, err := hasyncSyncItems(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid HA Synchronization Items", err.Error())
		return
	}
	if data.SyncOnCreate.ValueBool() {
		if err := r.sync(ctx, &data); err != nil {
			resp.Diagnostics.AddError("Unable to Synchronize OPNsense HA Configuration", err.Error())
			return
		}
	}
	data.ID = types.StringValue(hasyncSyncID(items, data.SyncOnCreate.ValueBool(), data.SyncOnDelete.ValueBool()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hasyncSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data hasyncSyncModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// State-only barrier: there is no remote object to refresh.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hasyncSyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hasyncSyncModel
	var state hasyncSyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.SyncOnCreate.ValueBool() && !plan.SyncOnDelete.ValueBool() {
		resp.Diagnostics.AddError("Invalid HA Synchronization Barrier", "At least one of sync_on_create or sync_on_delete must be true.")
		return
	}
	planItems, err := hasyncSyncItems(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid HA Synchronization Items", err.Error())
		return
	}
	stateItems, err := hasyncSyncItems(ctx, &state)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Prior HA Synchronization Items", err.Error())
		return
	}
	if plan.SyncOnCreate.ValueBool() {
		// Synchronize the union so removing the final object in a category also
		// propagates that deletion to the peer instead of leaving stale state.
		if err := r.syncItems(ctx, mergeHasyncSyncItems(stateItems, planItems)); err != nil {
			resp.Diagnostics.AddError("Unable to Synchronize OPNsense HA Configuration", err.Error())
			return
		}
	}
	plan.ID = types.StringValue(hasyncSyncID(planItems, plan.SyncOnCreate.ValueBool(), plan.SyncOnDelete.ValueBool()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hasyncSyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data hasyncSyncModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.SyncOnDelete.ValueBool() {
		if err := r.sync(ctx, &data); err != nil {
			resp.Diagnostics.AddError("Unable to Synchronize OPNsense HA Configuration After Delete", err.Error())
		}
	}
}
