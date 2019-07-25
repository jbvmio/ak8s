package ak8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// ConfigPath sets the location for kubeconfig.
	ConfigPath        string
	defaultConfigPath = string(homeDir() + "/.kube/config")
)

// GetKubeConfig returns a rest.Config using the current-context set in kubeconfig
// Use ConfigLocation variable to set the kubeconfig location if not located in $HOME/.kube/config
func GetKubeConfig() (*rest.Config, error) {
	var config *rest.Config
	var kubeconfig string
	if ConfigPath == "" {
		if fileExists(defaultConfigPath) {
			kubeconfig = defaultConfigPath
		} else {
			return config, fmt.Errorf("cannot locate kubeconfig at %v", defaultConfigPath)
		}
	} else {
		kubeconfig = ConfigPath
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// CreateClientSet returns a Clientset from your ~/.kube/config.
func CreateClientSet() (*kubernetes.Clientset, error) {
	config, err := GetKubeConfig()
	if err != nil {
		return &kubernetes.Clientset{}, err
	}
	return kubernetes.NewForConfig(config)
}

// CreateICClientSet returns a Clientset from within a running pod.
func CreateICClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return &kubernetes.Clientset{}, err
	}
	return kubernetes.NewForConfig(config)
}

// CreateClientSetFromConfig returns a Clientset from the given configPath.
func CreateClientSetFromConfig(configPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return &kubernetes.Clientset{}, err
	}
	return kubernetes.NewForConfig(config)
}

// CreateUserClientSet returns a clientset using username/password values.
func CreateUserClientSet(host, username, password string, insecure bool) (*kubernetes.Clientset, error) {
	tls := rest.TLSClientConfig{Insecure: insecure}
	config := rest.Config{
		Host:            host,
		Username:        username,
		Password:        password,
		TLSClientConfig: tls,
	}
	return kubernetes.NewForConfig(&config)
}
