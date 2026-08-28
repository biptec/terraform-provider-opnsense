package quagga

import (
	"context"
	"fmt"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/opnsense-go/pkg/quagga"
)

type ospfRouteMapReferenceRow struct {
	UUID       string              `json:"uuid"`
	PrefixList api.SelectedMapList `json:"match2"`
}

func withoutSelected(values api.SelectedMapList, id string) (api.SelectedMapList, bool) {
	out := make(api.SelectedMapList, 0, len(values))
	changed := false
	for _, value := range values {
		if value == id {
			changed = true
			continue
		}
		out = append(out, value)
	}
	return out, changed
}

func unlinkOSPFPrefixListFromRouteMaps(ctx context.Context, client opnsense.Client, id string) error {
	rows, err := api.Search[ospfRouteMapReferenceRow](client.Quagga().Client(), ctx, quagga.OSPFRouteMapOpts.Search)
	if err != nil {
		return fmt.Errorf("search OSPFv2 route-maps before prefix-list delete: %w", err)
	}
	for _, row := range rows.Rows {
		if row.UUID == "" {
			continue
		}
		if _, referenced := withoutSelected(row.PrefixList, id); !referenced {
			continue
		}
		routeMap, err := client.Quagga().GetOSPFRouteMap(ctx, row.UUID)
		if err != nil {
			return fmt.Errorf("read OSPFv2 route-map %s before prefix-list delete: %w", row.UUID, err)
		}
		routeMap.PrefixList, _ = withoutSelected(routeMap.PrefixList, id)
		if err := client.Quagga().UpdateOSPFRouteMap(ctx, row.UUID, routeMap); err != nil {
			return fmt.Errorf("unlink OSPFv2 prefix-list %s from route-map %s: %w", id, row.UUID, err)
		}
	}
	return nil
}

func unlinkOSPF6PrefixListFromRouteMaps(ctx context.Context, client opnsense.Client, id string) error {
	rows, err := api.Search[ospfRouteMapReferenceRow](client.Quagga().Client(), ctx, quagga.OSPF6RouteMapOpts.Search)
	if err != nil {
		return fmt.Errorf("search OSPFv3 route-maps before prefix-list delete: %w", err)
	}
	for _, row := range rows.Rows {
		if row.UUID == "" {
			continue
		}
		if _, referenced := withoutSelected(row.PrefixList, id); !referenced {
			continue
		}
		routeMap, err := client.Quagga().GetOSPF6RouteMap(ctx, row.UUID)
		if err != nil {
			return fmt.Errorf("read OSPFv3 route-map %s before prefix-list delete: %w", row.UUID, err)
		}
		routeMap.PrefixList, _ = withoutSelected(routeMap.PrefixList, id)
		if err := client.Quagga().UpdateOSPF6RouteMap(ctx, row.UUID, routeMap); err != nil {
			return fmt.Errorf("unlink OSPFv3 prefix-list %s from route-map %s: %w", id, row.UUID, err)
		}
	}
	return nil
}
