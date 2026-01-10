package k3s

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeployIgniteUI(clientset *kubernetes.Clientset) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "ignite-dashboard"},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "ignite-ui"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "ignite-ui"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "dashboard",
							Image: "nginx", // For now, let's use nginx as a placeholder!
							Ports: []corev1.ContainerPort{{ContainerPort: 80}},
						},
					},
				},
			},
		},
	}

	_, err := clientset.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})
	return err
}

func int32Ptr(i int32) *int32 { return &i }