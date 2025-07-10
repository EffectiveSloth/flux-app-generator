package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVariableType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		varType  VariableType
		expected string
	}{
		{"text type", VariableTypeText, "text"},
		{"bool type", VariableTypeBool, "bool"},
		{"select type", VariableTypeSelect, "select"},
		{"checkbox type", VariableTypeCheckbox, "checkbox"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.varType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.varType))
			}
		})
	}
}

func TestBasePlugin_Methods(t *testing.T) {
	variables := []Variable{
		{
			Name:        "test_var",
			Type:        VariableTypeText,
			Description: "Test variable",
			Required:    true,
		},
	}

	plugin := &BasePlugin{
		name:        "test-plugin",
		description: "Test plugin description",
		variables:   variables,
		template:    "test: {{.test_var}}",
		filePath:    "test/{{.test_var}}.yaml",
	}

	// Test getters
	if plugin.Name() != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got '%s'", plugin.Name())
	}

	if plugin.Description() != "Test plugin description" {
		t.Errorf("expected description 'Test plugin description', got '%s'", plugin.Description())
	}

	if len(plugin.Variables()) != 1 {
		t.Errorf("expected 1 variable, got %d", len(plugin.Variables()))
	}

	if plugin.Template() != "test: {{.test_var}}" {
		t.Errorf("expected template 'test: {{.test_var}}', got '%s'", plugin.Template())
	}

	if plugin.FilePath() != "test/{{.test_var}}.yaml" {
		t.Errorf("expected file path 'test/{{.test_var}}.yaml', got '%s'", plugin.FilePath())
	}
}

func TestBasePlugin_Validate(t *testing.T) {
	variables := []Variable{
		{
			Name:     "required_text",
			Type:     VariableTypeText,
			Required: true,
		},
		{
			Name:     "optional_text",
			Type:     VariableTypeText,
			Required: false,
		},
		{
			Name:     "required_bool",
			Type:     VariableTypeBool,
			Required: true,
		},
		{
			Name:     "select_var",
			Type:     VariableTypeSelect,
			Required: true,
			Options: []Option{
				{Label: "Option 1", Value: "value1"},
				{Label: "Option 2", Value: "value2"},
			},
		},
	}

	plugin := &BasePlugin{
		name:      "test-plugin",
		variables: variables,
	}

	tests := []struct {
		name        string
		values      map[string]interface{}
		expectError bool
		errorText   string
	}{
		{
			name: "valid values",
			values: map[string]interface{}{
				"required_text": "test",
				"required_bool": true,
				"select_var":    "value1",
			},
			expectError: false,
		},
		{
			name: "missing required field",
			values: map[string]interface{}{
				"required_bool": true,
				"select_var":    "value1",
			},
			expectError: true,
			errorText:   "required_text",
		},
		{
			name: "invalid type for text field",
			values: map[string]interface{}{
				"required_text": 123,
				"required_bool": true,
				"select_var":    "value1",
			},
			expectError: true,
			errorText:   "must be a string",
		},
		{
			name: "invalid type for bool field",
			values: map[string]interface{}{
				"required_text": "test",
				"required_bool": "not_bool",
				"select_var":    "value1",
			},
			expectError: true,
			errorText:   "must be a boolean",
		},
		{
			name: "invalid select option",
			values: map[string]interface{}{
				"required_text": "test",
				"required_bool": true,
				"select_var":    "invalid_value",
			},
			expectError: true,
			errorText:   "not one of the allowed options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.Validate(tt.values)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBasePlugin_GenerateFile(t *testing.T) {
	plugin := &BasePlugin{
		name:     "test-plugin",
		template: "name: {{.name}}\nnamespace: {{.Namespace}}\nvalue: {{.test_value}}",
		filePath: "test-files/{{.name}}.yaml",
	}

	tempDir := t.TempDir()
	appDir := filepath.Join(tempDir, "test-app")

	values := map[string]interface{}{
		"name":       "test-resource",
		"test_value": "example",
	}

	err := plugin.GenerateFile(values, appDir, "test-namespace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if file was created
	expectedPath := filepath.Join(appDir, "test-files", "test-resource.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file to be created at %s", expectedPath)
	}

	// Check file content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	expectedContent := "name: test-resource\nnamespace: test-namespace\nvalue: example\n"
	if string(content) != expectedContent {
		t.Errorf("expected content:\n%s\ngot:\n%s", expectedContent, string(content))
	}
}

func TestBasePlugin_GenerateFile_InvalidTemplate(t *testing.T) {
	plugin := &BasePlugin{
		name:     "test-plugin",
		template: "invalid template {{.missing_brace",
		filePath: "test.yaml",
	}

	tempDir := t.TempDir()
	values := map[string]interface{}{}

	err := plugin.GenerateFile(values, tempDir, "test-namespace")
	if err == nil {
		t.Errorf("expected error for invalid template")
	}

	if !strings.Contains(err.Error(), "template error") {
		t.Errorf("expected template error, got: %v", err)
	}
}

func TestBasePlugin_GenerateFile_InvalidFilePath(t *testing.T) {
	plugin := &BasePlugin{
		name:     "test-plugin",
		template: "test: value",
		filePath: "invalid/{{.missing_brace",
	}

	tempDir := t.TempDir()
	values := map[string]interface{}{}

	err := plugin.GenerateFile(values, tempDir, "test-namespace")
	if err == nil {
		t.Errorf("expected error for invalid file path template")
	}

	if !strings.Contains(err.Error(), "template error") {
		t.Errorf("expected template error, got: %v", err)
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Variable: "test_var",
		Message:  "test message",
	}

	expected := "validation error for variable 'test_var': test message"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestTemplateError(t *testing.T) {
	err := &TemplateError{
		Plugin:  "test-plugin",
		Type:    "yaml",
		Message: "test message",
	}

	expected := "template error in plugin 'test-plugin' (yaml): test message"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestFileError(t *testing.T) {
	err := &FileError{
		Plugin:    "test-plugin",
		Operation: "create_file",
		Path:      "/test/path",
		Message:   "test message",
	}

	expected := "file error in plugin 'test-plugin' during create_file for path '/test/path': test message"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}
