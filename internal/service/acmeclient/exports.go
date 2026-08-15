package acmeclient

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type resourceClient struct{ client opnsense.Client }

func (c *resourceClient) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	c.client = opnsense.NewClient(apiClient)
}

func Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newSettingsResource,
		newAccountResource,
		newValidationResource,
		newActionResource,
		newCertificateResource,
	}
}
