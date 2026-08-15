package core

import (
	"github.com/biptec/opnsense-go/pkg/api"
	apicore "github.com/biptec/opnsense-go/pkg/core"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const hasyncID = "core_hasync"

type hasyncModel struct {
	DisablePreempt  types.Bool   `tfsdk:"disable_preempt"`
	DisconnectPPPs  types.Bool   `tfsdk:"disconnect_ppps"`
	PfsyncInterface types.String `tfsdk:"pfsync_interface"`
	PfsyncPeerIP    types.String `tfsdk:"pfsync_peer_ip"`
	PfsyncVersion   types.String `tfsdk:"pfsync_version"`
	PfsyncDefer     types.Bool   `tfsdk:"pfsync_defer"`
	SynchronizeToIP types.String `tfsdk:"synchronize_to_ip"`
	VerifyPeer      types.Bool   `tfsdk:"verify_peer"`
	Username        types.String `tfsdk:"username"`
	Password        types.String `tfsdk:"password"`
	SyncItems       types.Set    `tfsdk:"sync_items"`
	ID              types.String `tfsdk:"id"`
}

func hasyncResourceSchema() schema.Schema {
	return schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages OPNsense High Availability synchronization settings, including pfsync state replication and optional XMLRPC configuration synchronization.",
		Attributes: map[string]schema.Attribute{
			"disable_preempt":   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "When enabled, a recovering preferred CARP node does not preempt an already active master."},
			"disconnect_ppps":   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "When enabled, PPP-type interfaces are disconnected while this node is a CARP backup."},
			"pfsync_interface":  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Logical interface used to synchronize firewall states with pfsync. An empty value disables pfsync."},
			"pfsync_peer_ip":    schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional unicast IPv4 peer used for pfsync state synchronization."},
			"pfsync_version":    schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{stringvalidator.OneOf("1301", "1400")}, MarkdownDescription: "pfsync compatibility version: `1301` for OPNsense 24.1 or below, or `1400` for OPNsense 24.7 or above."},
			"pfsync_defer":      schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "When enabled, transmission of the first packet in a state is deferred until the peer acknowledges insertion."},
			"synchronize_to_ip": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Address or URL of the peer that receives selected configuration sections through XMLRPC synchronization. Leave empty on the backup node or when Terraform manages both peers independently."},
			"verify_peer":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "When enabled, verify TLS for the XMLRPC synchronization peer."},
			"username":          schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Username used for XMLRPC configuration synchronization."},
			"password":          schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, MarkdownDescription: "Password used for XMLRPC configuration synchronization."},
			"sync_items":        schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, MarkdownDescription: "OPNsense HA synchronization section identifiers to replicate through XMLRPC, such as `virtualip`, `rules`, or `nat`."},
			"id":                schema.StringAttribute{Computed: true, MarkdownDescription: "Fixed singleton identifier `core_hasync`.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func hasyncDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Reads OPNsense High Availability synchronization settings.",
		Attributes: map[string]dschema.Attribute{
			"disable_preempt":   dschema.BoolAttribute{Computed: true},
			"disconnect_ppps":   dschema.BoolAttribute{Computed: true},
			"pfsync_interface":  dschema.StringAttribute{Computed: true},
			"pfsync_peer_ip":    dschema.StringAttribute{Computed: true},
			"pfsync_version":    dschema.StringAttribute{Computed: true},
			"pfsync_defer":      dschema.BoolAttribute{Computed: true},
			"synchronize_to_ip": dschema.StringAttribute{Computed: true},
			"verify_peer":       dschema.BoolAttribute{Computed: true},
			"username":          dschema.StringAttribute{Computed: true},
			"password":          dschema.StringAttribute{Computed: true, Sensitive: true},
			"sync_items":        dschema.SetAttribute{Computed: true, ElementType: types.StringType},
			"id":                dschema.StringAttribute{Computed: true},
		},
	}
}

func hasyncAPIToModel(d *apicore.HasyncSettings) *hasyncModel {
	return &hasyncModel{
		DisablePreempt:  types.BoolValue(tools.StringToBool(d.DisablePreempt)),
		DisconnectPPPs:  types.BoolValue(tools.StringToBool(d.DisconnectPPPs)),
		PfsyncInterface: types.StringValue(d.PfsyncInterface.String()),
		PfsyncPeerIP:    tools.StringOrNull(d.PfsyncPeerIP),
		PfsyncVersion:   types.StringValue(d.PfsyncVersion.String()),
		PfsyncDefer:     types.BoolValue(tools.StringToBool(d.PfsyncDefer)),
		SynchronizeToIP: tools.StringOrNull(d.SynchronizeToIP),
		VerifyPeer:      types.BoolValue(tools.StringToBool(d.VerifyPeer)),
		Username:        tools.StringOrNull(d.Username),
		Password:        tools.StringOrNull(d.Password),
		SyncItems:       tools.StringSliceToSet([]string(d.SyncItems)),
		ID:              types.StringValue(hasyncID),
	}
}

func applyHasyncModel(dst *apicore.HasyncSettings, d *hasyncModel) {
	dst.DisablePreempt = tools.BoolToString(d.DisablePreempt.ValueBool())
	dst.DisconnectPPPs = tools.BoolToString(d.DisconnectPPPs.ValueBool())
	dst.PfsyncInterface = api.SelectedMap(d.PfsyncInterface.ValueString())
	dst.PfsyncPeerIP = d.PfsyncPeerIP.ValueString()
	dst.PfsyncVersion = api.SelectedMap(d.PfsyncVersion.ValueString())
	dst.PfsyncDefer = tools.BoolToString(d.PfsyncDefer.ValueBool())
	dst.SynchronizeToIP = d.SynchronizeToIP.ValueString()
	dst.VerifyPeer = tools.BoolToString(d.VerifyPeer.ValueBool())
	dst.Username = d.Username.ValueString()
	if !d.Password.IsNull() && !d.Password.IsUnknown() {
		dst.Password = d.Password.ValueString()
	}
	dst.SyncItems = api.SelectedMapList(tools.SetToStringSlice(d.SyncItems))
}
