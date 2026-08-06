package tools

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetToStringSliceSortsValues(t *testing.T) {
	set := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("192.0.2.53"),
		types.StringValue("10.200.210.2"),
		types.StringValue("198.51.100.53"),
	})

	want := []string{"10.200.210.2", "192.0.2.53", "198.51.100.53"}
	got := SetToStringSlice(set)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected sorted set: got %v, want %v", got, want)
	}
}

func TestSetToStringUsesSortedValues(t *testing.T) {
	set := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("beta"),
		types.StringValue("alpha"),
		types.StringValue("gamma"),
	})

	if got, want := SetToString(set, "\n"), "alpha\nbeta\ngamma"; got != want {
		t.Fatalf("unexpected joined set: got %q, want %q", got, want)
	}
}
