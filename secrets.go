package ak8s

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Secret Constants:
const (
	SecretAPIVersion = `v1`
	SecretListKind   = `List`
	SecretKind       = `Secret`
)

// SecretCollection Contains a Collection of Secrets.
type SecretCollection struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.SecretList
}

// Secret contains a v1.Secret resource.
type Secret struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	*v1.Secret
}

// GetAllSecrets returns All Secrets for the current namespace set on the client or all Secrets across all namespaces if not set.
func (c *Client) GetAllSecrets() (*SecretCollection, error) {
	list, err := c.CS.CoreV1().Secrets(c.NS).List(c.Options[ListOption].(*ListAction).Get())
	if err != nil {
		return &SecretCollection{}, err
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Version: SecretAPIVersion,
		Kind:    SecretListKind,
	})
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Version: SecretAPIVersion,
			Kind:    SecretKind,
		})
	}
	return &SecretCollection{
		list.APIVersion,
		list.Kind,
		list,
	}, nil
}

// GetSecrets returns Secrets for the given namespaces.
// If the namespace is not set on the client, the "default" namespace is used.
func (c *Client) GetSecrets(names ...string) (*SecretCollection, error) {
	if len(names) < 1 {
		return &SecretCollection{}, fmt.Errorf("no Secrets specified")
	}
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	var Secrets []v1.Secret
	var list v1.SecretList
	var errd string
	var Err error
	for _, name := range names {
		p, err := c.CS.CoreV1().Secrets(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
		if err != nil {
			errd += (err.Error() + fmt.Sprintf("\n"))
		} else {
			Secrets = append(Secrets, *p)
		}
	}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Version: SecretAPIVersion,
		Kind:    SecretListKind,
	})
	switch {
	case len(Secrets) < 1 && errd != "":
		return &SecretCollection{}, fmt.Errorf("%v", errd)
	case len(Secrets) < 1 && errd == "":
		return &SecretCollection{}, fmt.Errorf("%d Secrets not found", len(names))
	case len(Secrets) > 0 && errd != "":
		Err = fmt.Errorf("%v", errd)
	}
	list.Items = Secrets
	for i := 0; i < len(list.Items); i++ {
		list.Items[i].SetGroupVersionKind(schema.GroupVersionKind{
			Version: SecretAPIVersion,
			Kind:    SecretKind,
		})
	}
	return &SecretCollection{
		list.APIVersion,
		list.Kind,
		&list,
	}, Err
}

// GetSecret returns the Secret for the given Secret name.
func (c *Client) GetSecret(name string) (*Secret, error) {
	ns := c.NS
	if ns == "" {
		ns = `default`
	}
	p, err := c.CS.CoreV1().Secrets(ns).Get(name, c.Options[GetOption].(*GetAction).Get())
	if err != nil {
		return &Secret{}, err
	}
	p.SetGroupVersionKind(schema.GroupVersionKind{
		Version: SecretAPIVersion,
		Kind:    SecretKind,
	})
	return &Secret{
		p.APIVersion,
		p.Kind,
		p,
	}, nil
}

// GetNames returns all item names contained within the Collection.
func (c *SecretCollection) GetNames() []string {
	var names []string
	for _, Secret := range c.Items {
		names = append(names, Secret.Name)
	}
	return names
}

// GetKind returns the collection kind.
func (c *SecretCollection) GetKind() string {
	return c.Kind
}

// Len returns the number of items in the collection.
func (c *SecretCollection) Len() int {
	return len(c.Items)
}

// Get returns an item by name.
func (c *SecretCollection) Get(name string) Secret {
	for _, item := range c.Items {
		if item.Name == name {
			return Secret{
				SecretAPIVersion,
				SecretKind,
				&item,
			}
		}
	}
	return Secret{}
}

// Search conducts a wildcard search by names and returns matching items.
func (c *SecretCollection) Search(names ...string) (collection *SecretCollection) {
	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(fmt.Sprintf("%s", r), `regexp`) {
				collection = c.secretSearchBak(names...)
			} else {
				panic(r)
			}
		}
	}()
	var list v1.SecretList
	var regex *regexp.Regexp
	var regexString string
	var matches []v1.Secret
	switch {
	case len(names) <= 0:
		return c
	case len(names) == 1:
		regexString = names[0]
		regex = regexp.MustCompile(regexString)
		for _, Secret := range c.Items {
			if regex.MatchString(Secret.Name) {
				matches = append(matches, Secret)
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
		for _, Secret := range c.Items {
			if regex.MatchString(Secret.Name) {
				matches = append(matches, Secret)
			}
		}
	}
	list.Items = matches
	collection = &SecretCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
	return
}

func (c *SecretCollection) secretSearchBak(names ...string) *SecretCollection {
	var list v1.SecretList
	var matches []v1.Secret
	for _, name := range names {
		for _, item := range c.Items {
			if strings.Contains(item.Name, name) {
				matches = append(matches, item)
			}
		}
	}
	list.Items = matches
	return &SecretCollection{
		c.APIVersion,
		c.Kind,
		&list,
	}
}

// GetName returns the name of the resource.
func (r *Secret) GetName() string {
	return r.Name
}

// GetUID returns the UID of the resource.
func (r *Secret) GetUID() types.UID {
	return r.UID
}
