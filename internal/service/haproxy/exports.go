package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newSettingsResource,
		newFrontendResource,
		newBackendResource,
		newServerResource,
		newHealthcheckResource,
		newACLResource,
		newActionResource,
	}
}

func DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newSettingsDataSource,
		newStatusDataSource,
		newConfigtestDataSource,
		newFrontendDataSource,
		newBackendDataSource,
		newServerDataSource,
		newHealthcheckDataSource,
		newACLDataSource,
		newActionDataSource,
	}
}
