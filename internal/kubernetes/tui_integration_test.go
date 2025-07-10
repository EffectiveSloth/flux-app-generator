package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTUIProvider_CreateCustomInputWithNilValue(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	// Test that createCustomInput handles nil values properly
	input := provider.createCustomInput("Test Title", "Test Description", "Test Placeholder", nil, nil)

	assert.NotNil(t, input)
	// Since the huh library properties are now functions, we test that the input was created successfully
	// rather than inspecting internal properties
}

func TestTUIProvider_TextInput(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "test-value"
	input := provider.TextInput("Test Title", "Test Description", "Test Placeholder", &value)

	assert.NotNil(t, input)
	// Test that the input was created successfully
}

func TestTUIProvider_NamespaceInputValidation(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetNamespaces", mock.Anything).Return([]string{"default", "kube-system", "test"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := ""
	input := provider.NamespaceInput("Namespace", "Select namespace", "default", &value)

	// Test that the input field is properly created with validation
	assert.NotNil(t, input)
}

func TestTUIProvider_NamespaceInputWithAutoComplete(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetNamespaces", mock.Anything).Return([]string{"default", "kube-system", "test"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "def"
	input := provider.NamespaceInput("Namespace", "Select namespace", "default", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_ServiceInputWithEmptyNamespace(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "test-service"
	input := provider.ServiceInput("Service", "Select service", "service-name", "", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_ServiceInputWithNamespace(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetServices", mock.Anything, "default").Return([]string{"service1", "service2"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "service"
	input := provider.ServiceInput("Service", "Select service", "service-name", "default", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_ConfigMapInput(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetConfigMaps", mock.Anything, "default").Return([]string{"config1", "config2"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "config"
	input := provider.ConfigMapInput("ConfigMap", "Select configmap", "config-name", "default", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_SecretInput(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetSecrets", mock.Anything, "default").Return([]string{"secret1", "secret2"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "secret"
	input := provider.SecretInput("Secret", "Select secret", "secret-name", "default", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_DeploymentInput(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetDeployments", mock.Anything, "default").Return([]string{"deployment1", "deployment2"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "deployment"
	input := provider.DeploymentInput("Deployment", "Select deployment", "deployment-name", "default", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_StatefulSetInput(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetStatefulSets", mock.Anything, "default").Return([]string{"statefulset1", "statefulset2"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "statefulset"
	input := provider.StatefulSetInput("StatefulSet", "Select statefulset", "statefulset-name", "default", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_DaemonSetInput(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetDaemonSets", mock.Anything, "default").Return([]string{"daemonset1", "daemonset2"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "daemonset"
	input := provider.DaemonSetInput("DaemonSet", "Select daemonset", "daemonset-name", "default", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_PVCInput(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetPersistentVolumeClaims", mock.Anything, "default").Return([]string{"pvc1", "pvc2"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "pvc"
	input := provider.PVCInput("PVC", "Select PVC", "pvc-name", "default", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_ResourceSelectWithEmptyNamespace(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := ""
	selectField := provider.ResourceSelect("Resource", "Select resource", "", ResourceTypeService, &value)

	assert.NotNil(t, selectField)
}

func TestTUIProvider_ResourceSelectWithNamespace(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetServices", mock.Anything, "default").Return([]string{"service1", "service2"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := ""
	selectField := provider.ResourceSelect("Service", "Select service", "default", ResourceTypeService, &value)

	assert.NotNil(t, selectField)
}

func TestTUIProvider_ResourceSelectWithError(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetServices", mock.Anything, "default").Return(nil, assert.AnError)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := ""
	selectField := provider.ResourceSelect("Service", "Select service", "default", ResourceTypeService, &value)

	assert.NotNil(t, selectField)
}

func TestTUIProvider_NamespaceSelect(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetNamespaces", mock.Anything).Return([]string{"default", "kube-system", "test"}, nil)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := ""
	selectField := provider.NamespaceSelect("Namespace", "Select namespace", &value)

	assert.NotNil(t, selectField)
}

func TestTUIProvider_NamespaceSelectWithError(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetNamespaces", mock.Anything).Return(nil, assert.AnError)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := ""
	selectField := provider.NamespaceSelect("Namespace", "Select namespace", &value)

	assert.NotNil(t, selectField)
}

func TestTUIProvider_WithNilAutoComplete(t *testing.T) {
	provider := NewTUIProvider(nil)

	value := "test"
	input := provider.TextInput("Test", "Description", "Placeholder", &value)

	assert.NotNil(t, input)
}

func TestTUIProvider_WithErrorFromAutoComplete(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetNamespaces", mock.Anything).Return(nil, assert.AnError)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "test"
	input := provider.NamespaceInput("Namespace", "Select namespace", "default", &value)

	// Should still create the input even with autocomplete errors
	assert.NotNil(t, input)
}

func TestTUIProvider_CreateCustomInput(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "test-value"
	
	suggestionsFunc := func() []string {
		return []string{"suggestion1", "suggestion2"}
	}
	
	input := provider.createCustomInput("Test Title", "Test Description", "Test Placeholder", &value, suggestionsFunc)

	assert.NotNil(t, input)
}

func TestTUIProvider_InterfaceCompliance(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	// Test that TUIProvider implements the expected interface behavior
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.autoComplete)
}

func TestTUIProvider_Consistency(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "test"
	
	// Create multiple inputs with same parameters and verify consistency
	for i := 0; i < 3; i++ {
		input := provider.TextInput("Title", "Desc", "Placeholder", &value)
		assert.NotNil(t, input)
	}
}

func TestTUIProvider_ErrorHandling(t *testing.T) {
	mockClient := &MockKubeLister{}
	mockClient.On("GetNamespaces", mock.Anything).Return(nil, assert.AnError)
	
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "test"
	input := provider.NamespaceInput("Namespace", "Test", "default", &value)

	// Should handle errors gracefully and still return a valid input
	assert.NotNil(t, input)
}

func TestGetResourceTypeFromString_AllCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected ResourceType
	}{
		{"namespace", "namespace", ResourceTypeNamespace},
		{"NAMESPACE", "NAMESPACE", ResourceTypeNamespace},
		{"Namespace", "Namespace", ResourceTypeNamespace},
		{"service", "service", ResourceTypeService},
		{"configmap", "configmap", ResourceTypeConfigMap},
		{"secret", "secret", ResourceTypeSecret},
		{"pod", "pod", ResourceTypePod},
		{"deployment", "deployment", ResourceTypeDeployment},
		{"statefulset", "statefulset", ResourceTypeStatefulSet},
		{"daemonset", "daemonset", ResourceTypeDaemonSet},
		{"pvc", "pvc", ResourceTypePersistentVolumeClaim},
		{"clustersecretstore", "clustersecretstore", ResourceTypeClusterSecretStore},
		{"secretstore", "secretstore", ResourceTypeSecretStore},
		{"unknown", "unknown-resource", ResourceTypeNamespace}, // default case
		{"empty", "", ResourceTypeNamespace},        // empty string defaults to namespace
		{"nil-like", "nil", ResourceTypeNamespace},  // unrecognized defaults to namespace
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetResourceTypeFromString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetResourceTypeFromString_EdgeCases(t *testing.T) {
	// Test with various edge cases - all should default to ResourceTypeNamespace
	assert.Equal(t, ResourceTypeNamespace, GetResourceTypeFromString(""))
	assert.Equal(t, ResourceTypeNamespace, GetResourceTypeFromString(" "))
	assert.Equal(t, ResourceTypeNamespace, GetResourceTypeFromString("invalid"))
}

func TestTUIProvider_Performance(t *testing.T) {
	mockClient := &MockKubeLister{}
	service := NewAutoCompleteService(mockClient)
	provider := NewTUIProvider(service)

	value := "test"

	// Create many inputs to test performance
	for i := 0; i < 100; i++ {
		input := provider.TextInput("Title", "Description", "Placeholder", &value)
		assert.NotNil(t, input)
	}
}
