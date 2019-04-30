package ak8s

import (
	"log"

	"k8s.io/client-go/kubernetes"
)

// Client intereacts with Kubernetes.
type Client struct {
	Options ActionsMap
	CS      *kubernetes.Clientset
	NS      string
}

// NewClient returns a new Client using your kube config or inCluster if running within a pod.
func NewClient(inCluster bool) (*Client, error) {
	var client Client
	if inCluster {
		cs, err := CreateICClientSet()
		if err != nil {
			return &client, err
		}
		client.CS = cs
		client.Options = makeActionMap()
		return &client, nil
	}
	cs, err := CreateClientSet()
	if err != nil {
		return &client, err
	}
	client.CS = cs
	client.Options = makeActionMap()
	return &client, nil
}

// NewUserClient returns a new Client using username/password values.
func NewUserClient(host, username, password string, insecure bool) (*Client, error) {
	var client Client
	cs, err := CreateUserClientSet(host, username, password, insecure)
	if err != nil {
		return &client, err
	}
	client.CS = cs
	client.Options = makeActionMap()
	return &client, nil
}

// GetAPIGroups returns the preferred API Group Versions.
func (c *Client) GetAPIGroups() []string {
	var groups []string
	list, err := c.CS.ServerGroups()
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}
	for _, g := range list.Groups {
		groups = append(groups, g.PreferredVersion.GroupVersion)
	}
	return groups
}
