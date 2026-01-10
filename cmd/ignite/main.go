package main

import (
	"fmt"
	"log"
	"os"
	"github.com/bakarr/k3s-ignite/internal/k3s"
	"github.com/bakarr/k3s-ignite/internal/ssh"
)

func main() {
	ip := "34.55.10.49"
	user := "ubuntu"
	keyPath := "C:\\Users\\BOSS\\.ssh\\id_rsa"

	// ... (your existing connection and install logic here) ...

	fmt.Println("🛰️  Fetching cluster credentials...")
	
	// We use 'sudo cat' because the config is owned by root
	rawConfig, err := ssh.ExecuteCommand(ip, user, keyPath, "sudo cat /etc/rancher/k3s/k3s.yaml")
	if err != nil {
		log.Fatalf("❌ Could not read remote config: %v", err)
	}

	// Fix the IP
	finalConfig := k3s.FetchAndFixConfig(rawConfig, ip)

	// Save to local file
	fileName := "ignite.kubeconfig"
	err = os.WriteFile(fileName, []byte(finalConfig), 0600)
	if err != nil {
		log.Fatalf("❌ Failed to save config locally: %v", err)
	}

	fmt.Printf("✅ Success! Config saved as %s\n", fileName)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("🔥 Run this to see your nodes: \nkubectl --kubeconfig=%s get nodes\n", fileName)

    kubeconfig := "ignite.kubeconfig"
    client, err := k3s.GetClient(kubeconfig)
    if err != nil {
        log.Fatalf("❌ Failed to create K8s client: %v", err)
    }

    fmt.Println("🔍 Querying Cluster for running pods...")
    
    // List all pods in the 'kube-system' namespace
    pods, err := client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        log.Fatalf("❌ Failed to list pods: %v", err)
    }

    fmt.Printf("📦 Found %d system pods:\n", len(pods.Items))
    for _, pod := range pods.Items {
        fmt.Printf("   - [%s] %s\n", pod.Status.Phase, pod.Name)
    }
}