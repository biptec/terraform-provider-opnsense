package provider

import (
	"context"
	"testing"
)

func TestNewProviderVersion(t *testing.T) {
	t.Parallel()

	instance, err := NewProvider(context.Background(), "1.2.3")
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	providerInstance, ok := instance.(*opnsenseProvider)
	if !ok {
		t.Fatalf("NewProvider() type = %T, want *opnsenseProvider", instance)
	}
	if providerInstance.version != "1.2.3" {
		t.Fatalf("provider version = %q, want %q", providerInstance.version, "1.2.3")
	}
}
