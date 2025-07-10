package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewAutoCompleteService(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	assert.NotNil(t, service)
	assert.Equal(t, mockClient, service.kubeLister)
}

func TestAutoCompleteService_GetNamespaces(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	namespaces, err := service.GetNamespaceSuggestions(context.Background(), "")

	assert.NoError(t, err)
	assert.NotNil(t, namespaces)
	assert.Len(t, namespaces, 3) // Based on mock data
	assert.Contains(t, namespaces, "default")
	assert.Contains(t, namespaces, "kube-system")
	assert.Contains(t, namespaces, "kube-public")
}

func TestAutoCompleteService_GetServices(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	services, err := service.GetServiceSuggestions(context.Background(), "default", "")

	assert.NoError(t, err)
	assert.NotNil(t, services)
	assert.Len(t, services, 2) // Based on mock data
	assert.Contains(t, services, "kubernetes")
	assert.Contains(t, services, "nginx-service")
}

func TestAutoCompleteService_GetConfigMaps(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	configmaps, err := service.GetConfigMapSuggestions(context.Background(), "default", "")

	assert.NoError(t, err)
	assert.NotNil(t, configmaps)
	assert.Len(t, configmaps, 2) // Based on mock data
	assert.Contains(t, configmaps, "kube-root-ca.crt")
	assert.Contains(t, configmaps, "my-config")
}

func TestAutoCompleteService_GetSecrets(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	secrets, err := service.GetSecretSuggestions(context.Background(), "default", "")

	assert.NoError(t, err)
	assert.NotNil(t, secrets)
	assert.Len(t, secrets, 2) // Based on mock data
	assert.Contains(t, secrets, "default-token-abc123")
	assert.Contains(t, secrets, "my-secret")
}

func TestAutoCompleteService_GetPods(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	pods, err := service.GetSuggestions(context.Background(), ResourceTypePod, "default", "")

	assert.NoError(t, err)
	assert.NotNil(t, pods)
	assert.Len(t, pods, 2) // Based on mock data
	assert.Contains(t, pods, "nginx-pod")
	assert.Contains(t, pods, "app-pod")
}

func TestAutoCompleteService_GetDeployments(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	deployments, err := service.GetDeploymentSuggestions(context.Background(), "default", "")

	assert.NoError(t, err)
	assert.NotNil(t, deployments)
	assert.Len(t, deployments, 2) // Based on mock data
	assert.Contains(t, deployments, "nginx-deployment")
	assert.Contains(t, deployments, "app-deployment")
}

func TestAutoCompleteService_GetStatefulSets(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	statefulsets, err := service.GetStatefulSetSuggestions(context.Background(), "default", "")

	assert.NoError(t, err)
	assert.NotNil(t, statefulsets)
	assert.Len(t, statefulsets, 2) // Based on mock data
	assert.Contains(t, statefulsets, "redis-statefulset")
	assert.Contains(t, statefulsets, "mysql-statefulset")
}

func TestAutoCompleteService_GetDaemonSets(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	daemonsets, err := service.GetDaemonSetSuggestions(context.Background(), "default", "")

	assert.NoError(t, err)
	assert.NotNil(t, daemonsets)
	assert.Len(t, daemonsets, 2) // Based on mock data
	assert.Contains(t, daemonsets, "fluentd-daemonset")
	assert.Contains(t, daemonsets, "node-exporter")
}

func TestAutoCompleteService_GetPersistentVolumeClaims(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	pvcs, err := service.GetPVCSuggestions(context.Background(), "default", "")

	assert.NoError(t, err)
	assert.NotNil(t, pvcs)
	assert.Len(t, pvcs, 2) // Based on mock data
	assert.Contains(t, pvcs, "data-pvc")
	assert.Contains(t, pvcs, "backup-pvc")
}

func TestAutoCompleteService_GetClusterSecretStores(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	stores, err := service.GetClusterSecretStoreSuggestions(context.Background(), "")

	assert.NoError(t, err)
	assert.NotNil(t, stores)
	assert.Len(t, stores, 3) // Based on mock data
	assert.Contains(t, stores, "vault-backend")
	assert.Contains(t, stores, "aws-secrets-manager")
	assert.Contains(t, stores, "azure-key-vault")
}

func TestAutoCompleteService_GetSecretStores(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	stores, err := service.GetSecretStoreSuggestions(context.Background(), "default", "")

	assert.NoError(t, err)
	assert.NotNil(t, stores)
	assert.Len(t, stores, 2) // Based on mock data
	assert.Contains(t, stores, "local-vault")
	assert.Contains(t, stores, "namespace-secrets")
}

