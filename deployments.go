package ak8s

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Deployment Constants:
const (
	DeploymentAPIGroup   = `apps`
	DeploymentAPIVersion = `v1`
	DeploymentListKind   = `List`
	DeploymentKind       = `Deployment`
)

// DeployomentCollection Contains a Collection of Deployments.
type DeployomentCollection struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.DeploymentList
}

// DepCollection Contains a Collection of Deployments.
type DepCollection struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.DeploymentList
}

// Deployment contains a v1.Deployment resource.
type Deployment struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.Deployment
}

// GetAllDeployments returns All Deployments for the current namespace set on the client or all Deployments across all namespaces if not set.
func (c *Client) GetAllDeployments() (*DeployomentCollection, error) {
	list, err := c.CS.AppsV1().Deployments(c.NS).List(c.Options[ListOption].(*ListAction).Get())
	if err != nil {
		return &DeployomentCollection{}, err
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   DeploymentAPIGroup,
		Version: DeploymentAPIVersion,
		Kind:    DeploymentListKind,
	})
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Group:   DeploymentAPIGroup,
			Version: DeploymentAPIVersion,
			Kind:    DeploymentKind,
		})
	}
	return &DeployomentCollection{
		list.APIVersion,
		list.Kind,
		list,
	}, nil
}

// GetDeployments returns Deployments for the given namespaces.
// If the namespace is not set on the client, the "default" namespace is used.
func (c *Client) GetDeployments(names ...string) (*DeployomentCollection, error) {
	if len(names) < 1 {
		return &DeployomentCollection{}, fmt.Errorf("no Deployments specified")
	}
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	var deployment []v1.Deployment
	var list v1.DeploymentList
	var errd string
	var Err error
	for _, name := range names {
		p, err := c.CS.AppsV1().Deployments(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
		if err != nil {
			errd += (err.Error() + fmt.Sprintf("\n"))
		} else {
			deployment = append(deployment, *p)
		}
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   DeploymentAPIGroup,
		Version: DeploymentAPIVersion,
		Kind:    DeploymentListKind,
	})
	switch {
	case len(deployment) < 1 && errd != "":
		return &DeployomentCollection{}, fmt.Errorf("%v", errd)
	case len(deployment) < 1 && errd == "":
		return &DeployomentCollection{}, fmt.Errorf("%d deployments not found", len(names))
	case len(deployment) > 0 && errd != "":
		Err = fmt.Errorf("%v", errd)
	}
	list.Items = deployment
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Group:   DeploymentAPIGroup,
			Version: DeploymentAPIVersion,
			Kind:    DeploymentKind,
		})
	}
	return &DeployomentCollection{
		list.APIVersion,
		list.Kind,
		&list,
	}, Err
}

// GetDeployment returns the deployment for the given name.
func (c *Client) GetDeployment(name string) (*Deployment, error) {
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	p, err := c.CS.AppsV1().Deployments(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
	if err != nil {
		return &Deployment{}, err
	}
	p.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   DeploymentAPIGroup,
		Version: DeploymentAPIVersion,
		Kind:    DeploymentKind,
	})
	return &Deployment{
		p.APIVersion,
		p.Kind,
		p,
	}, nil
}

// GetNames returns all item names contained within the Collection.
func (c *DeployomentCollection) GetNames() []string {
	var names []string
	for _, item := range c.Items {
		names = append(names, item.Name)
	}
	return names
}

// GetKind returns the collection kind.
func (c *DeployomentCollection) GetKind() string {
	return c.Kind
}

// Len returns the number of items in the collection.
func (c *DeployomentCollection) Len() int {
	return len(c.Items)
}

// Get returns an item by name.
func (c *DeployomentCollection) Get(name string) Deployment {
	for _, item := range c.Items {
		if item.Name == name {
			return Deployment{
				DeploymentAPIVersion,
				DeploymentKind,
				&item,
			}
		}
	}
	return Deployment{}
}

// Search conducts a wildcard search by names and returns matching items.
func (c *DeployomentCollection) Search(names ...string) (collection *DeployomentCollection) {
	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(fmt.Sprintf("%s", r), `regexp`) {
				collection = c.deploymentsearchBak(names...)
			} else {
				panic(r)
			}
		}
	}()
	var list v1.DeploymentList
	var regex *regexp.Regexp
	var regexString string
	var matches []v1.Deployment
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
	collection = &DeployomentCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
	return
}

func (c *DeployomentCollection) deploymentsearchBak(names ...string) *DeployomentCollection {
	var list v1.DeploymentList
	var matches []v1.Deployment
	for _, name := range names {
		for _, item := range c.Items {
			if strings.Contains(item.Name, name) {
				matches = append(matches, item)
			}
		}
	}
	list.Items = matches
	return &DeployomentCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
}

// GetName returns the name of the resource.
func (r *Deployment) GetName() string {
	return r.Name
}

// GetUID returns the UID of the resource.
func (r *Deployment) GetUID() types.UID {
	return r.UID
}
