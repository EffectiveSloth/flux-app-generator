package kubernetes

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewClient_WithKubeconfigEnvVar(t *testing.T) {
	// Test with KUBECONFIG environment variable
	originalKubeconfig := os.Getenv("KUBECONFIG")
	defer func() {
		_ = os.Setenv("KUBECONFIG", originalKubeconfig)
	}()

	// Set a non-existent kubeconfig to test error handling
	err := os.Setenv("KUBECONFIG", "/non/existent/path")
	require.NoError(t, err)

	client, err := NewClient()
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to load kubeconfig")
}

func TestNewClient_WithInvalidKubeconfig(t *testing.T) {
	// Create a temporary invalid kubeconfig file
	tempDir := t.TempDir()
	invalidKubeconfig := filepath.Join(tempDir, "invalid-config")

	// Write invalid content
	err := os.WriteFile(invalidKubeconfig, []byte("invalid yaml content"), 0o600)
	require.NoError(t, err)

	// Set KUBECONFIG to the invalid file
	originalKubeconfig := os.Getenv("KUBECONFIG")
	defer func() {
		_ = os.Setenv("KUBECONFIG", originalKubeconfig)
	}()
	err = os.Setenv("KUBECONFIG", invalidKubeconfig)
	require.NoError(t, err)

	client, err := NewClient()
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to load kubeconfig")
}

