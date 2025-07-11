package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// VariableType represents the different types of input variables a plugin can have.
type VariableType string

const (
	// VariableTypeText represents a text input variable.
	VariableTypeText VariableType = "text"
	// VariableTypeBool represents a boolean input variable.
	VariableTypeBool VariableType = "bool"
	// VariableTypeSelect represents a select dropdown input variable.
	VariableTypeSelect VariableType = "select"
	// VariableTypeCheckbox represents a checkbox input variable.
	VariableTypeCheckbox VariableType = "checkbox"
)

// Variable defines a configurable input for a plugin.
type Variable struct {
	Name        string       `json:"name" yaml:"name"`
	Type        VariableType `json:"type" yaml:"type"`
	Description string       `json:"description" yaml:"description"`
	Required    bool         `json:"required" yaml:"required"`
	Default     interface{}  `json:"default,omitempty" yaml:"default,omitempty"`
	Options     []Option     `json:"options,omitempty" yaml:"options,omitempty"` // For select type
}

// Option represents a choice for select-type variables.
type Option struct {
	Label string      `json:"label" yaml:"label"`
	Value interface{} `json:"value" yaml:"value"`
}

// PluginConfig holds the runtime configuration for a plugin instance.
type PluginConfig struct {
	PluginName string                 `json:"plugin_name" yaml:"plugin_name"`
	Values     map[string]interface{} `json:"values" yaml:"values"`
}

// Plugin defines the interface that all plugins must implement.
type Plugin interface {
	// Name returns the unique name/identifier of the plugin.
	Name() string

	// Description returns a human-readable description of the plugin.
	Description() string

	// Variables returns the list of configurable variables for this plugin.
	Variables() []Variable

	// Template returns the YAML template string for generating the output file.
	Template() string

	// FilePath returns the template for the output file path (can include variables).
	FilePath() string

	// Validate checks if the provided values are valid for this plugin.
	Validate(values map[string]interface{}) error

	// GenerateFile creates the output file using the template and values.
	GenerateFile(values map[string]interface{}, appDir, namespace string) error
}

// CustomConfigPlugin defines an interface for plugins that need custom configuration collection.
// This allows plugins to implement their own multi-screen or advanced configuration logic.
type CustomConfigPlugin interface {
	Plugin

	// CollectCustomConfig handles custom configuration collection beyond standard variables.
	// This method is called after standard variables are collected.
	// It should modify the values map to include any additional configuration.
	CollectCustomConfig(values map[string]interface{}) error
}

// BasePlugin provides common functionality for plugins.
type BasePlugin struct {
	name        string
	description string
	variables   []Variable
	template    string
	filePath    string
}

// Name returns the plugin name.
func (p *BasePlugin) Name() string {
	return p.name
}

// Description returns the plugin description.
func (p *BasePlugin) Description() string {
	return p.description
}

// Variables returns the plugin variables.
func (p *BasePlugin) Variables() []Variable {
	return p.variables
}

// Template returns the plugin template.
func (p *BasePlugin) Template() string {
	return p.template
}

// FilePath returns the plugin file path template.
func (p *BasePlugin) FilePath() string {
	return p.filePath
}

// Validate performs basic validation on the provided values.
func (p *BasePlugin) Validate(values map[string]interface{}) error {
	for _, variable := range p.variables {
		if variable.Required {
			if _, exists := values[variable.Name]; !exists {
				return &ValidationError{
					Variable: variable.Name,
					Message:  "required variable is missing",
				}
			}
		}

		// Type-specific validation
		if value, exists := values[variable.Name]; exists && value != nil {
			if err := p.validateVariableType(&variable, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateVariableType validates a single variable value against its type.
func (p *BasePlugin) validateVariableType(variable *Variable, value interface{}) error {
	switch variable.Type {
	case VariableTypeBool, VariableTypeCheckbox:
		if _, ok := value.(bool); !ok {
			return &ValidationError{
				Variable: variable.Name,
				Message:  "value must be a boolean",
			}
		}
	case VariableTypeText:
		if _, ok := value.(string); !ok {
			return &ValidationError{
				Variable: variable.Name,
				Message:  "value must be a string",
			}
		}
	case VariableTypeSelect:
		// Check if value is one of the allowed options
		found := false
		for _, option := range variable.Options {
			if option.Value == value {
				found = true
				break
			}
		}
		if !found {
			return &ValidationError{
				Variable: variable.Name,
				Message:  "value is not one of the allowed options",
			}
		}
	}
	return nil
}

// GenerateFile creates the output file using the template and values.
func (p *BasePlugin) GenerateFile(values map[string]interface{}, appDir, namespace string) error {
	// Create template data combining values with namespace
	templateData := make(map[string]interface{})
	for k, v := range values {
		templateData[k] = v
	}
	templateData["Namespace"] = namespace

	// Parse the file path template
	pathTmpl, err := template.New("filepath").Parse(p.filePath)
	if err != nil {
		return &TemplateError{
			Plugin:  p.name,
			Type:    "filepath",
			Message: err.Error(),
		}
	}

	var pathBuf strings.Builder
	if err := pathTmpl.Execute(&pathBuf, templateData); err != nil {
		return &TemplateError{
			Plugin:  p.name,
			Type:    "filepath",
			Message: err.Error(),
		}
	}

	outputPath := filepath.Join(appDir, pathBuf.String())

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return &FileError{
			Plugin:    p.name,
			Operation: "create_directory",
			Path:      filepath.Dir(outputPath),
			Message:   err.Error(),
		}
	}

	// Parse and execute the YAML template
	tmpl, err := template.New("plugin").Parse(p.template)
	if err != nil {
		return &TemplateError{
			Plugin:  p.name,
			Type:    "yaml",
			Message: err.Error(),
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return &FileError{
			Plugin:    p.name,
			Operation: "create_file",
			Path:      outputPath,
			Message:   err.Error(),
		}
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file %s: %v\n", outputPath, closeErr)
		}
	}()

	if err := tmpl.Execute(file, templateData); err != nil {
		return &TemplateError{
			Plugin:  p.name,
			Type:    "yaml",
			Message: err.Error(),
		}
	}

	// Ensure file ends with a newline
	if _, err := file.WriteString("\n"); err != nil {
		return &FileError{
			Plugin:    p.name,
			Operation: "write_newline",
			Path:      outputPath,
			Message:   err.Error(),
		}
	}

	return nil
}

// ValidationError represents a plugin validation error.
type ValidationError struct {
	Variable string
	Message  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for variable '%s': %s", e.Variable, e.Message)
}

// TemplateError represents a template processing error.
type TemplateError struct {
	Plugin  string
	Type    string // "yaml" or "filepath"
	Message string
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf("template error in plugin '%s' (%s): %s", e.Plugin, e.Type, e.Message)
}

// FileError represents a file operation error.
type FileError struct {
	Plugin    string
	Operation string
	Path      string
	Message   string
}

func (e *FileError) Error() string {
	return fmt.Sprintf("file error in plugin '%s' during %s for path '%s': %s", e.Plugin, e.Operation, e.Path, e.Message)
}
