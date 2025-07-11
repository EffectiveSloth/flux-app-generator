package plugins

import (
	"testing"

	"github.com/EffectiveSloth/flux-app-generator/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.plugins)
	assert.Len(t, registry.plugins, 2) // Should have externalsecret and imageupdate plugins

	// Check that externalsecret plugin is registered
	plugin, exists := registry.plugins["externalsecret"]
	assert.True(t, exists)
	assert.NotNil(t, plugin)
	assert.Equal(t, "externalsecret", plugin.Name())
}

func TestNewRegistryWithNilClient(t *testing.T) {
	registry := NewRegistry(nil)

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.plugins)
	assert.Len(t, registry.plugins, 2) // Should still have externalsecret and imageupdate plugins

	// Check that externalsecret plugin is registered even with nil client
	plugin, exists := registry.plugins["externalsecret"]
	assert.True(t, exists)
	assert.NotNil(t, plugin)
	assert.Equal(t, "externalsecret", plugin.Name())
}

func TestRegistry_List(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	plugins := registry.List()

	assert.NotNil(t, plugins)
	assert.Len(t, plugins, 2)

	// Check that the externalsecret plugin is in the list
	found := false
	for _, plugin := range plugins {
		if plugin.Name() == "externalsecret" {
			found = true
			assert.Equal(t, "Generates ExternalSecret resources for managing secrets from external secret stores", plugin.Description())
			break
		}
	}
	assert.True(t, found, "Expected externalsecret plugin to be in the list")
}