func TestNewClient_WithValidKubeconfig(t *testing.T) {
	// Create a temporary valid kubeconfig file
	tempDir := t.TempDir()
	validKubeconfig := filepath.Join(tempDir, "valid-config")

	// Write minimal valid kubeconfig content
	validConfig := `
apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    server: https://test-server:6443
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	err := os.WriteFile(validKubeconfig, []byte(validConfig), 0o600)
	require.NoError(t, err)

	// Set KUBECONFIG to the valid file
	originalKubeconfig := os.Getenv("KUBECONFIG")
	defer func() {
		_ = os.Setenv("KUBECONFIG", originalKubeconfig)
	}()
	err = os.Setenv("KUBECONFIG", validKubeconfig)
	require.NoError(t, err)

	// This should fail because the server doesn't exist, but it should get past kubeconfig loading
	_, err = NewClient()
	// The error should be about connection, not kubeconfig loading
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create kubernetes client")
	}
}

func TestClient_GetNamespaces_WithFakeClient(t *testing.T) {
	// Create a fake clientset for testing
	fakeClientset := fake.NewSimpleClientset()

	// Create test namespaces
	_, err := fakeClientset.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	_, err = fakeClientset.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "kube-system"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	// Create a client with the fake clientset
	client := &Client{
		clientset: fakeClientset,
	}

	namespaces, err := client.GetNamespaces(context.Background())
	assert.NoError(t, err)
	assert.Len(t, namespaces, 2)
	assert.Contains(t, namespaces, "default")
	assert.Contains(t, namespaces, "kube-system")
}

func TestClient_GetServices_WithFakeClient(t *testing.T) {
	// Create a fake clientset for testing
	fakeClientset := fake.NewSimpleClientset()

	// Create test services
	_, err := fakeClientset.CoreV1().Services("default").Create(context.Background(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "service1"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	_, err = fakeClientset.CoreV1().Services("default").Create(context.Background(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "service2"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	// Create a client with the fake clientset
	client := &Client{
		clientset: fakeClientset,
	}

	services, err := client.GetServices(context.Background(), "default")
	assert.NoError(t, err)
	assert.Len(t, services, 2)
	assert.Contains(t, services, "service1")
	assert.Contains(t, services, "service2")
}

func TestClient_GetServices_EmptyNamespace(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
	}

	services, err := client.GetServices(context.Background(), "empty-namespace")
	assert.NoError(t, err)
	assert.Len(t, services, 0)
}

func TestClient_GetConfigMaps_WithFakeClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	// Create test configmaps
	_, err := fakeClientset.CoreV1().ConfigMaps("default").Create(context.Background(), &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "config1"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	client := &Client{
		clientset: fakeClientset,
	}

	configmaps, err := client.GetConfigMaps(context.Background(), "default")
	assert.NoError(t, err)
	assert.Len(t, configmaps, 1)
	assert.Contains(t, configmaps, "config1")
}

func TestClient_GetSecrets_WithFakeClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	// Create test secrets
	_, err := fakeClientset.CoreV1().Secrets("default").Create(context.Background(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "secret1"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	client := &Client{
		clientset: fakeClientset,
	}

	secrets, err := client.GetSecrets(context.Background(), "default")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Contains(t, secrets, "secret1")
}

func TestClient_GetPods_WithFakeClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	// Create test pods
	_, err := fakeClientset.CoreV1().Pods("default").Create(context.Background(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	client := &Client{
		clientset: fakeClientset,
	}

	pods, err := client.GetPods(context.Background(), "default")
	assert.NoError(t, err)
	assert.Len(t, pods, 1)
	assert.Contains(t, pods, "pod1")
}

func TestClient_GetDeployments_WithFakeClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	// Create test deployments
	_, err := fakeClientset.AppsV1().Deployments("default").Create(context.Background(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "deployment1"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	client := &Client{
		clientset: fakeClientset,
	}

	deployments, err := client.GetDeployments(context.Background(), "default")
	assert.NoError(t, err)
	assert.Len(t, deployments, 1)
	assert.Contains(t, deployments, "deployment1")
}

func TestClient_GetStatefulSets_WithFakeClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	// Create test statefulsets
	_, err := fakeClientset.AppsV1().StatefulSets("default").Create(context.Background(), &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "statefulset1"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	client := &Client{
		clientset: fakeClientset,
	}

	statefulsets, err := client.GetStatefulSets(context.Background(), "default")
	assert.NoError(t, err)
	assert.Len(t, statefulsets, 1)
	assert.Contains(t, statefulsets, "statefulset1")
}

func TestClient_GetDaemonSets_WithFakeClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	// Create test daemonsets
	_, err := fakeClientset.AppsV1().DaemonSets("default").Create(context.Background(), &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "daemonset1"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	client := &Client{
		clientset: fakeClientset,
	}

	daemonsets, err := client.GetDaemonSets(context.Background(), "default")
	assert.NoError(t, err)
	assert.Len(t, daemonsets, 1)
	assert.Contains(t, daemonsets, "daemonset1")
}

func TestClient_GetPersistentVolumeClaims_WithFakeClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	// Create test PVCs
	_, err := fakeClientset.CoreV1().PersistentVolumeClaims("default").Create(context.Background(), &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "pvc1"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	client := &Client{
		clientset: fakeClientset,
	}

	pvcs, err := client.GetPersistentVolumeClaims(context.Background(), "default")
	assert.NoError(t, err)
	assert.Len(t, pvcs, 1)
	assert.Contains(t, pvcs, "pvc1")
}

func TestClient_TestConnection_WithFakeClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
	}

	err := client.TestConnection(context.Background())
	assert.NoError(t, err)
}

func TestClient_TestConnection_WithNilClient(t *testing.T) {
	client := &Client{
		clientset: nil,
	}

	err := client.TestConnection(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kubernetes client is not initialized")
}

func TestClient_GetClusterSecretStores_WithNilDynamicClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		dynamic:   nil,
	}

	// This should handle nil dynamic client gracefully
	_, err := client.GetClusterSecretStores(context.Background())
	assert.Error(t, err)
}

func TestClient_GetSecretStores_WithNilDynamicClient(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
		dynamic:   nil,
	}

	// This should handle nil dynamic client gracefully
	_, err := client.GetSecretStores(context.Background(), "default")
	assert.Error(t, err)
}

func TestClient_InterfaceCompliance(_ *testing.T) {
	// Test that Client implements KubeLister interface
	var _ KubeLister = (*Client)(nil)
}

func TestClient_NilChecks(t *testing.T) {
	// Test that methods handle nil clientset gracefully
	client := &Client{
		clientset: nil,
		dynamic:   nil,
	}

	ctx := context.Background()

	// All methods should return errors when clientset is nil
	_, err := client.GetNamespaces(ctx)
	assert.Error(t, err)

	_, err = client.GetServices(ctx, "default")
	assert.Error(t, err)

	_, err = client.GetConfigMaps(ctx, "default")
	assert.Error(t, err)

	_, err = client.GetSecrets(ctx, "default")
	assert.Error(t, err)

	_, err = client.GetPods(ctx, "default")
	assert.Error(t, err)

	_, err = client.GetDeployments(ctx, "default")
	assert.Error(t, err)

	_, err = client.GetStatefulSets(ctx, "default")
	assert.Error(t, err)

	_, err = client.GetDaemonSets(ctx, "default")
	assert.Error(t, err)

	_, err = client.GetPersistentVolumeClaims(ctx, "default")
	assert.Error(t, err)

	err = client.TestConnection(ctx)
	assert.Error(t, err)
}

func TestClient_ErrorHandling(t *testing.T) {
	// Test error handling with invalid namespace names
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
	}

	// Test with empty namespace
	_, err := client.GetServices(context.Background(), "")
	assert.NoError(t, err) // Empty namespace is valid in fake client

	// Test with very long namespace name
	longNamespace := string(make([]byte, 1000))
	_, err = client.GetServices(context.Background(), longNamespace)
	assert.NoError(t, err) // Fake client doesn't validate namespace length
}

func TestClient_ContextHandling(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
	}

	// Test with normal context
	ctx := context.Background()
	_, err := client.GetNamespaces(ctx)
	assert.NoError(t, err)

	// Test with cancelled context (fake client may not respect context cancellation)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.GetNamespaces(ctx)
	// Fake client might not respect cancelled context, so we just test it doesn't panic
	// Real k8s client would return an error, but fake client might not
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
}

func TestClient_ResourceTypeHandling(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: fakeClientset,
	}

	// Test that all resource types are handled correctly
	ctx := context.Background()
	namespace := "default"

	// Test all resource types
	_, err := client.GetServices(ctx, namespace)
	assert.NoError(t, err)

	_, err = client.GetConfigMaps(ctx, namespace)
	assert.NoError(t, err)

	_, err = client.GetSecrets(ctx, namespace)
	assert.NoError(t, err)

	_, err = client.GetPods(ctx, namespace)
	assert.NoError(t, err)

	_, err = client.GetDeployments(ctx, namespace)
	assert.NoError(t, err)

	_, err = client.GetStatefulSets(ctx, namespace)
	assert.NoError(t, err)

	_, err = client.GetDaemonSets(ctx, namespace)
	assert.NoError(t, err)

	_, err = client.GetPersistentVolumeClaims(ctx, namespace)
	assert.NoError(t, err)
}
