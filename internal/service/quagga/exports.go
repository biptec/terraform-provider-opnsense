package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newBGPASPathResource,
		newBGPCommunityListResource,
		newBGPNeighborResource,
		newBGPPrefixListResource,
		newBGPRouteMapResource,
		newFrrGeneralResource,
		newOspfSettingsResource,
		newOspfPrefixListResource,
		newOspfRouteMapResource,
		newOspfRedistributionResource,
		newOspfNetworkResource,
		newOspfInterfaceResource,
		newOspf6SettingsResource,
		newOspf6PrefixListResource,
		newOspf6RouteMapResource,
		newOspf6RedistributionResource,
		newOspf6NetworkResource,
		newOspf6InterfaceResource,
	}
}

func DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newBGPASPathDataSource,
		newBGPCommunityListDataSource,
		newBGPNeighborDataSource,
		newBGPPrefixListDataSource,
		newBGPRouteMapDataSource,
		newFrrGeneralDataSource,
		newOspfSettingsDataSource,
		newOspfPrefixListDataSource,
		newOspfRouteMapDataSource,
		newOspfRedistributionDataSource,
		newOspfNetworkDataSource,
		newOspfInterfaceDataSource,
		newOspf6SettingsDataSource,
		newOspf6PrefixListDataSource,
		newOspf6RouteMapDataSource,
		newOspf6RedistributionDataSource,
		newOspf6NetworkDataSource,
		newOspf6InterfaceDataSource,
	}
}
