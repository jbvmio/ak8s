package ak8s

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// DaemonSet Constants:
const (
	DaemonSetAPIGroup   = `apps`
	DaemonSetAPIVersion = `v1`
	DaemonSetListKind   = `List`
	DaemonSetKind       = `DaemonSet`
)

// DaemonSetCollection Contains a Collection of DaemonSets.
type DaemonSetCollection struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.DaemonSetList
}

// DaemonSet contains a v1.DaemonSet resource.
type DaemonSet struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.DaemonSet
}

// GetAllDaemonSets returns All DaemonSets for the current namespace set on the client or all DaemonSets across all namespaces if not set.
func (c *Client) GetAllDaemonSets() (*DaemonSetCollection, error) {
	list, err := c.CS.AppsV1().DaemonSets(c.NS).List(c.Options[ListOption].(*ListAction).Get())
	if err != nil {
		return &DaemonSetCollection{}, err
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   DaemonSetAPIGroup,
		Version: DaemonSetAPIVersion,
		Kind:    DaemonSetListKind,
	})
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Group:   DaemonSetAPIGroup,
			Version: DaemonSetAPIVersion,
			Kind:    DaemonSetKind,
		})
	}
	return &DaemonSetCollection{
		list.APIVersion,
		list.Kind,
		list,
	}, nil
}

// GetDaemonSets returns DaemonSets for the given namespaces.
// If the namespace is not set on the client, the "default" namespace is used.
func (c *Client) GetDaemonSets(names ...string) (*DaemonSetCollection, error) {
	if len(names) < 1 {
		return &DaemonSetCollection{}, fmt.Errorf("no DaemonSets specified")
	}
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	var DaemonSet []v1.DaemonSet
	var list v1.DaemonSetList
	var errd string
	var Err error
	for _, name := range names {
		p, err := c.CS.AppsV1().DaemonSets(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
		if err != nil {
			errd += (err.Error() + fmt.Sprintf("\n"))
		} else {
			DaemonSet = append(DaemonSet, *p)
		}
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   DaemonSetAPIGroup,
		Version: DaemonSetAPIVersion,
		Kind:    DaemonSetListKind,
	})
	switch {
	case len(DaemonSet) < 1 && errd != "":
		return &DaemonSetCollection{}, fmt.Errorf("%v", errd)
	case len(DaemonSet) < 1 && errd == "":
		return &DaemonSetCollection{}, fmt.Errorf("%d DaemonSets not found", len(names))
	case len(DaemonSet) > 0 && errd != "":
		Err = fmt.Errorf("%v", errd)
	}
	list.Items = DaemonSet
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Group:   DaemonSetAPIGroup,
			Version: DaemonSetAPIVersion,
			Kind:    DaemonSetKind,
		})
	}
	return &DaemonSetCollection{
		list.APIVersion,
		list.Kind,
		&list,
	}, Err
}

// GetDaemonSet returns the DaemonSet for the given name.
func (c *Client) GetDaemonSet(name string) (*DaemonSet, error) {
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	p, err := c.CS.AppsV1().DaemonSets(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
	if err != nil {
		return &DaemonSet{}, err
	}
	p.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   DaemonSetAPIGroup,
		Version: DaemonSetAPIVersion,
		Kind:    DaemonSetKind,
	})
	return &DaemonSet{
		p.APIVersion,
		p.Kind,
		p,
	}, nil
}

// GetNames returns all item names contained within the Collection.
func (c *DaemonSetCollection) GetNames() []string {
	var names []string
	for _, item := range c.Items {
		names = append(names, item.Name)
	}
	return names
}

// GetKind returns the collection kind.
func (c *DaemonSetCollection) GetKind() string {
	return c.Kind
}

// Len returns the number of items in the collection.
func (c *DaemonSetCollection) Len() int {
	return len(c.Items)
}

// Get returns an item by name.
func (c *DaemonSetCollection) Get(name string) DaemonSet {
	for _, item := range c.Items {
		if item.Name == name {
			return DaemonSet{
				DaemonSetAPIVersion,
				DaemonSetKind,
				&item,
			}
		}
	}
	return DaemonSet{}
}

// Search conducts a wildcard search by names and returns matching items.
func (c *DaemonSetCollection) Search(names ...string) (collection *DaemonSetCollection) {
	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(fmt.Sprintf("%s", r), `regexp`) {
				collection = c.daemonSetsearchBak(names...)
			} else {
				panic(r)
			}
		}
	}()
	var list v1.DaemonSetList
	var regex *regexp.Regexp
	var regexString string
	var matches []v1.DaemonSet
	switch {
	case len(names) <= 0:
		return c
	case len(names) == 1:
		regexString = names[0]
		regex = regexp.MustCompile(regexString)
		for _, item := range c.Items {
			if regex.MatchString(item.Name) {
				matches = append(matches, item)
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
		for _, item := range c.Items {
			if regex.MatchString(item.Name) {
				matches = append(matches, item)
			}
		}
	}
	list.Items = matches
	collection = &DaemonSetCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
	return
}

func (c *DaemonSetCollection) daemonSetsearchBak(names ...string) *DaemonSetCollection {
	var list v1.DaemonSetList
	var matches []v1.DaemonSet
	for _, name := range names {
		for _, item := range c.Items {
			if strings.Contains(item.Name, name) {
				matches = append(matches, item)
			}
		}
	}
	list.Items = matches
	return &DaemonSetCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
}

// GetName returns the name of the resource.
func (r *DaemonSet) GetName() string {
	return r.Name
}

// GetUID returns the UID of the resource.
func (r *DaemonSet) GetUID() types.UID {
	return r.UID
}
