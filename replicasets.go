package ak8s

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// ReplicaSets Constants:
const (
	ReplicaSetAPIGroup   = `apps`
	ReplicaSetAPIVersion = `v1`
	ReplicaSetListKind   = `List`
	ReplicaSetKind       = `ReplicaSet`
)

// ReplicaSetCollection Contains a Collection of ReplicaSets.
type ReplicaSetCollection struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.ReplicaSetList
}

// ReplicaSet contains a v1.ReplicaSets resource.
type ReplicaSet struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.ReplicaSet
}

// GetAllReplicaSets returns All ReplicaSets for the current namespace set on the client or all ReplicaSets across all namespaces if not set.
func (c *Client) GetAllReplicaSets() (*ReplicaSetCollection, error) {
	//list, err := c.CS.ExtensionsV1beta1().ReplicaSets(c.NS).List(c.Options[ListOption].(*ListAction).Get())
	list, err := c.CS.AppsV1().ReplicaSets(c.NS).List(c.Options[ListOption].(*ListAction).Get())
	if err != nil {
		return &ReplicaSetCollection{}, err
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   ReplicaSetAPIGroup,
		Version: ReplicaSetAPIVersion,
		Kind:    ReplicaSetListKind,
	})
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Group:   ReplicaSetAPIGroup,
			Version: ReplicaSetAPIVersion,
			Kind:    ReplicaSetKind,
		})
	}
	return &ReplicaSetCollection{
		list.APIVersion,
		list.Kind,
		list,
	}, nil
}

// GetReplicaSets returns ReplicaSets for the given namespaces.
// If the namespace is not set on the client, the "default" namespace is used.
func (c *Client) GetReplicaSets(names ...string) (*ReplicaSetCollection, error) {
	if len(names) < 1 {
		return &ReplicaSetCollection{}, fmt.Errorf("no ReplicaSets specified")
	}
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	var ReplicaSets []v1.ReplicaSet
	var list v1.ReplicaSetList
	var errd string
	var Err error
	for _, name := range names {
		p, err := c.CS.AppsV1().ReplicaSets(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
		if err != nil {
			errd += (err.Error() + fmt.Sprintf("\n"))
		} else {
			ReplicaSets = append(ReplicaSets, *p)
		}
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   ReplicaSetAPIGroup,
		Version: ReplicaSetAPIVersion,
		Kind:    ReplicaSetListKind,
	})
	switch {
	case len(ReplicaSets) < 1 && errd != "":
		return &ReplicaSetCollection{}, fmt.Errorf("%v", errd)
	case len(ReplicaSets) < 1 && errd == "":
		return &ReplicaSetCollection{}, fmt.Errorf("%d ReplicaSets not found", len(names))
	case len(ReplicaSets) > 0 && errd != "":
		Err = fmt.Errorf("%v", errd)
	}
	list.Items = ReplicaSets
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Group:   ReplicaSetAPIGroup,
			Version: ReplicaSetAPIVersion,
			Kind:    ReplicaSetKind,
		})
	}
	return &ReplicaSetCollection{
		list.APIVersion,
		list.Kind,
		&list,
	}, Err
}

// GetReplicaSet returns the ReplicaSets for the given ReplicaSets name.
func (c *Client) GetReplicaSet(name string) (*ReplicaSet, error) {
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	p, err := c.CS.AppsV1().ReplicaSets(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
	if err != nil {
		return &ReplicaSet{}, err
	}
	p.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   ReplicaSetAPIGroup,
		Version: ReplicaSetAPIVersion,
		Kind:    ReplicaSetKind,
	})
	return &ReplicaSet{
		p.APIVersion,
		p.Kind,
		p,
	}, nil
}

// GetNames returns all item names contained within the Collection.
func (c *ReplicaSetCollection) GetNames() []string {
	var names []string
	for _, ReplicaSets := range c.Items {
		names = append(names, ReplicaSets.Name)
	}
	return names
}

// GetKind returns the collection kind.
func (c *ReplicaSetCollection) GetKind() string {
	return c.Kind
}

// Len returns the number of items in the collection.
func (c *ReplicaSetCollection) Len() int {
	return len(c.Items)
}

// Get returns an item by name.
func (c *ReplicaSetCollection) Get(name string) ReplicaSet {
	for _, item := range c.Items {
		if item.Name == name {
			return ReplicaSet{
				ReplicaSetAPIVersion,
				ReplicaSetKind,
				&item,
			}
		}
	}
	return ReplicaSet{}
}

// Search conducts a wildcard search by names and returns matching items.
func (c *ReplicaSetCollection) Search(names ...string) (collection *ReplicaSetCollection) {
	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(fmt.Sprintf("%s", r), `regexp`) {
				collection = c.replicaSetsearchBak(names...)
			} else {
				panic(r)
			}
		}
	}()
	var list v1.ReplicaSetList
	var regex *regexp.Regexp
	var regexString string
	var matches []v1.ReplicaSet
	switch {
	case len(names) <= 0:
		return c
	case len(names) == 1:
		regexString = names[0]
		regex = regexp.MustCompile(regexString)
		for _, ReplicaSets := range c.Items {
			if regex.MatchString(ReplicaSets.Name) {
				matches = append(matches, ReplicaSets)
			}
		}
	default:
		first := names[0]
		rest := names[1:]
		regexString += first
		for _, r := range rest {
			regexString += `|` + r
		}
		regex = regexp.MustCompile(regexString)
		for _, ReplicaSets := range c.Items {
			if regex.MatchString(ReplicaSets.Name) {
				matches = append(matches, ReplicaSets)
			}
		}
	}
	list.Items = matches
	collection = &ReplicaSetCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
	return
}

func (c *ReplicaSetCollection) replicaSetsearchBak(names ...string) *ReplicaSetCollection {
	var list v1.ReplicaSetList
	var matches []v1.ReplicaSet
	for _, name := range names {
		for _, item := range c.Items {
			if strings.Contains(item.Name, name) {
				matches = append(matches, item)
			}
		}
	}
	list.Items = matches
	return &ReplicaSetCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
}

// GetName returns the name of the resource.
func (r *ReplicaSet) GetName() string {
	return r.Name
}

// GetUID returns the UID of the resource.
func (r *ReplicaSet) GetUID() types.UID {
	return r.UID
}
