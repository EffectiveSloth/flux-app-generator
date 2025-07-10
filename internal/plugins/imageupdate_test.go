package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewImageUpdatePlugin(t *testing.T) {
	plugin := NewImageUpdatePlugin()

	// Test basic properties
	if plugin.Name() != "imageupdate" {
		t.Errorf("expected name 'imageupdate', got '%s'", plugin.Name())
	}

	expectedDesc := "Generates Flux image update automation resources for automatic container image updates"
	if plugin.Description() != expectedDesc {
		t.Errorf("expected description '%s', got '%s'", expectedDesc, plugin.Description())
	}

	// Test variables
	variables := plugin.Variables()
	expectedVariables := []string{
		"automation_name", "git_repository_name", "git_repository_namespace",
		"update_path", "update_strategy", "git_branch", "author_name",
		"author_email", "commit_message_template", "automation_interval",
		"image_repositories", "image_policies",
	}

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

	// Test file path template
	expectedFilePath := "update/"
	if plugin.FilePath() != expectedFilePath {
		t.Errorf("expected file path '%s', got '%s'", expectedFilePath, plugin.FilePath())
	}
}

func TestImageUpdatePlugin_Variables(t *testing.T) {
	plugin := NewImageUpdatePlugin()
	variables := plugin.Variables()

	// Test specific variable properties
	variableMap := make(map[string]Variable)
	for _, v := range variables {
		variableMap[v.Name] = v
	}

	// Test automation_name variable
	if nameVar, exists := variableMap["automation_name"]; exists {
		if nameVar.Type != VariableTypeText {
			t.Errorf("automation_name variable should be text type")
		}
		if !nameVar.Required {
			t.Errorf("automation_name variable should be required")
		}
	} else {
		t.Errorf("automation_name variable not found")
	}

	// Test update_strategy variable
	if strategyVar, exists := variableMap["update_strategy"]; exists {
		if strategyVar.Type != VariableTypeSelect {
			t.Errorf("update_strategy variable should be select type")
		}
		if !strategyVar.Required {
			t.Errorf("update_strategy variable should be required")
		}
		if len(strategyVar.Options) != 1 {
			t.Errorf("update_strategy should have 1 option, got %d", len(strategyVar.Options))
		}
		if strategyVar.Default != "Setters" {
			t.Errorf("update_strategy default should be Setters, got %v", strategyVar.Default)
		}
	} else {
		t.Errorf("update_strategy variable not found")
	}

	// Test automation_interval variable
	if intervalVar, exists := variableMap["automation_interval"]; exists {
		if intervalVar.Type != VariableTypeSelect {
			t.Errorf("automation_interval variable should be select type")
		}
		if !intervalVar.Required {
			t.Errorf("automation_interval variable should be required")
		}
		if intervalVar.Default != "1m" {
			t.Errorf("automation_interval default should be 1m, got %v", intervalVar.Default)
		}
		if len(intervalVar.Options) == 0 {
			t.Errorf("automation_interval should have options")
		}
	} else {
		t.Errorf("automation_interval variable not found")
	}

	// Test JSON array variables
	for _, varName := range []string{"image_repositories", "image_policies"} {
		if jsonVar, exists := variableMap[varName]; exists {
			if jsonVar.Type != VariableTypeText {
				t.Errorf("%s variable should be text type", varName)
			}
			if !jsonVar.Required {
				t.Errorf("%s variable should be required", varName)
			}
		} else {
			t.Errorf("%s variable not found", varName)
		}
	}
}

