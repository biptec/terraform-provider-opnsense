package haproxy

import (
	"context"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func optionalSelectedMapList(value types.Set) api.SelectedMapList {
	if value.ElementType(context.Background()) == nil || value.IsNull() || value.IsUnknown() {
		return nil
	}
	return api.SelectedMapList(tools.SetToStringSlice(value))
}
