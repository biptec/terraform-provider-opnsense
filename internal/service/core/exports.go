package core

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{newHasyncResource}
}

func DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{newHasyncDataSource}
}