func TestImageUpdatePlugin_Validate(t *testing.T) {
	plugin := NewImageUpdatePlugin()

	// Create test data
	imageRepositories := []ImageRepository{
		{
			Name:     "zigbee2mqtt",
			Image:    "koenkk/zigbee2mqtt",
			Interval: "60m",
		},
		{
			Name:      "app-daemon",
			Image:     "harbor.example.com/apps/app-daemon",
			Interval:  "1m",
			SecretRef: "harbor-docker-creds",
		},
	}

	imagePolicies := []ImagePolicy{
		{
			Name:       "zigbee2mqtt",
			Repository: "zigbee2mqtt",
			PolicyType: "semver",
			Range:      "*",
		},
		{
			Name:       "app-daemon",
			Repository: "app-daemon",
			PolicyType: "numerical",
			Pattern:    "^main-[a-f0-9]+-(?P<ts>[0-9]+)",
			Extract:    "$ts",
			Order:      "asc",
		},
	}

	reposJSON, _ := json.Marshal(imageRepositories)
	policiesJSON, _ := json.Marshal(imagePolicies)

	tests := []struct {
		name        string
		values      map[string]interface{}
		expectError bool
		errorText   string
	}{
		{
			name: "valid configuration",
			values: map[string]interface{}{
				"automation_name":           "home-automation",
				"git_repository_name":       "flux-system",
				"git_repository_namespace":  "flux-system",
				"update_path":               "./apps/home-automation",
				"update_strategy":           "Setters",
				"git_branch":                "main",
				"author_name":               "Homelab Flux",
				"author_email":              "flux@example.com",
				"commit_message_template":   "chore: update container versions",
				"automation_interval":       "1m",
				"image_repositories":        string(reposJSON),
				"image_policies":            string(policiesJSON),
			},
			expectError: false,
		},
		{
			name: "missing required automation_name",
			values: map[string]interface{}{
				"git_repository_name":      "flux-system",
				"git_repository_namespace": "flux-system",
				"update_path":              "./apps/home-automation",
				"update_strategy":          "Setters",
				"git_branch":               "main",
				"author_name":              "Homelab Flux",
				"author_email":             "flux@example.com",
				"commit_message_template":  "chore: update container versions",
				"automation_interval":      "1m",
				"image_repositories":       string(reposJSON),
				"image_policies":           string(policiesJSON),
			},
			expectError: true,
			errorText:   "automation_name",
		},
		{
			name: "invalid image_repositories JSON",
			values: map[string]interface{}{
				"automation_name":           "home-automation",
				"git_repository_name":       "flux-system",
				"git_repository_namespace":  "flux-system",
				"update_path":               "./apps/home-automation",
				"update_strategy":           "Setters",
				"git_branch":                "main",
				"author_name":               "Homelab Flux",
				"author_email":              "flux@example.com",
				"commit_message_template":   "chore: update container versions",
				"automation_interval":       "1m",
				"image_repositories":        "invalid json",
				"image_policies":            string(policiesJSON),
			},
			expectError: true,
			errorText:   "invalid JSON format",
		},
		{
			name: "missing repository name",
			values: map[string]interface{}{
				"automation_name":           "home-automation",
				"git_repository_name":       "flux-system",
				"git_repository_namespace":  "flux-system",
				"update_path":               "./apps/home-automation",
				"update_strategy":           "Setters",
				"git_branch":                "main",
				"author_name":               "Homelab Flux",
				"author_email":              "flux@example.com",
				"commit_message_template":   "chore: update container versions",
				"automation_interval":       "1m",
				"image_repositories":        `[{"image": "test", "interval": "1m"}]`,
				"image_policies":            string(policiesJSON),
			},
			expectError: true,
			errorText:   "name is required",
		},
		{
			name: "missing policy range for semver",
			values: map[string]interface{}{
				"automation_name":           "home-automation",
				"git_repository_name":       "flux-system",
				"git_repository_namespace":  "flux-system",
				"update_path":               "./apps/home-automation",
				"update_strategy":           "Setters",
				"git_branch":                "main",
				"author_name":               "Homelab Flux",
				"author_email":              "flux@example.com",
				"commit_message_template":   "chore: update container versions",
				"automation_interval":       "1m",
				"image_repositories":        string(reposJSON),
				"image_policies":            `[{"name": "test", "repository": "test", "policyType": "semver"}]`,
			},
			expectError: true,
			errorText:   "range is required for semver policy",
		},
		{
			name: "missing pattern for numerical policy",
			values: map[string]interface{}{
				"automation_name":           "home-automation",
				"git_repository_name":       "flux-system",
				"git_repository_namespace":  "flux-system",
				"update_path":               "./apps/home-automation",
				"update_strategy":           "Setters",
				"git_branch":                "main",
				"author_name":               "Homelab Flux",
				"author_email":              "flux@example.com",
				"commit_message_template":   "chore: update container versions",
				"automation_interval":       "1m",
				"image_repositories":        string(reposJSON),
				"image_policies":            `[{"name": "test", "repository": "test", "policyType": "numerical"}]`,
			},
			expectError: true,
			errorText:   "pattern is required for numerical policy",
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

func TestImageUpdatePlugin_GenerateFile(t *testing.T) {
	plugin := NewImageUpdatePlugin()

	tempDir := t.TempDir()
	appDir := filepath.Join(tempDir, "test-app")

	// Create test data
	imageRepositories := []ImageRepository{
		{
			Name:     "zigbee2mqtt",
			Image:    "koenkk/zigbee2mqtt",
			Interval: "60m",
		},
		{
			Name:      "app-daemon",
			Image:     "harbor.example.com/apps/app-daemon",
			Interval:  "1m",
			SecretRef: "harbor-docker-creds",
		},
	}

	imagePolicies := []ImagePolicy{
		{
			Name:       "zigbee2mqtt",
			Repository: "zigbee2mqtt",
			PolicyType: "semver",
			Range:      "*",
		},
		{
			Name:       "app-daemon",
			Repository: "app-daemon",
			PolicyType: "numerical",
			Pattern:    "^main-[a-f0-9]+-(?P<ts>[0-9]+)",
			Extract:    "$ts",
			Order:      "asc",
		},
	}

	reposJSON, _ := json.Marshal(imageRepositories)
	policiesJSON, _ := json.Marshal(imagePolicies)

	values := map[string]interface{}{
		"automation_name":           "home-automation",
		"git_repository_name":       "flux-system",
		"git_repository_namespace":  "flux-system",
		"update_path":               "./apps/home-automation",
		"update_strategy":           "Setters",
		"git_branch":                "main",
		"author_name":               "Homelab Flux",
		"author_email":              "flux@example.com",
		"commit_message_template":   "chore: update container versions",
		"automation_interval":       "1m",
		"image_repositories":        string(reposJSON),
		"image_policies":            string(policiesJSON),
	}

	err := plugin.GenerateFile(values, appDir, "home-automation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if update directory was created
	updateDir := filepath.Join(appDir, "update")
	if _, err := os.Stat(updateDir); os.IsNotExist(err) {
		t.Errorf("expected update directory to be created at %s", updateDir)
	}

	// Check if all three files were created
	expectedFiles := []string{
		"image-repository.yaml",
		"image-policy.yaml",
		"image-update-automation.yaml",
	}

	for _, filename := range expectedFiles {
		expectedPath := filepath.Join(updateDir, filename)
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to be created", filename)
		}

		// Read and check file content
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("failed to read generated file %s: %v", filename, err)
		}

		contentStr := string(content)

		// Ensure file ends with newline
		if !strings.HasSuffix(contentStr, "\n") {
			t.Errorf("generated file %s should end with newline", filename)
		}

		// Check specific content based on file type
		switch filename {
		case "image-repository.yaml":
			expectedContent := []string{
				"apiVersion: image.toolkit.fluxcd.io/v1beta2",
				"kind: ImageRepository",
				"name: zigbee2mqtt",
				"image: koenkk/zigbee2mqtt",
				"interval: 60m",
				"name: app-daemon",
				"image: harbor.example.com/apps/app-daemon",
				"interval: 1m",
				"secretRef:",
				"name: harbor-docker-creds",
			}
			for _, expected := range expectedContent {
				if !strings.Contains(contentStr, expected) {
					t.Errorf("image-repository.yaml should contain '%s'\nGenerated content:\n%s", expected, contentStr)
				}
			}

		case "image-policy.yaml":
			expectedContent := []string{
				"apiVersion: image.toolkit.fluxcd.io/v1beta2",
				"kind: ImagePolicy",
				"name: zigbee2mqtt",
				"imageRepositoryRef:",
				"name: zigbee2mqtt",
				"policy:",
				"semver:",
				"range: '*'",
				"name: app-daemon",
				"imageRepositoryRef:",
				"name: app-daemon",
				"filterTags:",
				"pattern: '^main-[a-f0-9]+-(?P<ts>[0-9]+)'",
				"extract: '$ts'",
				"numerical:",
				"order: asc",
			}
			for _, expected := range expectedContent {
				if !strings.Contains(contentStr, expected) {
					t.Errorf("image-policy.yaml should contain '%s'\nGenerated content:\n%s", expected, contentStr)
				}
			}

		case "image-update-automation.yaml":
			expectedContent := []string{
				"apiVersion: image.toolkit.fluxcd.io/v1beta1",
				"kind: ImageUpdateAutomation",
				"name: home-automation",
				"namespace: home-automation",
				"interval: 1m",
				"sourceRef:",
				"kind: GitRepository",
				"name: flux-system",
				"namespace: flux-system",
				"git:",
				"commit:",
				"author:",
				"email: flux@example.com",
				"name: Homelab Flux",
				"messageTemplate: \"chore: update container versions\"",
				"push:",
				"branch: main",
				"update:",
				"path: ./apps/home-automation",
				"strategy: Setters",
			}
			for _, expected := range expectedContent {
				if !strings.Contains(contentStr, expected) {
					t.Errorf("image-update-automation.yaml should contain '%s'\nGenerated content:\n%s", expected, contentStr)
				}
			}
		}
	}
}

func TestImageUpdatePlugin_GenerateFile_EmptyArrays(t *testing.T) {
	plugin := NewImageUpdatePlugin()

	tempDir := t.TempDir()
	appDir := filepath.Join(tempDir, "test-app")

	values := map[string]interface{}{
		"automation_name":           "empty-test",
		"git_repository_name":       "flux-system",
		"git_repository_namespace":  "flux-system",
		"update_path":               "./apps/empty-test",
		"update_strategy":           "Setters",
		"git_branch":                "main",
		"author_name":               "Test User",
		"author_email":              "test@example.com",
		"commit_message_template":   "chore: update versions",
		"automation_interval":       "5m",
		"image_repositories":        "[]",
		"image_policies":            "[]",
	}

	err := plugin.GenerateFile(values, appDir, "test-namespace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that files are created even with empty arrays
	updateDir := filepath.Join(appDir, "update")
	expectedFiles := []string{
		"image-repository.yaml",
		"image-policy.yaml",
		"image-update-automation.yaml",
	}

	for _, filename := range expectedFiles {
		expectedPath := filepath.Join(updateDir, filename)
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to be created even with empty arrays", filename)
		}
	}

	// Check image-update-automation.yaml content
	automationPath := filepath.Join(updateDir, "image-update-automation.yaml")
	content, err := os.ReadFile(automationPath)
	if err != nil {
		t.Fatalf("failed to read automation file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "name: empty-test") {
		t.Errorf("automation file should contain automation name")
	}
	if !strings.Contains(contentStr, "namespace: test-namespace") {
		t.Errorf("automation file should contain correct namespace")
	}
}