package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeLister defines the interface for listing Kubernetes resources.
type KubeLister interface {
	GetNamespaces(ctx context.Context) ([]string, error)
	GetServices(ctx context.Context, namespace string) ([]string, error)
	GetConfigMaps(ctx context.Context, namespace string) ([]string, error)
	GetSecrets(ctx context.Context, namespace string) ([]string, error)
	GetPods(ctx context.Context, namespace string) ([]string, error)
	GetDeployments(ctx context.Context, namespace string) ([]string, error)
	GetStatefulSets(ctx context.Context, namespace string) ([]string, error)
	GetDaemonSets(ctx context.Context, namespace string) ([]string, error)
	GetPersistentVolumeClaims(ctx context.Context, namespace string) ([]string, error)
	GetClusterSecretStores(ctx context.Context) ([]string, error)
	GetSecretStores(ctx context.Context, namespace string) ([]string, error)
	TestConnection(ctx context.Context) error
}

// Client wraps the Kubernetes client for resource fetching.
type Client struct {
	clientset kubernetes.Interface
	dynamic   dynamic.Interface
}

// NewClient creates a new Kubernetes client using the default kubeconfig.
func NewClient() (*Client, error) {
	// Get the default kubeconfig path
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if envKubeconfig := os.Getenv("KUBECONFIG"); envKubeconfig != "" {
		kubeconfig = envKubeconfig
	}

	// Load the kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create the dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{
		clientset: clientset,
		dynamic:   dynamicClient,
	}, nil
}

// GetNamespaces returns a list of all namespaces in the cluster.
func (c *Client) GetNamespaces(ctx context.Context) ([]string, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	names := make([]string, len(namespaces.Items))
	for i := range namespaces.Items {
		names[i] = namespaces.Items[i].Name
	}
	return names, nil
}

// GetServices returns a list of services in the specified namespace.
func (c *Client) GetServices(ctx context.Context, namespace string) ([]string, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services in namespace %s: %w", namespace, err)
	}

	names := make([]string, len(services.Items))
	for i := range services.Items {
		names[i] = services.Items[i].Name
	}
	return names, nil
}

// GetConfigMaps returns a list of configmaps in the specified namespace.
func (c *Client) GetConfigMaps(ctx context.Context, namespace string) ([]string, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	configmaps, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps in namespace %s: %w", namespace, err)
	}

	names := make([]string, len(configmaps.Items))
	for i := range configmaps.Items {
		names[i] = configmaps.Items[i].Name
	}
	return names, nil
}

// GetSecrets returns a list of secrets in the specified namespace.
func (c *Client) GetSecrets(ctx context.Context, namespace string) ([]string, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	secrets, err := c.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets in namespace %s: %w", namespace, err)
	}

	names := make([]string, len(secrets.Items))
	for i := range secrets.Items {
		names[i] = secrets.Items[i].Name
	}
	return names, nil
}

// GetPods returns a list of pods in the specified namespace.
func (c *Client) GetPods(ctx context.Context, namespace string) ([]string, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	names := make([]string, len(pods.Items))
	for i := range pods.Items {
		names[i] = pods.Items[i].Name
	}
	return names, nil
}

// GetDeployments returns a list of deployments in the specified namespace.
func (c *Client) GetDeployments(ctx context.Context, namespace string) ([]string, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments in namespace %s: %w", namespace, err)
	}

	names := make([]string, len(deployments.Items))
	for i := range deployments.Items {
		names[i] = deployments.Items[i].Name
	}
	return names, nil
}

// GetStatefulSets returns a list of statefulsets in the specified namespace.
func (c *Client) GetStatefulSets(ctx context.Context, namespace string) ([]string, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	statefulsets, err := c.clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets in namespace %s: %w", namespace, err)
	}

	names := make([]string, len(statefulsets.Items))
	for i := range statefulsets.Items {
		names[i] = statefulsets.Items[i].Name
	}
	return names, nil
}

// GetDaemonSets returns a list of daemonsets in the specified namespace.
func (c *Client) GetDaemonSets(ctx context.Context, namespace string) ([]string, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	daemonsets, err := c.clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets in namespace %s: %w", namespace, err)
	}

	names := make([]string, len(daemonsets.Items))
	for i := range daemonsets.Items {
		names[i] = daemonsets.Items[i].Name
	}
	return names, nil
}

// GetPersistentVolumeClaims returns a list of PVCs in the specified namespace.
func (c *Client) GetPersistentVolumeClaims(ctx context.Context, namespace string) ([]string, error) {
	if c.clientset == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}

	pvcs, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list PVCs in namespace %s: %w", namespace, err)
	}

	names := make([]string, len(pvcs.Items))
	for i := range pvcs.Items {
		names[i] = pvcs.Items[i].Name
	}
	return names, nil
}

// TestConnection tests if the Kubernetes client can connect to the cluster.
func (c *Client) TestConnection(ctx context.Context) error {
	if c.clientset == nil {
		return fmt.Errorf("kubernetes client is not initialized")
	}
	_, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("failed to connect to kubernetes cluster: %w", err)
	}
	return nil
}

// GetClusterSecretStores returns a list of all ClusterSecretStore resources in the cluster.
func (c *Client) GetClusterSecretStores(ctx context.Context) ([]string, error) {
	if c.dynamic == nil {
		return nil, fmt.Errorf("dynamic client is not initialized")
	}

	gvr := schema.GroupVersionResource{
		Group:    "external-secrets.io",
		Version:  "v1",
		Resource: "clustersecretstores",
	}

	stores, err := c.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ClusterSecretStores: %w", err)
	}

	names := make([]string, len(stores.Items))
	for i := range stores.Items {
		names[i] = stores.Items[i].GetName()
	}
	return names, nil
}

// GetSecretStores returns a list of SecretStore resources in the specified namespace.
func (c *Client) GetSecretStores(ctx context.Context, namespace string) ([]string, error) {
	if c.dynamic == nil {
		return nil, fmt.Errorf("dynamic client is not initialized")
	}

	gvr := schema.GroupVersionResource{
		Group:    "external-secrets.io",
		Version:  "v1",
		Resource: "secretstores",
	}

	stores, err := c.dynamic.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list SecretStores in namespace %s: %w", namespace, err)
	}

	names := make([]string, len(stores.Items))
	for i := range stores.Items {
		names[i] = stores.Items[i].GetName()
	}
	return names, nil
}
