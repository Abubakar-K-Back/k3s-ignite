package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/bakarr/k3s-ignite/internal/k3s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The HTML Template (Professional look with simple CSS)
const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>K3s-Ignite Dashboard</title>
    <style>
        body { font-family: sans-serif; background: #121212; color: white; padding: 20px; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #333; }
        th { background-color: #1f1f1f; }
        .status-running { color: #00ff00; font-weight: bold; }
        .status-error { color: #ff4444; font-weight: bold; }
        .header { display: flex; justify-content: space-between; align-items: center; }
    </style>
    <script>setTimeout(() => { location.reload(); }, 5000);</script>
</head>
<body>
   <div class="header">
    <h1>🔥 K3s-Ignite</h1>
    <div>
        <label>Filter Namespace: </label>
        <select onchange="window.location.href='/?ns=' + this.value" style="background: #333; color: white; padding: 5px;">
            <option value="">All Namespaces</option>
            {{$current := .CurrentNs}}
            {{range .Namespaces}}
                <option value="{{.}}" {{if eq . $current}}selected{{end}}>{{.}}</option>
            {{end}}
        </select>
    </div>
</div>
<table>
        <tr>
            <th>Pod Name</th>
            <th>Namespace</th>
            <th>Status</th>
            <th>Actions</th> </tr>
        {{range .Pods}} <tr>
            <td>{{.Name}}</td>
            <td>{{.Namespace}}</td>
            <td class="{{if eq .Status "Running"}}status-running{{else}}status-error{{end}}">{{.Status}}</td>
            <td>
                <a href="/api/logs?name={{.Name}}&ns={{.Namespace}}" target="_blank" style="color: #44aaff;">[ View Logs ]</a>
            </td>
        </tr>
        {{end}}
    </table>
</body>
</html>
`

type PodInfo struct {
	Name      string
	Namespace string
	Status    string
}

func main() {
	client, err := k3s.GetInClusterClient()
	if err != nil {
		panic(err)
	}

	// 1. The HTML Dashboard
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get the namespace from the URL: e.g., /?ns=default
		selectedNs := r.URL.Query().Get("ns")

		// If no namespace is selected, we look at all namespaces ("")
		pods, _ := client.CoreV1().Pods(selectedNs).List(context.TODO(), metav1.ListOptions{})
		// We also need a list of ALL namespaces to build the filter menu
		nsList, _ := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})

		type PageData struct {
			Pods       []PodInfo
			Namespaces []string
			CurrentNs  string
		}

		var podData []PodInfo
		for _, p := range pods.Items {
			podData = append(podData, PodInfo{
				Name:      p.Name,
				Namespace: p.Namespace,
				Status:    string(p.Status.Phase),
			})
		}

		var namespaces []string
		for _, n := range nsList.Items {
			namespaces = append(namespaces, n.Name)
		}

		data := PageData{
			Pods:       podData,
			Namespaces: namespaces,
			CurrentNs:  selectedNs,
		}

		tmpl := template.Must(template.New("dashboard").Parse(dashboardHTML))
		tmpl.Execute(w, data)
	})

	// 2. The JSON API (Keep this for other tools)
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		pods, _ := client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pods.Items)
	})
	// Add this new handler in your main() function
	http.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		podName := r.URL.Query().Get("name")
		namespace := r.URL.Query().Get("ns")

		podLogOptions := &v1.PodLogOptions{
			TailLines: int64Ptr(100),
		}

		req := client.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
		logs, err := req.DoRaw(context.TODO())
		if err != nil {
			http.Error(w, "Failed to get logs", 500)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write(logs)
	})

	fmt.Println("🧠 Brain UI live on :8080")
	http.ListenAndServe(":8080", nil)
}

// Helper function to handle pointers
func int64Ptr(i int64) *int64 { return &i }
