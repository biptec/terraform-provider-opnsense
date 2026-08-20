package bind

import (
	"context"
	"fmt"
	"strings"

	apibind "github.com/biptec/opnsense-go/pkg/bind"
)

func normalizeBindViewName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeBindDomainName(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

type resolvedBindView struct {
	ID   string
	Name string
}

func selectBindViewByName(rows []apibind.View, requested string) (resolvedBindView, error) {
	target := normalizeBindViewName(requested)
	matches := make([]apibind.View, 0, 1)
	for _, row := range rows {
		if normalizeBindViewName(row.Name) == target {
			matches = append(matches, row)
		}
	}
	switch len(matches) {
	case 0:
		return resolvedBindView{}, fmt.Errorf("BIND view %q was not found", target)
	case 1:
		if matches[0].UUID == "" {
			return resolvedBindView{}, fmt.Errorf("BIND view %q search result did not include a UUID", target)
		}
		return resolvedBindView{ID: matches[0].UUID, Name: normalizeBindViewName(matches[0].Name)}, nil
	default:
		return resolvedBindView{}, fmt.Errorf("multiple BIND views named %q were found", target)
	}
}

type resolvedPrimaryDomain struct {
	ID     string
	Domain *apibind.PrimaryDomain
}

func selectPrimaryDomainInView(
	ctx context.Context,
	rows []apibind.PrimaryDomain,
	requestedDomain string,
	viewID string,
	viewName string,
	get func(context.Context, string) (*apibind.PrimaryDomain, error),
) (resolvedPrimaryDomain, error) {
	target := normalizeBindDomainName(requestedDomain)
	matches := make([]resolvedPrimaryDomain, 0, 1)
	for _, row := range rows {
		if normalizeBindDomainName(row.DomainName) != target {
			continue
		}
		if row.UUID == "" {
			return resolvedPrimaryDomain{}, fmt.Errorf("BIND primary domain %q search result did not include a UUID", target)
		}
		remote, err := get(ctx, row.UUID)
		if err != nil {
			return resolvedPrimaryDomain{}, fmt.Errorf("read BIND primary domain %q (%s): %w", target, row.UUID, err)
		}
		if remote.View.String() == viewID {
			matches = append(matches, resolvedPrimaryDomain{ID: row.UUID, Domain: remote})
		}
	}

	viewLabel := strings.TrimSpace(viewName)
	if viewLabel == "" {
		viewLabel = viewID
	}
	switch len(matches) {
	case 0:
		return resolvedPrimaryDomain{}, fmt.Errorf("BIND primary domain %q was not found in view %q", target, viewLabel)
	case 1:
		return matches[0], nil
	default:
		return resolvedPrimaryDomain{}, fmt.Errorf("multiple BIND primary domains named %q were found in view %q", target, viewLabel)
	}
}
