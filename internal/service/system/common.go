package system

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func configureClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) opnsense.Client {
	if req.ProviderData == nil {
		return nil
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData),
		)
		return nil
	}
	return opnsense.NewClient(client)
}

func stringSet(ctx context.Context, value types.Set) ([]string, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	var result []string
	diagnostics := value.ElementsAs(ctx, &result, false)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("unable to decode string set: %v", diagnostics.Errors())
	}
	sort.Strings(result)
	return result, nil
}

func stringSetValue(values []string) types.Set {
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	items := make([]attr.Value, 0, len(sorted))
	for _, value := range sorted {
		items = append(items, types.StringValue(value))
	}
	return types.SetValueMust(types.StringType, items)
}

func sameStrings(left, right []string) bool {
	leftCopy := append([]string(nil), left...)
	rightCopy := append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	return strings.Join(leftCopy, "\x00") == strings.Join(rightCopy, "\x00")
}

func validationMessage(validations map[string]string) string {
	if len(validations) == 0 {
		return ""
	}
	keys := make([]string, 0, len(validations))
	for key := range validations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", key, validations[key]))
	}
	return strings.Join(parts, "; ")
}
