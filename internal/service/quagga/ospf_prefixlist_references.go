package quagga

import (
	"context"
	"fmt"
	"sync"

	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/opnsense"
	"github.com/biptec/opnsense-go/pkg/quagga"
)

var (
	ospfRouteMapMutationMu  sync.Mutex
	ospf6RouteMapMutationMu sync.Mutex
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
	// A prefix-list delete is a read-modify-write of a shared route-map. Terraform
	// may delete several prefix lists concurrently, so serialize the full unlink
	// operation to prevent a stale route-map write from reintroducing an ID that
	// another goroutine has already removed.
	ospfRouteMapMutationMu.Lock()
	defer ospfRouteMapMutationMu.Unlock()

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
	// OSPFv3 route-maps have the same shared read-modify-write hazard as OSPFv2.
	ospf6RouteMapMutationMu.Lock()
	defer ospf6RouteMapMutationMu.Unlock()

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
