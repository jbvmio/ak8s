package ak8s

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Pod Constants:
const (
	PodAPIVersion = `v1`
	PodListKind   = `List`
	PodKind       = `Pod`
)

// PodCollection Contains a Collection of Pods.
type PodCollection struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.PodList
}

// Pod contains a v1.Pod resource.
type Pod struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.Pod
}

// GetAllPods returns All Pods for the current namespace set on the client or all Pods across all namespaces if not set.
func (c *Client) GetAllPods() (*PodCollection, error) {
	list, err := c.CS.CoreV1().Pods(c.NS).List(c.Options[ListOption].(*ListAction).Get())
	if err != nil {
		return &PodCollection{}, err
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Version: PodAPIVersion,
		Kind:    PodListKind,
	})
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Version: PodAPIVersion,
			Kind:    PodKind,
		})
	}
	return &PodCollection{
		list.APIVersion,
		list.Kind,
		list,
	}, nil
}

// GetPods returns Pods for the given namespaces.
// If the namespace is not set on the client, the "default" namespace is used.
func (c *Client) GetPods(names ...string) (*PodCollection, error) {
	if len(names) < 1 {
		return &PodCollection{}, fmt.Errorf("no pods specified")
	}
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	var pods []v1.Pod
	var list v1.PodList
	var errd string
	var Err error
	for _, name := range names {
		p, err := c.CS.CoreV1().Pods(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
		if err != nil {
			errd += (err.Error() + fmt.Sprintf("\n"))
		} else {
			pods = append(pods, *p)
		}
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Version: PodAPIVersion,
		Kind:    PodListKind,
	})
	switch {
	case len(pods) < 1 && errd != "":
		return &PodCollection{}, fmt.Errorf("%v", errd)
	case len(pods) < 1 && errd == "":
		return &PodCollection{}, fmt.Errorf("%d pods not found", len(names))
	case len(pods) > 0 && errd != "":
		Err = fmt.Errorf("%v", errd)
	}
	list.Items = pods
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Version: PodAPIVersion,
			Kind:    PodKind,
		})
	}
	return &PodCollection{
		list.APIVersion,
		list.Kind,
		&list,
	}, Err
}

// GetPod returns the Pod for the given pod name.
func (c *Client) GetPod(name string) (*Pod, error) {
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	p, err := c.CS.CoreV1().Pods(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
	if err != nil {
		return &Pod{}, err
	}
	p.SetGroupVersionKind(schema.GroupVersionKind{
		Version: PodAPIVersion,
		Kind:    PodKind,
	})
	return &Pod{
		p.APIVersion,
		p.Kind,
		p,
	}, nil
}

// GetNames returns all item names contained within the Collection.
func (c *PodCollection) GetNames() []string {
	var names []string
	for _, pod := range c.Items {
		names = append(names, pod.Name)
	}
	return names
}

// GetKind returns the collection kind.
func (c *PodCollection) GetKind() string {
	return c.Kind
}

// Len returns the number of items in the collection.
func (c *PodCollection) Len() int {
	return len(c.Items)
}

// Get returns an item by name.
func (c *PodCollection) Get(name string) Pod {
	for _, item := range c.Items {
		if item.Name == name {
			return Pod{
				PodAPIVersion,
				PodKind,
				&item,
			}
		}
	}
	return Pod{}
}

// Search conducts a wildcard search by names and returns matching items.
func (c *PodCollection) Search(names ...string) (collection *PodCollection) {
	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(fmt.Sprintf("%s", r), `regexp`) {
				collection = c.podSearchBak(names...)
			} else {
				panic(r)
			}
		}
	}()
	var list v1.PodList
	var regex *regexp.Regexp
	var regexString string
	var matches []v1.Pod
	switch {
	case len(names) <= 0:
		return c
	case len(names) == 1:
		regexString = names[0]
		regex = regexp.MustCompile(regexString)
		for _, pod := range c.Items {
			if regex.MatchString(pod.Name) {
				matches = append(matches, pod)
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
		for _, pod := range c.Items {
			if regex.MatchString(pod.Name) {
				matches = append(matches, pod)
			}
		}
	}
	list.Items = matches
	collection = &PodCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
	return
}

func (c *PodCollection) podSearchBak(names ...string) *PodCollection {
	var list v1.PodList
	var matches []v1.Pod
	for _, name := range names {
		for _, item := range c.Items {
			if strings.Contains(item.Name, name) {
				matches = append(matches, item)
			}
		}
	}
	list.Items = matches
	return &PodCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
}

// GetName returns the name of the resource.
func (r *Pod) GetName() string {
	return r.Name
}

// GetUID returns the UID of the resource.
func (r *Pod) GetUID() types.UID {
	return r.UID
}
