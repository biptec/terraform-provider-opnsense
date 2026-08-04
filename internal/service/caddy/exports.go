package caddy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newSettingsResource,
		newDomainResource,
		newHandlerResource,
		newAccessListResource,
		newHeaderResource,
	}
}

func DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newSettingsDataSource,
		newDomainDataSource,
		newHandlerDataSource,
		newAccessListDataSource,
		newHeaderDataSource,
		newStatusDataSource,
	}
}
