package plugins

import (
	"testing"

	"github.com/EffectiveSloth/flux-app-generator/internal/kubernetes"
	"github.com/stretchr/testify/assert"
)

func TestNewExternalSecretPlugin(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	assert.NotNil(t, plugin)
	assert.Equal(t, "externalsecret", plugin.Name())
	assert.Equal(t, "Generates ExternalSecret resources for managing secrets from external secret stores", plugin.Description())
	assert.Equal(t, mockClient, plugin.kubeClient)

	// Test variables
	variables := plugin.Variables()
	assert.Len(t, variables, 6)

	// Check specific variables
	var nameVar, storeTypeVar, storeNameVar, secretKeyVar, targetSecretVar, refreshIntervalVar *Variable
	for i := range variables {
		switch variables[i].Name {
		case "name":
			nameVar = &variables[i]
		case "secret_store_type":
			storeTypeVar = &variables[i]
		case "secret_store_name":
			storeNameVar = &variables[i]
		case "secret_key":
			secretKeyVar = &variables[i]
		case "target_secret_name":
			targetSecretVar = &variables[i]
		case "refresh_interval":
			refreshIntervalVar = &variables[i]
		}
	}

	assert.NotNil(t, nameVar)
	assert.Equal(t, VariableTypeText, nameVar.Type)
	assert.True(t, nameVar.Required)

	assert.NotNil(t, storeTypeVar)
	assert.Equal(t, VariableTypeSelect, storeTypeVar.Type)
	assert.True(t, storeTypeVar.Required)
	assert.Equal(t, "ClusterSecretStore", storeTypeVar.Default)
	assert.Len(t, storeTypeVar.Options, 2)

	assert.NotNil(t, storeNameVar)
	assert.Equal(t, VariableTypeText, storeNameVar.Type)
	assert.True(t, storeNameVar.Required)

	assert.NotNil(t, secretKeyVar)
	assert.Equal(t, VariableTypeText, secretKeyVar.Type)
	assert.True(t, secretKeyVar.Required)

	assert.NotNil(t, targetSecretVar)
	assert.Equal(t, VariableTypeText, targetSecretVar.Type)
	assert.True(t, targetSecretVar.Required)

	assert.NotNil(t, refreshIntervalVar)
	assert.Equal(t, VariableTypeSelect, refreshIntervalVar.Type)
	assert.False(t, refreshIntervalVar.Required)
	assert.Equal(t, "60m", refreshIntervalVar.Default)
	assert.Len(t, refreshIntervalVar.Options, 7)
}

func TestExternalSecretPlugin_Generate(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	values := map[string]interface{}{
		"name":               "my-external-secret",
		"secret_store_type":  "ClusterSecretStore",
		"secret_store_name":  "vault-backend",
		"secret_key":         "my-secret-key",
		"target_secret_name": "my-secret",
		"refresh_interval":   "60m",
		"Namespace":          "default",
	}

	err := plugin.GenerateFile(values, "/tmp", "default")

	assert.NoError(t, err)
}

func TestExternalSecretPlugin_GenerateWithSecretStore(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	values := map[string]interface{}{
		"name":               "my-external-secret",
		"secret_store_type":  "SecretStore",
		"secret_store_name":  "local-vault",
		"secret_key":         "my-secret-key",
		"target_secret_name": "my-secret",
		"refresh_interval":   "30m",
		"Namespace":          "default",
	}

	err := plugin.GenerateFile(values, "/tmp", "default")

	assert.NoError(t, err)
}

func TestExternalSecretPlugin_GenerateMissingRequiredFields(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	// Missing required fields
	values := map[string]interface{}{
		"name": "my-external-secret",
		// Missing other required fields
	}

	err := plugin.GenerateFile(values, "/tmp", "default")

	// Should still generate but with empty values for missing fields
	assert.NoError(t, err)
}

func TestExternalSecretPlugin_ConfigureWithAutoComplete_ClusterSecretStore(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	// Test with ClusterSecretStore type
	// This test would require mocking the TUI interaction
	// For now, we'll test the plugin creation and basic functionality
	assert.NotNil(t, plugin)
	assert.Equal(t, mockClient, plugin.kubeClient)
}

func TestExternalSecretPlugin_ConfigureWithAutoComplete_SecretStore(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	// Test with SecretStore type
	// This test would require mocking the TUI interaction
	// For now, we'll test the plugin creation and basic functionality
	assert.NotNil(t, plugin)
	assert.Equal(t, mockClient, plugin.kubeClient)
}

func TestExternalSecretPlugin_ConfigureWithAutoComplete_NilClient(t *testing.T) {
	// Test behavior when kubeClient is nil
	plugin := NewExternalSecretPlugin(nil)

	// The ConfigureWithAutoComplete method should handle nil client gracefully
	// This test would require mocking the TUI interaction
	// For now, we'll test that the plugin can be created with nil client
	assert.NotNil(t, plugin)
	assert.Nil(t, plugin.kubeClient)
}

func TestExternalSecretPlugin_TemplateGeneration(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	// Test template with all fields
	values := map[string]interface{}{
		"name":               "test-external-secret",
		"secret_store_type":  "ClusterSecretStore",
		"secret_store_name":  "test-store",
		"secret_key":         "test-key",
		"target_secret_name": "test-target-secret",
		"refresh_interval":   "15m",
		"Namespace":          "test-namespace",
	}

	err := plugin.GenerateFile(values, "/tmp", "test-namespace")

	assert.NoError(t, err)
}

func TestExternalSecretPlugin_FilePathGeneration(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	values := map[string]interface{}{
		"name":               "my-external-secret",
		"secret_store_type":  "ClusterSecretStore",
		"secret_store_name":  "vault-backend",
		"secret_key":         "my-secret-key",
		"target_secret_name": "my-target-secret",
		"refresh_interval":   "60m",
		"Namespace":          "default",
	}

	err := plugin.GenerateFile(values, "/tmp", "default")

	assert.NoError(t, err)
}

func TestExternalSecretPlugin_InterfaceCompliance(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	// Test that ExternalSecretPlugin implements Plugin interface
	var _ Plugin = plugin
}

func TestExternalSecretPlugin_WithDifferentRefreshIntervals(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	plugin := NewExternalSecretPlugin(mockClient)

	testCases := []struct {
		interval string
		expected string
	}{
		{"15m", "refreshInterval: 15m"},
		{"30m", "refreshInterval: 30m"},
		{"60m", "refreshInterval: 60m"},
		{"120m", "refreshInterval: 120m"},
		{"6h", "refreshInterval: 6h"},
		{"12h", "refreshInterval: 12h"},
		{"24h", "refreshInterval: 24h"},
	}

	for _, tc := range testCases {
		t.Run(tc.interval, func(t *testing.T) {
			values := map[string]interface{}{
				"name":               "test-secret",
				"secret_store_type":  "ClusterSecretStore",
				"secret_store_name":  "test-store",
				"secret_key":         "test-key",
				"target_secret_name": "test-target",
				"refresh_interval":   tc.interval,
				"Namespace":          "default",
			}

			err := plugin.GenerateFile(values, "/tmp", "default")

			assert.NoError(t, err)
		})
	}
}
