package ak8s

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Node Constants:
const (
	NodeAPIVersion = `v1`
	NodeListKind   = `List`
	NodeKind       = `Node`
)

// NodeCollection Contains a Collection of Nodes.
type NodeCollection struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.NodeList
}

// Node contains a v1.Node resource.
type Node struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.Node
}

// GetAllNodes returns All Nodes for the current namespace set on the client or all Nodes across all namespaces if not set.
func (c *Client) GetAllNodes() (*NodeCollection, error) {
	list, err := c.CS.CoreV1().Nodes().List(c.Options[ListOption].(*ListAction).Get())
	if err != nil {
		return &NodeCollection{}, err
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Version: NodeAPIVersion,
		Kind:    NodeListKind,
	})
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Version: NodeAPIVersion,
			Kind:    NodeKind,
		})
	}
	return &NodeCollection{
		list.APIVersion,
		list.Kind,
		list,
	}, nil
}

// GetNodes returns Nodes for the given namespaces.
// If the namespace is not set on the client, the "default" namespace is used.
func (c *Client) GetNodes(names ...string) (*NodeCollection, error) {
	if len(names) < 1 {
		return &NodeCollection{}, fmt.Errorf("no Nodes specified")
	}
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	var Nodes []v1.Node
	var list v1.NodeList
	var errd string
	var Err error
	for _, name := range names {
		p, err := c.CS.CoreV1().Nodes().Get(name, c.Options[GetOption].(*GetAction).Get())
		if err != nil {
			errd += (err.Error() + fmt.Sprintf("\n"))
		} else {
			Nodes = append(Nodes, *p)
		}
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Version: NodeAPIVersion,
		Kind:    NodeListKind,
	})
	switch {
	case len(Nodes) < 1 && errd != "":
		return &NodeCollection{}, fmt.Errorf("%v", errd)
	case len(Nodes) < 1 && errd == "":
		return &NodeCollection{}, fmt.Errorf("%d Nodes not found", len(names))
	case len(Nodes) > 0 && errd != "":
		Err = fmt.Errorf("%v", errd)
	}
	list.Items = Nodes
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Version: NodeAPIVersion,
			Kind:    NodeKind,
		})
	}
	return &NodeCollection{
		list.APIVersion,
		list.Kind,
		&list,
	}, Err
}

// GetNode returns the Node for the given Node name.
func (c *Client) GetNode(name string) (*Node, error) {
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	p, err := c.CS.CoreV1().Nodes().Get(name, c.Options[GetOption].(*GetAction).Get())
	if err != nil {
		return &Node{}, err
	}
	p.SetGroupVersionKind(schema.GroupVersionKind{
		Version: NodeAPIVersion,
		Kind:    NodeKind,
	})
	return &Node{
		p.APIVersion,
		p.Kind,
		p,
	}, nil
}

// GetNames returns all item names contained within the Collection.
func (c *NodeCollection) GetNames() []string {
	var names []string
	for _, Node := range c.Items {
		names = append(names, Node.Name)
	}
	return names
}

// GetKind returns the collection kind.
func (c *NodeCollection) GetKind() string {
	return c.Kind
}

// Len returns the number of items in the collection.
func (c *NodeCollection) Len() int {
	return len(c.Items)
}

// Get returns an item by name.
func (c *NodeCollection) Get(name string) Node {
	for _, item := range c.Items {
		if item.Name == name {
			return Node{
				NodeAPIVersion,
				NodeKind,
				&item,
			}
		}
	}
	return Node{}
}

// Search conducts a wildcard search by names and returns matching items.
func (c *NodeCollection) Search(names ...string) (collection *NodeCollection) {
	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(fmt.Sprintf("%s", r), `regexp`) {
				collection = c.nodeSearchBak(names...)
			} else {
				panic(r)
			}
		}
	}()
	var list v1.NodeList
	var regex *regexp.Regexp
	var regexString string
	var matches []v1.Node
	switch {
	case len(names) <= 0:
		return c
	case len(names) == 1:
		regexString = names[0]
		regex = regexp.MustCompile(regexString)
		for _, Node := range c.Items {
			if regex.MatchString(Node.Name) {
				matches = append(matches, Node)
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
		for _, Node := range c.Items {
			if regex.MatchString(Node.Name) {
				matches = append(matches, Node)
			}
		}
	}
	list.Items = matches
	collection = &NodeCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
	return
}

func (c *NodeCollection) nodeSearchBak(names ...string) *NodeCollection {
	var list v1.NodeList
	var matches []v1.Node
	for _, name := range names {
		for _, item := range c.Items {
			if strings.Contains(item.Name, name) {
				matches = append(matches, item)
			}
		}
	}
	list.Items = matches
	return &NodeCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
}

// GetName returns the name of the resource.
func (r *Node) GetName() string {
	return r.Name
}

// GetUID returns the UID of the resource.
func (r *Node) GetUID() types.UID {
	return r.UID
}
