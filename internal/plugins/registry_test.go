package plugins

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatalf("expected registry to be created")
	}

	// Should have at least the built-in ExternalSecret plugin registered
	if registry.Count() == 0 {
		t.Errorf("expected at least one built-in plugin to be registered")
	}

	// Check if ExternalSecret plugin is registered
	if !registry.Exists("externalsecret") {
		t.Errorf("expected externalsecret plugin to be registered by default")
	}

	// Check if ImageUpdate plugin is registered
	if !registry.Exists("imageupdate") {
		t.Errorf("expected imageupdate plugin to be registered by default")
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	initialCount := registry.Count()

	// Create a test plugin
	testPlugin := &BasePlugin{
		name:        "test-plugin",
		description: "Test plugin for testing",
		variables:   []Variable{},
		template:    "test: value",
		filePath:    "test.yaml",
	}

	// Test successful registration
	err := registry.Register(testPlugin)
	if err != nil {
		t.Errorf("unexpected error registering plugin: %v", err)
	}

	if registry.Count() != initialCount+1 {
		t.Errorf("expected count to increase by 1, got %d", registry.Count())
	}

	if !registry.Exists("test-plugin") {
		t.Errorf("expected test-plugin to be registered")
	}

	// Test registering nil plugin
	err = registry.Register(nil)
	if err == nil {
		t.Errorf("expected error when registering nil plugin")
	}

	// Test registering plugin with empty name
	emptyNamePlugin := &BasePlugin{
		name: "",
	}
	err = registry.Register(emptyNamePlugin)
	if err == nil {
		t.Errorf("expected error when registering plugin with empty name")
	}

	// Test registering duplicate plugin
	err = registry.Register(testPlugin)
	if err == nil {
		t.Errorf("expected error when registering duplicate plugin")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	// Test getting existing plugin
	plugin, exists := registry.Get("externalsecret")
	if !exists {
		t.Errorf("expected externalsecret plugin to exist")
	}
	if plugin == nil {
		t.Errorf("expected plugin to not be nil")
	}
	if plugin.Name() != "externalsecret" {
		t.Errorf("expected plugin name to be 'externalsecret', got '%s'", plugin.Name())
	}

	// Test getting non-existent plugin
	plugin, exists = registry.Get("non-existent")
	if exists {
		t.Errorf("expected non-existent plugin to not exist")
	}
	if plugin != nil {
		t.Errorf("expected plugin to be nil for non-existent plugin")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	plugins := registry.List()
	if len(plugins) == 0 {
		t.Errorf("expected at least one plugin in list")
	}

	// Add a test plugin
	testPlugin := &BasePlugin{
		name:        "test-plugin-2",
		description: "Another test plugin",
		variables:   []Variable{},
		template:    "test: value",
		filePath:    "test.yaml",
	}

	if err := registry.Register(testPlugin); err != nil {
		t.Fatalf("failed to register test plugin: %v", err)
	}

	pluginsAfter := registry.List()
	if len(pluginsAfter) != len(plugins)+1 {
		t.Errorf("expected plugin list to grow by 1, got %d -> %d", len(plugins), len(pluginsAfter))
	}

	// Check that all plugins are unique
	pluginNames := make(map[string]bool)
	for _, plugin := range pluginsAfter {
		name := plugin.Name()
		if pluginNames[name] {
			t.Errorf("duplicate plugin name '%s' in list", name)
		}
		pluginNames[name] = true
	}
}

func TestRegistry_GetNames(t *testing.T) {
	registry := NewRegistry()

	names := registry.GetNames()
	if len(names) == 0 {
		t.Errorf("expected at least one plugin name")
	}

	// Check that externalsecret is in the list
	found := false
	for _, name := range names {
		if name == "externalsecret" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected 'externalsecret' to be in plugin names list")
	}

	// Add a test plugin and verify it appears in names
	testPlugin := &BasePlugin{
		name:        "test-plugin-3",
		description: "Test plugin for names test",
		variables:   []Variable{},
		template:    "test: value",
		filePath:    "test.yaml",
	}

	if err := registry.Register(testPlugin); err != nil {
		t.Fatalf("failed to register test plugin: %v", err)
	}

	namesAfter := registry.GetNames()
	if len(namesAfter) != len(names)+1 {
		t.Errorf("expected names list to grow by 1")
	}

	foundTest := false
	for _, name := range namesAfter {
		if name == "test-plugin-3" {
			foundTest = true
			break
		}
	}

	if !foundTest {
		t.Errorf("expected 'test-plugin-3' to be in plugin names list")
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()

	initialCount := registry.Count()
	if initialCount == 0 {
		t.Errorf("expected at least one plugin to be registered initially")
	}

	// Add plugins and verify count increases
	testPlugin1 := &BasePlugin{
		name:        "count-test-1",
		description: "Test plugin 1",
		variables:   []Variable{},
		template:    "test: value",
		filePath:    "test.yaml",
	}

	testPlugin2 := &BasePlugin{
		name:        "count-test-2",
		description: "Test plugin 2",
		variables:   []Variable{},
		template:    "test: value",
		filePath:    "test.yaml",
	}

	if err := registry.Register(testPlugin1); err != nil {
		t.Fatalf("failed to register test plugin 1: %v", err)
	}
	if registry.Count() != initialCount+1 {
		t.Errorf("expected count to be %d, got %d", initialCount+1, registry.Count())
	}

	if err := registry.Register(testPlugin2); err != nil {
		t.Fatalf("failed to register test plugin 2: %v", err)
	}
	if registry.Count() != initialCount+2 {
		t.Errorf("expected count to be %d, got %d", initialCount+2, registry.Count())
	}
}

func TestRegistry_Exists(t *testing.T) {
	registry := NewRegistry()

	// Test existing plugin
	if !registry.Exists("externalsecret") {
		t.Errorf("expected externalsecret plugin to exist")
	}

	// Test non-existing plugin
	if registry.Exists("non-existent-plugin") {
		t.Errorf("expected non-existent-plugin to not exist")
	}

	// Add a plugin and test existence
	testPlugin := &BasePlugin{
		name:        "exists-test",
		description: "Test plugin for exists test",
		variables:   []Variable{},
		template:    "test: value",
		filePath:    "test.yaml",
	}

	if err := registry.Register(testPlugin); err != nil {
		t.Fatalf("failed to register test plugin: %v", err)
	}

	if !registry.Exists("exists-test") {
		t.Errorf("expected exists-test plugin to exist after registration")
	}
}

func TestRegistry_BuiltinPlugins(t *testing.T) {
	registry := NewRegistry()

	// Test that all expected built-in plugins are registered
	expectedPlugins := []string{"externalsecret", "imageupdate"}

	for _, pluginName := range expectedPlugins {
		if !registry.Exists(pluginName) {
			t.Errorf("expected built-in plugin '%s' to be registered", pluginName)
		}

		plugin, exists := registry.Get(pluginName)
		if !exists {
			t.Errorf("expected to get built-in plugin '%s'", pluginName)
		}

		if plugin == nil {
			t.Errorf("expected built-in plugin '%s' to not be nil", pluginName)
		}

		if plugin.Name() != pluginName {
			t.Errorf("expected plugin name to be '%s', got '%s'", pluginName, plugin.Name())
		}
	}
}

func TestRegistry_Integration(t *testing.T) {
	// Test full integration workflow
	registry := NewRegistry()

	// Get the externalsecret plugin
	plugin, exists := registry.Get("externalsecret")
	if !exists {
		t.Fatalf("expected externalsecret plugin to exist")
	}

	// Test plugin functionality
	variables := plugin.Variables()
	if len(variables) == 0 {
		t.Errorf("expected externalsecret plugin to have variables")
	}

	template := plugin.Template()
	if template == "" {
		t.Errorf("expected externalsecret plugin to have template")
	}

	filePath := plugin.FilePath()
	if filePath == "" {
		t.Errorf("expected externalsecret plugin to have file path")
	}

	// Test validation with valid data
	validValues := map[string]interface{}{
		"name":               "test-secret",
		"secret_store_type":  "ClusterSecretStore",
		"secret_store_name":  "test-store",
		"secret_key":         "test-key",
		"target_secret_name": "test-target",
		"refresh_interval":   "60m",
	}

	err := plugin.Validate(validValues)
	if err != nil {
		t.Errorf("expected valid values to pass validation: %v", err)
	}
}
