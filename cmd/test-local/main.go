package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	fmt.Println("🏠 Local Ignition Started...")

	// Instead of SSH, we use local 'kubectl'
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
    image: bakarr/ignite-brain:v1
    ports:
    - containerPort: 8080
`
	// We "pipe" the manifest into the local kubectl
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	stdin, _ := cmd.StdinPipe()
	go func() {
		defer stdin.Close()
		fmt.Fprint(stdin, manifest)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("❌ Local deployment failed: %v\nOutput: %s", err, string(out))
	}

	fmt.Println("✅ Local Brain Injected!")
	fmt.Println("👉 Run 'kubectl port-forward pod/ignite-brain 8080:8080' to see your API")
}