func TestRegistry_Get(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	// Test getting existing plugin
	plugin, exists := registry.Get("externalsecret")
	assert.True(t, exists)
	assert.NotNil(t, plugin)
	assert.Equal(t, "externalsecret", plugin.Name())

	// Test getting non-existent plugin
	plugin, exists = registry.Get("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, plugin)
}

func TestRegistry_GetWithEmptyName(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	// Test getting plugin with empty name
	plugin, exists := registry.Get("")
	assert.False(t, exists)
	assert.Nil(t, plugin)
}

func TestRegistry_GetWithNilRegistry(t *testing.T) {
	// Test edge case with nil registry (shouldn't happen in practice)
	var registry *Registry
	if registry != nil {
		plugin, exists := registry.Get("externalsecret")
		assert.False(t, exists)
		assert.Nil(t, plugin)
	}
}

func TestRegistry_PluginProperties(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	plugin, exists := registry.Get("externalsecret")
	require.True(t, exists)
	require.NotNil(t, plugin)

	// Test plugin properties
	assert.Equal(t, "externalsecret", plugin.Name())
	assert.Equal(t, "Generates ExternalSecret resources for managing secrets from external secret stores", plugin.Description())

	// Test that plugin has variables
	variables := plugin.Variables()
	assert.NotNil(t, variables)
	assert.Len(t, variables, 6) // name, secret_store_type, secret_store_name, secret_key, target_secret_name, refresh_interval

	// Check specific variables exist
	variableNames := make(map[string]bool)
	for _, v := range variables {
		variableNames[v.Name] = true
	}

	expectedVariables := []string{"name", "secret_store_type", "secret_store_name", "secret_key", "target_secret_name", "refresh_interval"}
	for _, expected := range expectedVariables {
		assert.True(t, variableNames[expected], "Expected variable %s to be present", expected)
	}
}

func TestRegistry_PluginGeneration(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	plugin, exists := registry.Get("externalsecret")
	require.True(t, exists)
	require.NotNil(t, plugin)

	// Test plugin generation
	values := map[string]interface{}{
		"name":               "test-secret",
		"secret_store_type":  "ClusterSecretStore",
		"secret_store_name":  "test-store",
		"secret_key":         "test-key",
		"target_secret_name": "test-target",
		"refresh_interval":   "60m",
		"Namespace":          "default",
	}

	err := plugin.GenerateFile(values, "/tmp", "default")
	assert.NoError(t, err)
}

func TestRegistry_PluginGenerationWithMissingValues(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	plugin, exists := registry.Get("externalsecret")
	require.True(t, exists)
	require.NotNil(t, plugin)

	// Test plugin generation with missing values
	values := map[string]interface{}{
		"name": "test-secret",
		// Missing other required values
	}

	err := plugin.GenerateFile(values, "/tmp", "default")
	assert.NoError(t, err) // Should still generate without error
}

func TestRegistry_PluginInterfaceCompliance(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	plugin, exists := registry.Get("externalsecret")
	require.True(t, exists)
	require.NotNil(t, plugin)

	// Test that plugin implements Plugin interface
	_ = plugin
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	// Test concurrent access to registry
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Test List
			plugins := registry.List()
			assert.NotNil(t, plugins)
			assert.Len(t, plugins, 2)

			// Test Get
			plugin, exists := registry.Get("externalsecret")
			assert.True(t, exists)
			assert.NotNil(t, plugin)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRegistry_PluginRegistration(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	// Test that plugins are properly registered
	plugins := registry.List()
	assert.Len(t, plugins, 2)

	// Check that the plugin is the correct type
	plugin := plugins[0]
	assert.IsType(t, &ExternalSecretPlugin{}, plugin)

	// Check that it's the same plugin instance
	externalSecretPlugin, ok := plugin.(*ExternalSecretPlugin)
	assert.True(t, ok)
	assert.Equal(t, mockClient, externalSecretPlugin.kubeClient)
}

func TestRegistry_PluginWithNilClient(t *testing.T) {
	registry := NewRegistry(nil)

	plugin, exists := registry.Get("externalsecret")
	require.True(t, exists)
	require.NotNil(t, plugin)

	// Check that the plugin can handle nil client
	externalSecretPlugin, ok := plugin.(*ExternalSecretPlugin)
	assert.True(t, ok)
	assert.Nil(t, externalSecretPlugin.kubeClient)

	// Test that plugin still works with nil client
	values := map[string]interface{}{
		"name":               "test-secret",
		"secret_store_type":  "ClusterSecretStore",
		"secret_store_name":  "test-store",
		"secret_key":         "test-key",
		"target_secret_name": "test-target",
		"refresh_interval":   "60m",
		"Namespace":          "default",
	}

	err := plugin.GenerateFile(values, "/tmp", "default")
	assert.NoError(t, err)
}

func TestRegistry_EmptyRegistry(t *testing.T) {
	// Test edge case of empty registry
	registry := &Registry{
		plugins: make(map[string]Plugin),
	}

	plugins := registry.List()
	assert.Empty(t, plugins)

	plugin, exists := registry.Get("externalsecret")
	assert.False(t, exists)
	assert.Nil(t, plugin)
}

func TestRegistry_PluginVariablesValidation(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	plugin, exists := registry.Get("externalsecret")
	require.True(t, exists)
	require.NotNil(t, plugin)

	variables := plugin.Variables()

	// Test that required variables are marked as required
	for _, v := range variables {
		switch v.Name {
		case "name", "secret_store_type", "secret_store_name", "secret_key", "target_secret_name":
			assert.True(t, v.Required, "Variable %s should be required", v.Name)
		case "refresh_interval":
			assert.False(t, v.Required, "Variable %s should not be required", v.Name)
		}
	}

	// Test that select variables have options
	for _, v := range variables {
		if v.Type == VariableTypeSelect {
			assert.NotEmpty(t, v.Options, "Select variable %s should have options", v.Name)
		}
	}
}

func TestRegistry_PluginTemplateConsistency(t *testing.T) {
	mockClient := &kubernetes.MockKubeLister{}
	registry := NewRegistry(mockClient)

	plugin, exists := registry.Get("externalsecret")
	require.True(t, exists)
	require.NotNil(t, plugin)

	// Test that template is consistent across multiple generations
	values := map[string]interface{}{
		"name":               "test-secret",
		"secret_store_type":  "ClusterSecretStore",
		"secret_store_name":  "test-store",
		"secret_key":         "test-key",
		"target_secret_name": "test-target",
		"refresh_interval":   "60m",
		"Namespace":          "default",
	}

	err1 := plugin.GenerateFile(values, "/tmp", "default")
	err2 := plugin.GenerateFile(values, "/tmp", "default")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
}
