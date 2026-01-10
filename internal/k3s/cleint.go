package k3s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClient creates a native Kubernetes client from a local config file
func GetClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	// 1. Build the config from the file we just saved
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	// 2. Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}