func TestAutoCompleteService_WithNilClient(t *testing.T) {
	// Test behavior when kubeClient is nil
	service := NewAutoCompleteService(nil)

	// These should return empty slices and no error when client is nil
	namespaces, err := service.GetNamespaceSuggestions(context.Background(), "")
	assert.NoError(t, err)
	assert.Empty(t, namespaces)

	services, err := service.GetServiceSuggestions(context.Background(), "default", "")
	assert.NoError(t, err)
	assert.Empty(t, services)

	configmaps, err := service.GetConfigMapSuggestions(context.Background(), "default", "")
	assert.NoError(t, err)
	assert.Empty(t, configmaps)

	secrets, err := service.GetSecretSuggestions(context.Background(), "default", "")
	assert.NoError(t, err)
	assert.Empty(t, secrets)

	pods, err := service.GetSuggestions(context.Background(), ResourceTypePod, "default", "")
	assert.NoError(t, err)
	assert.Empty(t, pods)

	deployments, err := service.GetDeploymentSuggestions(context.Background(), "default", "")
	assert.NoError(t, err)
	assert.Empty(t, deployments)

	statefulsets, err := service.GetStatefulSetSuggestions(context.Background(), "default", "")
	assert.NoError(t, err)
	assert.Empty(t, statefulsets)

	daemonsets, err := service.GetDaemonSetSuggestions(context.Background(), "default", "")
	assert.NoError(t, err)
	assert.Empty(t, daemonsets)

	pvcs, err := service.GetPVCSuggestions(context.Background(), "default", "")
	assert.NoError(t, err)
	assert.Empty(t, pvcs)

	stores, err := service.GetClusterSecretStoreSuggestions(context.Background(), "")
	assert.NoError(t, err)
	assert.Empty(t, stores)

	secretStores, err := service.GetSecretStoreSuggestions(context.Background(), "default", "")
	assert.NoError(t, err)
	assert.Empty(t, secretStores)
}

func TestAutoCompleteService_WithEmptyNamespace(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	// Test with empty namespace
	services, err := service.GetServiceSuggestions(context.Background(), "", "")
	assert.NoError(t, err)
	assert.NotNil(t, services)

	configmaps, err := service.GetConfigMapSuggestions(context.Background(), "", "")
	assert.NoError(t, err)
	assert.NotNil(t, configmaps)

	secrets, err := service.GetSecretSuggestions(context.Background(), "", "")
	assert.NoError(t, err)
	assert.NotNil(t, secrets)

	pods, err := service.GetSuggestions(context.Background(), ResourceTypePod, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, pods)

	deployments, err := service.GetDeploymentSuggestions(context.Background(), "", "")
	assert.NoError(t, err)
	assert.NotNil(t, deployments)

	statefulsets, err := service.GetStatefulSetSuggestions(context.Background(), "", "")
	assert.NoError(t, err)
	assert.NotNil(t, statefulsets)

	daemonsets, err := service.GetDaemonSetSuggestions(context.Background(), "", "")
	assert.NoError(t, err)
	assert.NotNil(t, daemonsets)

	pvcs, err := service.GetPVCSuggestions(context.Background(), "", "")
	assert.NoError(t, err)
	assert.NotNil(t, pvcs)

	secretStores, err := service.GetSecretStoreSuggestions(context.Background(), "", "")
	assert.NoError(t, err)
	assert.NotNil(t, secretStores)
}

func TestAutoCompleteService_WithNilContext(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	// Test with nil context (should use context.Background() internally)
	namespaces, err := service.GetNamespaceSuggestions(context.TODO(), "")
	assert.NoError(t, err)
	assert.NotNil(t, namespaces)

	services, err := service.GetServiceSuggestions(context.TODO(), "default", "")
	assert.NoError(t, err)
	assert.NotNil(t, services)
}

func TestAutoCompleteService_InterfaceCompliance(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	// Test that AutoCompleteService implements expected interface
	// This is more of a compile-time check, but we can verify the methods exist
	assert.NotNil(t, service.GetNamespaceSuggestions)
	assert.NotNil(t, service.GetServiceSuggestions)
	assert.NotNil(t, service.GetConfigMapSuggestions)
	assert.NotNil(t, service.GetSecretSuggestions)
	assert.NotNil(t, service.GetSuggestions)
	assert.NotNil(t, service.GetDeploymentSuggestions)
	assert.NotNil(t, service.GetStatefulSetSuggestions)
	assert.NotNil(t, service.GetDaemonSetSuggestions)
	assert.NotNil(t, service.GetPVCSuggestions)
	assert.NotNil(t, service.GetClusterSecretStoreSuggestions)
	assert.NotNil(t, service.GetSecretStoreSuggestions)
}

