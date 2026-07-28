package interfaces

import "github.com/hashicorp/terraform-plugin-framework/types"

func typesString(value string) types.String { return types.StringValue(value) }
