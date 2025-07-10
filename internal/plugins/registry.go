package plugins

import (
	"fmt"
)

// Registry manages all available plugins.
type Registry struct {
	plugins map[string]Plugin
}

// NewRegistry creates a new plugin registry with all built-in plugins registered.
func NewRegistry() *Registry {
	registry := &Registry{
		plugins: make(map[string]Plugin),
	}

	// Register built-in plugins
	registry.registerBuiltinPlugins()

	return registry
}

// registerBuiltinPlugins registers all built-in plugins.
func (r *Registry) registerBuiltinPlugins() {
	// Register the ExternalSecret plugin
	if err := r.Register(NewExternalSecretPlugin()); err != nil {
		panic(fmt.Sprintf("failed to register built-in externalsecret plugin: %v", err))
	}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("cannot register nil plugin")
	}

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin with name '%s' is already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// Get retrieves a plugin by name.
func (r *Registry) Get(name string) (Plugin, bool) {
	plugin, exists := r.plugins[name]
	return plugin, exists
}

// List returns all registered plugins.
func (r *Registry) List() []Plugin {
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// GetNames returns the names of all registered plugins.
func (r *Registry) GetNames() []string {
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered plugins.
func (r *Registry) Count() int {
	return len(r.plugins)
}

// Exists checks if a plugin with the given name is registered.
func (r *Registry) Exists(name string) bool {
	_, exists := r.plugins[name]
	return exists
}
