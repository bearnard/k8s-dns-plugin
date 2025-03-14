package utils

import (
	"context"
	"fmt"
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetK8sClient initializes a Kubernetes client from the kubeconfig file.
func GetK8sClient(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	return clientset, nil
}

// GetServices retrieves all services in the specified namespace.
func GetServices(clientset *kubernetes.Clientset, namespace string) ([]v1.Service, error) {
	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	return services.Items, nil
}

// LogError logs an error message.
func LogError(err error) {
	if err != nil {
		log.Printf("Error: %v", err)
	}
}