package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bakarr/k3s-ignite/internal/ssh"
)

func main() {
	ip := "34.55.10.49"
	user := "ubuntu"
	keyPath := "C:\\Users\\BOSS\\.ssh\\id_rsa"

	fmt.Println("🚀 Phase 1: Installing K3s...")
	ssh.ExecuteCommand(ip, user, keyPath, "curl -sfL https://get.k3s.io | sh -")

	fmt.Println("🧠 Phase 2: Injecting Monitoring Brain...")
	// We combine RBAC and Pod into one command
	manifest := `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: monitor-sa
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: monitor-admin
subjects:
- kind: ServiceAccount
  name: monitor-sa
  namespace: default
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: Pod
metadata:
  name: ignite-brain
spec:
  serviceAccountName: monitor-sa
  containers:
  - name: engine
    image: bakarr/ignite-brain:latest
    ports:
    - containerPort: 8080
`
	fmt.Println("⏳ Waiting for Cluster to stabilize...")
	for i := 0; i < 10; i++ {
		_, err := ssh.ExecuteCommand(ip, user, keyPath, "sudo k3s kubectl get nodes")
		if err == nil {
			fmt.Println("✅ Cluster is Ready!")
			break
		}
		fmt.Printf("... Still waiting (%d/10)\n", i+1)
		time.Sleep(10 * time.Second)
	}
	cmd := fmt.Sprintf("echo '%s' | sudo k3s kubectl apply -f -", manifest)
	_, err := ssh.ExecuteCommand(ip, user, keyPath, cmd)
	if err != nil {
		log.Fatalf("❌ Injection failed: %v", err)
	}

	fmt.Println("✅ Ignition Complete! Cluster is now self-monitoring.")
}
