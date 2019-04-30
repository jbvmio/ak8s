package ak8s

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Ingresses Constants:
const (
	IngressAPIGroup   = `extensions`
	IngressAPIVersion = `v1beta1`
	IngressListKind   = `List`
	IngressKind       = `Ingress`
)

// IngressCollection Contains a Collection of Ingresses.
type IngressCollection struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1beta1.IngressList
}

// Ingress contains a v1.Ingresses resource.
type Ingress struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1beta1.Ingress
}

// GetAllIngress returns All Ingresses for the current namespace set on the client or all Ingresses across all namespaces if not set.
func (c *Client) GetAllIngress() (*IngressCollection, error) {
	list, err := c.CS.ExtensionsV1beta1().Ingresses(c.NS).List(c.Options[ListOption].(*ListAction).Get())
	if err != nil {
		return &IngressCollection{}, err
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   IngressAPIGroup,
		Version: IngressAPIVersion,
		Kind:    IngressListKind,
	})
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Group:   IngressAPIGroup,
			Version: IngressAPIVersion,
			Kind:    IngressKind,
		})
	}
	return &IngressCollection{
		list.APIVersion,
		list.Kind,
		list,
	}, nil
}

// GetIngresses returns Ingresses for the given namespaces.
// If the namespace is not set on the client, the "default" namespace is used.
func (c *Client) GetIngresses(names ...string) (*IngressCollection, error) {
	if len(names) < 1 {
		return &IngressCollection{}, fmt.Errorf("no Ingresses specified")
	}
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	var Ingresses []v1beta1.Ingress
	var list v1beta1.IngressList
	var errd string
	var Err error
	for _, name := range names {
		p, err := c.CS.ExtensionsV1beta1().Ingresses(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
		if err != nil {
			errd += (err.Error() + fmt.Sprintf("\n"))
		} else {
			Ingresses = append(Ingresses, *p)
		}
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   IngressAPIGroup,
		Version: IngressAPIVersion,
		Kind:    IngressListKind,
	})
	switch {
	case len(Ingresses) < 1 && errd != "":
		return &IngressCollection{}, fmt.Errorf("%v", errd)
	case len(Ingresses) < 1 && errd == "":
		return &IngressCollection{}, fmt.Errorf("%d Ingresses not found", len(names))
	case len(Ingresses) > 0 && errd != "":
		Err = fmt.Errorf("%v", errd)
	}
	list.Items = Ingresses
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Group:   IngressAPIGroup,
			Version: IngressAPIVersion,
			Kind:    IngressKind,
		})
	}
	return &IngressCollection{
		list.APIVersion,
		list.Kind,
		&list,
	}, Err
}

// GetIngress returns the Ingresses for the given name.
func (c *Client) GetIngress(name string) (*Ingress, error) {
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	p, err := c.CS.ExtensionsV1beta1().Ingresses(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
	if err != nil {
		return &Ingress{}, err
	}
	p.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   IngressAPIGroup,
		Version: IngressAPIVersion,
		Kind:    IngressKind,
	})
	return &Ingress{
		p.APIVersion,
		p.Kind,
		p,
	}, nil
}

// GetNames returns all item names contained within the Collection.
func (c *IngressCollection) GetNames() []string {
	var names []string
	for _, item := range c.Items {
		names = append(names, item.Name)
	}
	return names
}

// GetKind returns the collection kind.
func (c *IngressCollection) GetKind() string {
	return c.Kind
}

// Len returns the number of items in the collection.
func (c *IngressCollection) Len() int {
	return len(c.Items)
}

// Get returns an item by name.
func (c *IngressCollection) Get(name string) Ingress {
	for _, item := range c.Items {
		if item.Name == name {
			return Ingress{
				IngressAPIVersion,
				IngressKind,
				&item,
			}
		}
	}
	return Ingress{}
}

// Search conducts a wildcard search by names and returns matching items.
func (c *IngressCollection) Search(names ...string) (collection *IngressCollection) {
	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(fmt.Sprintf("%s", r), `regexp`) {
				collection = c.ingresssearchBak(names...)
			} else {
				panic(r)
			}
		}
	}()
	var list v1beta1.IngressList
	var regex *regexp.Regexp
	var regexString string
	var matches []v1beta1.Ingress
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
	collection = &IngressCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
	return
}

func (c *IngressCollection) ingresssearchBak(names ...string) *IngressCollection {
	var list v1beta1.IngressList
	var matches []v1beta1.Ingress
	for _, name := range names {
		for _, item := range c.Items {
			if strings.Contains(item.Name, name) {
				matches = append(matches, item)
			}
		}
	}
	list.Items = matches
	return &IngressCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
}

// GetName returns the name of the resource.
func (r *Ingress) GetName() string {
	return r.Name
}

// GetUID returns the UID of the resource.
func (r *Ingress) GetUID() types.UID {
	return r.UID
}
