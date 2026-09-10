// Package kube wraps the small slice of the Kubernetes API the app needs:
// create a pod, read a pod, make sure the namespace exists.
package kube

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Namespace is where the app creates and watches pods.
const Namespace = "pod-manager"

// Client is what the API and the workers need from Kubernetes.
type Client interface {
	EnsureNamespace(ctx context.Context) error
	CreatePod(ctx context.Context, name, image string) error
	GetPod(ctx context.Context, name string) (*corev1.Pod, error)
}

type client struct {
	cs *kubernetes.Clientset
}

// New builds a client from KUBECONFIG / the default kubeconfig
// (e.g. the context minikube creates), so it talks to the local cluster.
func New() (Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	restCfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	cs, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %w", err)
	}
	return &client{cs: cs}, nil
}

// EnsureNamespace creates the working namespace if it does not exist yet.
func (c *client) EnsureNamespace(ctx context.Context) error {
	_, err := c.cs.CoreV1().Namespaces().Get(ctx, Namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("get namespace %s: %w", Namespace, err)
	}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: Namespace}}
	if _, err := c.cs.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create namespace %s: %w", Namespace, err)
	}
	return nil
}

// CreatePod creates a simple single-container pod.
func (c *client) CreatePod(ctx context.Context, name, image string) error {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: Namespace,
			Labels:    map[string]string{"managed-by": "pod-manager"},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			Containers: []corev1.Container{{
				Name:            "main",
				Image:           image,
				ImagePullPolicy: corev1.PullIfNotPresent,
			}},
		},
	}
	if _, err := c.cs.CoreV1().Pods(Namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

// GetPod reads one pod from the cluster.
func (c *client) GetPod(ctx context.Context, name string) (*corev1.Pod, error) {
	return c.cs.CoreV1().Pods(Namespace).Get(ctx, name, metav1.GetOptions{})
}

// NewWithTimeout is a helper used at startup.
func NewWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}