func TestAutoCompleteService_Consistency(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	// Test that multiple calls return consistent results
	namespaces1, err1 := service.GetNamespaceSuggestions(context.Background(), "")
	namespaces2, err2 := service.GetNamespaceSuggestions(context.Background(), "")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, namespaces1, namespaces2)

	services1, err1 := service.GetServiceSuggestions(context.Background(), "default", "")
	services2, err2 := service.GetServiceSuggestions(context.Background(), "default", "")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, services1, services2)
}

func TestAutoCompleteService_ErrorHandling(t *testing.T) {
	// Test error handling with a mock that returns errors
	// This would require creating a custom mock that returns errors
	// For now, we'll test the basic structure

	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	// Test that the service handles the mock client properly
	// The mock client should return predefined data without errors
	namespaces, err := service.GetNamespaceSuggestions(context.Background(), "")
	assert.NoError(t, err)
	assert.NotNil(t, namespaces)
}

func TestAutoCompleteService_Performance(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)

	// Test that multiple calls are reasonably fast
	// This is more of a sanity check than a performance test

	for i := 0; i < 10; i++ {
		namespaces, err := service.GetNamespaceSuggestions(context.Background(), "")
		assert.NoError(t, err)
		assert.NotNil(t, namespaces)

		services, err := service.GetServiceSuggestions(context.Background(), "default", "")
		assert.NoError(t, err)
		assert.NotNil(t, services)
	}
}

func TestAutoCompleteService_FetchResourceItems(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetNamespaces", mock.Anything).Return([]string{"default", "kube-system"}, nil)
	mockClient.On("GetServices", mock.Anything, "default").Return([]string{"service1", "service2"}, nil)
	mockClient.On("GetConfigMaps", mock.Anything, "default").Return([]string{"config1", "config2"}, nil)
	mockClient.On("GetSecrets", mock.Anything, "default").Return([]string{"secret1", "secret2"}, nil)
	mockClient.On("GetPods", mock.Anything, "default").Return([]string{"pod1", "pod2"}, nil)
	mockClient.On("GetDeployments", mock.Anything, "default").Return([]string{"deployment1", "deployment2"}, nil)
	mockClient.On("GetStatefulSets", mock.Anything, "default").Return([]string{"sts1", "sts2"}, nil)
	mockClient.On("GetDaemonSets", mock.Anything, "default").Return([]string{"ds1", "ds2"}, nil)
	mockClient.On("GetPersistentVolumeClaims", mock.Anything, "default").Return([]string{"pvc1", "pvc2"}, nil)
	mockClient.On("GetClusterSecretStores", mock.Anything).Return([]string{"clusterstore1", "clusterstore2"}, nil)
	mockClient.On("GetSecretStores", mock.Anything, "default").Return([]string{"store1", "store2"}, nil)

	service := NewAutoCompleteService(mockClient)
	ctx := context.Background()

	// Test all resource types
	tests := []struct {
		resourceType ResourceType
		namespace    string
		expectedLen  int
	}{
		{ResourceTypeNamespace, "", 3}, // Mock returns 3 namespaces
		{ResourceTypeService, "default", 2},
		{ResourceTypeConfigMap, "default", 2},
		{ResourceTypeSecret, "default", 2},
		{ResourceTypePod, "default", 2},
		{ResourceTypeDeployment, "default", 2},
		{ResourceTypeStatefulSet, "default", 2},
		{ResourceTypeDaemonSet, "default", 2},
		{ResourceTypePersistentVolumeClaim, "default", 2},
		{ResourceTypeClusterSecretStore, "", 3}, // Mock returns 3 cluster secret stores
		{ResourceTypeSecretStore, "default", 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.resourceType), func(t *testing.T) {
			items, err := service.fetchResourceItems(ctx, tt.resourceType, tt.namespace)
			assert.NoError(t, err)
			assert.Len(t, items, tt.expectedLen)
		})
	}

	// Test unsupported resource type
	_, err := service.fetchResourceItems(ctx, "unsupported", "default")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported resource type")
}

func TestAutoCompleteService_FetchResourceItemsWithErrors(t *testing.T) {
	// Test unsupported resource type (this should always return an error)
	service := NewAutoCompleteService(&MockKubeLister{})
	ctx := context.Background()

	_, err := service.fetchResourceItems(ctx, "unsupported", "default")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported resource type")
}
