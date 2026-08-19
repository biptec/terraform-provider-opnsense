package haproxy

import "fmt"

type namedRemoteItem struct {
	ID   string
	Name string
}

func validateUniqueRemoteName(kind, name, ownID string, items []namedRemoteItem) error {
	for _, item := range items {
		if item.Name == name && item.ID != ownID {
			return fmt.Errorf("%s name %q is already used by remote object %s", kind, name, item.ID)
		}
	}
	return nil
}
