package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/bakarr/k3s-ignite/internal/k3s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Data structures for the template
type PodInfo struct {
	Name      string
	Namespace string
	Status    string
}

type PageData struct {
	Pods         []PodInfo
	Deployments  []appsv1.Deployment
	StatefulSets []appsv1.StatefulSet
	PVCs         []corev1.PersistentVolumeClaim
	Namespaces   []string
	CurrentNs    string
}

var client *kubernetes.Clientset

func main() {
	var err error
	client, err = k3s.GetInClusterClient()
	if err != nil {
		log.Fatalf("Failed to initialize K8s client: %v", err)
	}

	// Routes
	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/api/deploy", handleDeploy)
	http.HandleFunc("/api/delete", handleDelete)
	http.HandleFunc("/api/logs", handleLogs)
	http.HandleFunc("/api/status", handleStatusJSON)

	fmt.Println("🧠 K3s-Ignite Control Plane live on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// --- HANDLERS ---

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	selectedNs := r.URL.Query().Get("ns")
	ctx := context.TODO()

	// Fetch Data
	pods, _ := client.CoreV1().Pods(selectedNs).List(ctx, metav1.ListOptions{})
	deployments, _ := client.AppsV1().Deployments(selectedNs).List(ctx, metav1.ListOptions{})
	statefulsets, _ := client.AppsV1().StatefulSets(selectedNs).List(ctx, metav1.ListOptions{})
	pvcs, _ := client.CoreV1().PersistentVolumeClaims(selectedNs).List(ctx, metav1.ListOptions{})
	nsList, _ := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

	// Process Pod Status
	var podData []PodInfo
	for _, p := range pods.Items {
		status := string(p.Status.Phase)
		for _, cs := range p.Status.ContainerStatuses {
			if cs.State.Waiting != nil {
				status = cs.State.Waiting.Reason
			} else if cs.State.Terminated != nil {
				status = cs.State.Terminated.Reason
			}
		}
		podData = append(podData, PodInfo{Name: p.Name, Namespace: p.Namespace, Status: status})
	}

	var namespaces []string
	for _, n := range nsList.Items {
		namespaces = append(namespaces, n.Name)
	}

	data := PageData{
		Pods:         podData,
		Deployments:  deployments.Items,
		StatefulSets: statefulsets.Items,
		PVCs:         pvcs.Items,
		Namespaces:   namespaces,
		CurrentNs:    selectedNs,
	}

	// Load Template from File
	tmplPath := filepath.Join("templates", "dashboard.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), 500)
		return
	}
	tmpl.Execute(w, data)
}

func handleDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", 405)
		return
	}

	appName := r.FormValue("appName")
	image := r.FormValue("image")
	port, _ := strconv.Atoi(r.FormValue("port"))

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: appName},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": appName}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": appName}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "app",
						Image: image,
						Ports: []corev1.ContainerPort{{ContainerPort: int32(port)}},
					}},
				},
			},
		},
	}

	_, err := client.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		http.Error(w, "Deploy failed: "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	ns := r.URL.Query().Get("ns")
	client.CoreV1().Pods(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	http.Redirect(w, r, "/?ns="+ns, http.StatusSeeOther)
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	podName := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("ns")
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{TailLines: int64Ptr(100)})
	logs, _ := req.DoRaw(context.TODO())
	w.Header().Set("Content-Type", "text/plain")
	w.Write(logs)
}

func handleStatusJSON(w http.ResponseWriter, r *http.Request) {
	pods, _ := client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	w.Header().Set("Content-Type", "application/json")
	// Simplified response for JSON
	fmt.Fprintf(w, "{\"pod_count\": %d}", len(pods.Items))
}

// --- HELPERS ---
func int64Ptr(i int64) *int64 { return &i }
func int32Ptr(i int32) *int32 { return &i }
