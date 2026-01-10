package k3s

import (
	"strings"
)

// FetchAndFixConfig reaches into the VM and pulls the credentials
func FetchAndFixConfig(remoteConfig string, publicIP string) string {
	// K3s saves the server as https://127.0.0.1:6443
	// We need it to be https://34.55.10.49:6443
	fixedConfig := strings.Replace(remoteConfig, "127.0.0.1", publicIP, 1)
	
	// Optional: Rename the cluster context so it doesn't conflict with local ones
	fixedConfig = strings.Replace(fixedConfig, "default", "ignite-cluster", -1)
	
	return fixedConfig
}