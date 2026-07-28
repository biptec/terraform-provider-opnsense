package interfaces

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newAssignmentResource,
		newInterfaceSettingsResource,
		newVipResource,
		newVlanResource,
	}
}

func DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newAssignmentDataSource,
		newInterfaceSettingsDataSource,
		newVipDataSource,
		newVlanDataSource,
		newOverviewInterfaceDataSource,
		newOverviewAllDataSource,
	}
}
