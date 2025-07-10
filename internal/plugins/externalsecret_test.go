package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewExternalSecretPlugin(t *testing.T) {
	plugin := NewExternalSecretPlugin()

	// Test basic properties
	if plugin.Name() != "externalsecret" {
		t.Errorf("expected name 'externalsecret', got '%s'", plugin.Name())
	}

	expectedDesc := "Generates ExternalSecret resources for managing secrets from external secret stores"
	if plugin.Description() != expectedDesc {
		t.Errorf("expected description '%s', got '%s'", expectedDesc, plugin.Description())
	}

	// Test variables
	variables := plugin.Variables()
	expectedVariables := []string{"name", "secret_store_type", "secret_store_name", "secret_key", "target_secret_name", "refresh_interval"}

	if len(variables) != len(expectedVariables) {
		t.Errorf("expected %d variables, got %d", len(expectedVariables), len(variables))
	}

	variableNames := make(map[string]bool)
	for _, v := range variables {
		variableNames[v.Name] = true
	}

	for _, expectedName := range expectedVariables {
		if !variableNames[expectedName] {
			t.Errorf("expected variable '%s' not found", expectedName)
		}
	}

	// Test template contains expected content
	template := plugin.Template()
	expectedTemplateContent := []string{
		"apiVersion: external-secrets.io/v1beta1",
		"kind: ExternalSecret",
		"metadata:",
		"spec:",
		"secretStoreRef:",
		"dataFrom:",
		"target:",
	}

	for _, content := range expectedTemplateContent {
		if !strings.Contains(template, content) {
			t.Errorf("template should contain '%s'", content)
		}
	}

	// Test file path template
	expectedFilePath := "dependencies/external-secret-{{.target_secret_name}}.yaml"
	if plugin.FilePath() != expectedFilePath {
		t.Errorf("expected file path '%s', got '%s'", expectedFilePath, plugin.FilePath())
	}
}

func TestExternalSecretPlugin_Variables(t *testing.T) {
	plugin := NewExternalSecretPlugin()
	variables := plugin.Variables()

	// Test specific variable properties
	variableMap := make(map[string]Variable)
	for _, v := range variables {
		variableMap[v.Name] = v
	}

	// Test name variable
	if nameVar, exists := variableMap["name"]; exists {
		if nameVar.Type != VariableTypeText {
			t.Errorf("name variable should be text type")
		}
		if !nameVar.Required {
			t.Errorf("name variable should be required")
		}
	} else {
		t.Errorf("name variable not found")
	}

	// Test secret_store_type variable
	if storeTypeVar, exists := variableMap["secret_store_type"]; exists {
		if storeTypeVar.Type != VariableTypeSelect {
			t.Errorf("secret_store_type variable should be select type")
		}
		if !storeTypeVar.Required {
			t.Errorf("secret_store_type variable should be required")
		}
		if len(storeTypeVar.Options) != 2 {
			t.Errorf("secret_store_type should have 2 options, got %d", len(storeTypeVar.Options))
		}

		// Check options
		optionValues := make(map[string]bool)
		for _, option := range storeTypeVar.Options {
			if valueStr, ok := option.Value.(string); ok {
				optionValues[valueStr] = true
			}
		}

		if !optionValues["ClusterSecretStore"] || !optionValues["SecretStore"] {
			t.Errorf("secret_store_type should have ClusterSecretStore and SecretStore options")
		}

		if storeTypeVar.Default != "ClusterSecretStore" {
			t.Errorf("secret_store_type default should be ClusterSecretStore, got %v", storeTypeVar.Default)
		}
	} else {
		t.Errorf("secret_store_type variable not found")
	}

	// Test refresh_interval variable
	if intervalVar, exists := variableMap["refresh_interval"]; exists {
		if intervalVar.Type != VariableTypeSelect {
			t.Errorf("refresh_interval variable should be select type")
		}
		if intervalVar.Required {
			t.Errorf("refresh_interval variable should not be required")
		}
		if intervalVar.Default != "60m" {
			t.Errorf("refresh_interval default should be 60m, got %v", intervalVar.Default)
		}
		if len(intervalVar.Options) == 0 {
			t.Errorf("refresh_interval should have options")
		}
	} else {
		t.Errorf("refresh_interval variable not found")
	}
}

