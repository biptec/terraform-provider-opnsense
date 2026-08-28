package core

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHasyncSyncIDIsDeterministic(t *testing.T) {
	left := hasyncSyncID([]string{"interface_vlans", "rules"}, true, false)
	right := hasyncSyncID([]string{"interface_vlans", "rules"}, true, false)
	if left != right || left == "" {
		t.Fatalf("unexpected deterministic ID: %q vs %q", left, right)
	}
	if left == hasyncSyncID([]string{"interface_vlans", "rules"}, false, true) {
		t.Fatal("create and delete barriers must have distinct IDs")
	}
}

func TestHasyncSyncItemsAreSorted(t *testing.T) {
	set, diagnostics := types.SetValueFrom(context.Background(), types.StringType, []string{"rules", "interface_vlans"})
	if diagnostics.HasError() {
		t.Fatalf("build test set: %v", diagnostics)
	}
	items, err := hasyncSyncItems(context.Background(), &hasyncSyncModel{Items: set})
	if err != nil {
		t.Fatalf("hasyncSyncItems() error = %v", err)
	}
	if len(items) != 2 || items[0] != "interface_vlans" || items[1] != "rules" {
		t.Fatalf("unexpected sorted items: %#v", items)
	}
}
func TestMergeHasyncSyncItemsPreservesRemovedCategories(t *testing.T) {
	merged := mergeHasyncSyncItems(
		[]string{"rules", "staticroutes"},
		[]string{"rules"},
	)
	if len(merged) != 2 || merged[0] != "rules" || merged[1] != "staticroutes" {
		t.Fatalf("unexpected merged HA sync items: %#v", merged)
	}
}
