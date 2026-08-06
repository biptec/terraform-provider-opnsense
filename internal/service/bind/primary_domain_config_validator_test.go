package bind

import (
	"testing"

	"github.com/biptec/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidatePrimaryDomainTransfer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		transferKeyID types.String
		alsoNotify    types.Set
		wantError     bool
	}{
		"empty configuration": {
			transferKeyID: types.StringValue(""),
			alsoNotify:    tools.StringSliceToSet(nil),
		},
		"transfer key only": {
			transferKeyID: types.StringValue("11111111-1111-4111-8111-111111111111"),
			alsoNotify:    tools.StringSliceToSet(nil),
		},
		"authenticated notify": {
			transferKeyID: types.StringValue("11111111-1111-4111-8111-111111111111"),
			alsoNotify:    tools.StringSliceToSet([]string{"192.0.2.54"}),
		},
		"notify without key": {
			transferKeyID: types.StringValue(""),
			alsoNotify:    tools.StringSliceToSet([]string{"192.0.2.54"}),
			wantError:     true,
		},
		"unknown key during planning": {
			transferKeyID: types.StringUnknown(),
			alsoNotify:    tools.StringSliceToSet([]string{"192.0.2.54"}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := validatePrimaryDomainTransfer(test.transferKeyID, test.alsoNotify)
			if (err != nil) != test.wantError {
				t.Fatalf("validatePrimaryDomainTransfer() error = %v, wantError %t", err, test.wantError)
			}
		})
	}
}
