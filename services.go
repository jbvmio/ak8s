package ak8s

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Service Constants:
const (
	ServiceAPIVersion = `v1`
	ServiceListKind   = `List`
	ServiceKind       = `Service`
)

// ServiceCollection Contains a Collection of Services.
type ServiceCollection struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.ServiceList
}

// Service contains a v1.Service resource.
type Service struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.Service
}

// GetAllServices returns All Services for the current namespace set on the client or all Services across all namespaces if not set.
func (c *Client) GetAllServices() (*ServiceCollection, error) {
	list, err := c.CS.CoreV1().Services(c.NS).List(c.Options[ListOption].(*ListAction).Get())
	if err != nil {
		return &ServiceCollection{}, err
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Version: ServiceAPIVersion,
		Kind:    ServiceListKind,
	})
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Version: ServiceAPIVersion,
			Kind:    ServiceKind,
		})
	}
	return &ServiceCollection{
		list.APIVersion,
		list.Kind,
		list,
	}, nil
}

// GetServices returns Services for the given namespaces.
// If the namespace is not set on the client, the "default" namespace is used.
func (c *Client) GetServices(names ...string) (*ServiceCollection, error) {
	if len(names) < 1 {
		return &ServiceCollection{}, fmt.Errorf("no Services specified")
	}
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	var Services []v1.Service
	var list v1.ServiceList
	var errd string
	var Err error
	for _, name := range names {
		p, err := c.CS.CoreV1().Services(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
		if err != nil {
			errd += (err.Error() + fmt.Sprintf("\n"))
		} else {
			Services = append(Services, *p)
		}
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Version: ServiceAPIVersion,
		Kind:    ServiceListKind,
	})
	switch {
	case len(Services) < 1 && errd != "":
		return &ServiceCollection{}, fmt.Errorf("%v", errd)
	case len(Services) < 1 && errd == "":
		return &ServiceCollection{}, fmt.Errorf("%d Services not found", len(names))
	case len(Services) > 0 && errd != "":
		Err = fmt.Errorf("%v", errd)
	}
	list.Items = Services
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Version: ServiceAPIVersion,
			Kind:    ServiceKind,
		})
	}
	return &ServiceCollection{
		list.APIVersion,
		list.Kind,
		&list,
	}, Err
}

// GetService returns the Service for the given Service name.
func (c *Client) GetService(name string) (*Service, error) {
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	p, err := c.CS.CoreV1().Services(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
	if err != nil {
		return &Service{}, err
	}
	p.SetGroupVersionKind(schema.GroupVersionKind{
		Version: ServiceAPIVersion,
		Kind:    ServiceKind,
	})
	return &Service{
		p.APIVersion,
		p.Kind,
		p,
	}, nil
}

// GetNames returns all item names contained within the Collection.
func (c *ServiceCollection) GetNames() []string {
	var names []string
	for _, Service := range c.Items {
		names = append(names, Service.Name)
	}
	return names
}

// GetKind returns the collection kind.
func (c *ServiceCollection) GetKind() string {
	return c.Kind
}

// Len returns the number of items in the collection.
func (c *ServiceCollection) Len() int {
	return len(c.Items)
}

// Get returns an item by name.
func (c *ServiceCollection) Get(name string) Service {
	for _, item := range c.Items {
		if item.Name == name {
			return Service{
				ServiceAPIVersion,
				ServiceKind,
				&item,
			}
		}
	}
	return Service{}
}

// Search conducts a wildcard search by names and returns matching items.
func (c *ServiceCollection) Search(names ...string) (collection *ServiceCollection) {
	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(fmt.Sprintf("%s", r), `regexp`) {
				collection = c.serviceSearchBak(names...)
			} else {
				panic(r)
			}
		}
	}()
	var list v1.ServiceList
	var regex *regexp.Regexp
	var regexString string
	var matches []v1.Service
	switch {
	case len(names) <= 0:
		return c
	case len(names) == 1:
		regexString = names[0]
		regex = regexp.MustCompile(regexString)
		for _, Service := range c.Items {
			if regex.MatchString(Service.Name) {
				matches = append(matches, Service)
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
		for _, Service := range c.Items {
			if regex.MatchString(Service.Name) {
				matches = append(matches, Service)
			}
		}
	}
	list.Items = matches
	collection = &ServiceCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
	return
}

func (c *ServiceCollection) serviceSearchBak(names ...string) *ServiceCollection {
	var list v1.ServiceList
	var matches []v1.Service
	for _, name := range names {
		for _, item := range c.Items {
			if strings.Contains(item.Name, name) {
				matches = append(matches, item)
			}
		}
	}
	list.Items = matches
	return &ServiceCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
}

// GetName returns the name of the resource.
func (r *Service) GetName() string {
	return r.Name
}

// GetUID returns the UID of the resource.
func (r *Service) GetUID() types.UID {
	return r.UID
}
