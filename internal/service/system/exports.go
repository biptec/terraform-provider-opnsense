package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newPluginResource,
		newWebguiResource,
		newSshResource,
		newNtpSettingsResource,
		newCarpHealthResource,
		newCarpHealthCheckResource,
		newInterfaceSyncPolicyResource,
		newInterfacePolicyAssignmentResource,
	}
}

func DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newCarpHealthStatusDataSource,
	}
}
