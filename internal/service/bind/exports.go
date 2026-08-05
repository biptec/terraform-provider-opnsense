package bind

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newSettingsResource,
		newAclResource,
		newViewResource,
		newTsigKeyResource,
		newPrimaryDomainResource,
		newSecondaryDomainResource,
		newForwardDomainResource,
		newRecordResource,
	}
}

func DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newSettingsDataSource,
		newStatusDataSource,
		newAclDataSource,
		newViewDataSource,
		newTsigKeyDataSource,
		newPrimaryDomainDataSource,
		newSecondaryDomainDataSource,
		newForwardDomainDataSource,
		newRecordDataSource,
		newDNSSECStatusDataSource,
	}
}