func TestExternalSecretPlugin_Validate(t *testing.T) {
	plugin := NewExternalSecretPlugin()

	tests := []struct {
		name        string
		values      map[string]interface{}
		expectError bool
		errorText   string
	}{
		{
			name: "valid configuration",
			values: map[string]interface{}{
				"name":               "test-secret",
				"secret_store_type":  "ClusterSecretStore",
				"secret_store_name":  "aws-secret-store",
				"secret_key":         "datadog-api-key",
				"target_secret_name": "datadog-secret-api",
				"refresh_interval":   "60m",
			},
			expectError: false,
		},
		{
			name: "missing required name",
			values: map[string]interface{}{
				"secret_store_type":  "ClusterSecretStore",
				"secret_store_name":  "aws-secret-store",
				"secret_key":         "datadog-api-key",
				"target_secret_name": "datadog-secret-api",
			},
			expectError: true,
			errorText:   "name",
		},
		{
			name: "invalid secret store type",
			values: map[string]interface{}{
				"name":               "test-secret",
				"secret_store_type":  "InvalidStore",
				"secret_store_name":  "aws-secret-store",
				"secret_key":         "datadog-api-key",
				"target_secret_name": "datadog-secret-api",
			},
			expectError: true,
			errorText:   "not one of the allowed options",
		},
		{
			name: "missing all required fields",
			values: map[string]interface{}{
				"refresh_interval": "30m",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.Validate(tt.values)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorText != "" && !strings.Contains(err.Error(), tt.errorText) {
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

func TestExternalSecretPlugin_GenerateFile(t *testing.T) {
	plugin := NewExternalSecretPlugin()

	tempDir := t.TempDir()
	appDir := filepath.Join(tempDir, "test-app")

	values := map[string]interface{}{
		"name":               "datadog-secret-api",
		"secret_store_type":  "ClusterSecretStore",
		"secret_store_name":  "aws-secret-store",
		"secret_key":         "datadog-api-key",
		"target_secret_name": "datadog-secret-api",
		"refresh_interval":   "60m",
	}

	err := plugin.GenerateFile(values, appDir, "coderamp-system")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if file was created at expected location
	expectedPath := filepath.Join(appDir, "dependencies", "external-secret-datadog-secret-api.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file to be created at %s", expectedPath)
	}

	// Check file content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Check for expected YAML structure
	expectedContent := []string{
		"apiVersion: external-secrets.io/v1beta1",
		"kind: ExternalSecret",
		"metadata:",
		"name: datadog-secret-api",
		"namespace: coderamp-system",
		"spec:",
		"secretStoreRef:",
		"kind: ClusterSecretStore",
		"name: aws-secret-store",
		"dataFrom:",
		"extract:",
		"key: datadog-api-key",
		"refreshInterval: 60m",
		"target:",
		"creationPolicy: Owner",
		"name: datadog-secret-api",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("generated file should contain '%s'\nGenerated content:\n%s", expected, contentStr)
		}
	}

	// Ensure file ends with newline
	if !strings.HasSuffix(contentStr, "\n") {
		t.Errorf("generated file should end with newline")
	}
}

func TestExternalSecretPlugin_GenerateFile_SecretStore(t *testing.T) {
	plugin := NewExternalSecretPlugin()

	tempDir := t.TempDir()
	appDir := filepath.Join(tempDir, "test-app")

	// Test with SecretStore instead of ClusterSecretStore
	values := map[string]interface{}{
		"name":               "vault-secret",
		"secret_store_type":  "SecretStore",
		"secret_store_name":  "vault-store",
		"secret_key":         "app-secret",
		"target_secret_name": "app-secret",
		"refresh_interval":   "30m",
	}

	err := plugin.GenerateFile(values, appDir, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file content for SecretStore
	expectedPath := filepath.Join(appDir, "dependencies", "external-secret-app-secret.yaml")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "kind: SecretStore") {
		t.Errorf("should contain 'kind: SecretStore', got: %s", contentStr)
	}

	if !strings.Contains(contentStr, "name: vault-store") {
		t.Errorf("should contain 'name: vault-store', got: %s", contentStr)
	}

	if !strings.Contains(contentStr, "refreshInterval: 30m") {
		t.Errorf("should contain 'refreshInterval: 30m', got: %s", contentStr)
	}
}
