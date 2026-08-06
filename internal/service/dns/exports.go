package dns

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newServiceCutoverResource,
	}
}
