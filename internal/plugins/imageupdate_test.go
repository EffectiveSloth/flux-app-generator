package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewImageUpdatePlugin(t *testing.T) {
	plugin := NewImageUpdatePlugin()

	if plugin.Name() != "imageupdate" {
		t.Errorf("expected name 'imageupdate', got '%s'", plugin.Name())
	}

	if plugin.Description() == "" {
		t.Error("description should not be empty")
	}

	variables := plugin.Variables()
	if len(variables) != 1 {
		t.Errorf("plugin should have 1 variable, got %d", len(variables))
	}

	// Check for the only required variable
	if variables[0].Name != "automation_name" {
		t.Errorf("expected variable 'automation_name', got '%s'", variables[0].Name)
	}

	if !variables[0].Required {
		t.Error("automation_name should be required")
	}

	// Test file path template
	expectedFilePath := "image-update-automation.yaml"
	if plugin.FilePath() != expectedFilePath {
		t.Errorf("expected file path '%s', got '%s'", expectedFilePath, plugin.FilePath())
	}
}

func TestImageUpdatePlugin_Variables(t *testing.T) {
	plugin := NewImageUpdatePlugin()
	variables := plugin.Variables()

	if len(variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(variables))
	}

	// Test automation_name variable
	variable := variables[0]
	if variable.Name != "automation_name" {
		t.Errorf("expected variable name 'automation_name', got '%s'", variable.Name)
	}

	if variable.Type != VariableTypeText {
		t.Errorf("automation_name variable should be text type")
	}

	if !variable.Required {
		t.Errorf("automation_name variable should be required")
	}

	if variable.Description == "" {
		t.Errorf("automation_name should have a description")
	}
}

// Consolidated validation tests for all scenarios
func TestImageUpdatePlugin_Validate(t *testing.T) {
	plugin := NewImageUpdatePlugin()

	tests := []struct {
		name        string
		values      map[string]interface{}
		expectError bool
		errorText   string
	}{
		{
			name: "valid configuration with mock data",
			values: map[string]interface{}{
				"automation_name":          "home-automation",
				"image_repositories":       `[{"name":"myapp","image":"myregistry/myapp","interval":"6h"}]`,
				"image_policies":           `[{"name":"myapp","repository":"myapp","policyType":"semver","range":"*"}]`,
				"git_repository_name":      DefaultFluxNamespace,
				"git_repository_namespace": DefaultFluxNamespace,
				"update_path":              "./apps/test",
				"git_branch":               "main",
				"author_name":              "Test Author",
				"author_email":             "test@example.com",
				"automation_interval":      "10m",
				"update_strategy":          "Setters",
				"commit_message_template":  "chore: update container versions",
			},
			expectError: false,
		},
		{
			name: "missing required automation_name",
			values: map[string]interface{}{
				"git_repository_name": DefaultFluxNamespace,
			},
			expectError: true,
			errorText:   "automation_name",
		},
		// Repository validation tests
		{
			name: "invalid repositories JSON format",
			values: map[string]interface{}{
				"automation_name":    "test",
				"image_repositories": `[{"name":"app","image":"myregistry/app",invalid}]`,
			},
			expectError: true,
			errorText:   "invalid JSON format",
		},
		{
			name: "missing repository name",
			values: map[string]interface{}{
				"automation_name":    "test",
				"image_repositories": `[{"image":"myregistry/app","interval":"6h"}]`,
			},
			expectError: true,
			errorText:   "name is required",
		},
		{
			name: "missing repository image",
			values: map[string]interface{}{
				"automation_name":    "test",
				"image_repositories": `[{"name":"app","interval":"6h"}]`,
			},
			expectError: true,
			errorText:   "image is required",
		},
		{
			name: "missing repository interval",
			values: map[string]interface{}{
				"automation_name":    "test",
				"image_repositories": `[{"name":"app","image":"myregistry/app"}]`,
			},
			expectError: true,
			errorText:   "interval is required",
		},
		// Policy validation tests
		{
			name: "valid semver policy",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"name":"app","repository":"app","policyType":"semver","range":"*"}]`,
			},
			expectError: false,
		},
		{
			name: "valid numerical policy",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"name":"app","repository":"app","policyType":"numerical","pattern":"^main-[a-f0-9]+","extract":"$ts","order":"asc"}]`,
			},
			expectError: false,
		},
		{
			name: "invalid policies JSON",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"name":"app",invalid}]`,
			},
			expectError: true,
			errorText:   "invalid JSON format",
		},
		{
			name: "missing policy name",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"repository":"app","policyType":"semver","range":"*"}]`,
			},
			expectError: true,
			errorText:   "name is required",
		},
		{
			name: "missing policy repository",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"name":"app","policyType":"semver","range":"*"}]`,
			},
			expectError: true,
			errorText:   "repository is required",
		},
		{
			name: "missing policy type",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"name":"app","repository":"app","range":"*"}]`,
			},
			expectError: true,
			errorText:   "policyType is required",
		},
		{
			name: "semver policy missing range",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"name":"app","repository":"app","policyType":"semver"}]`,
			},
			expectError: true,
			errorText:   "range is required for semver policy",
		},
		{
			name: "numerical policy missing pattern",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"name":"app","repository":"app","policyType":"numerical","extract":"$ts","order":"asc"}]`,
			},
			expectError: true,
			errorText:   "pattern is required for numerical policy",
		},
		{
			name: "numerical policy missing extract",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"name":"app","repository":"app","policyType":"numerical","pattern":"^main-[a-f0-9]+","order":"asc"}]`,
			},
			expectError: true,
			errorText:   "extract is required for numerical policy",
		},
		{
			name: "numerical policy missing order",
			values: map[string]interface{}{
				"automation_name": "test",
				"image_policies":  `[{"name":"app","repository":"app","policyType":"numerical","pattern":"^main-[a-f0-9]+","extract":"$ts"}]`,
			},
			expectError: true,
			errorText:   "order is required for numerical policy",
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

// Consolidated GenerateFile tests
func TestImageUpdatePlugin_GenerateFile(t *testing.T) {
	plugin := NewImageUpdatePlugin()

	tests := []struct {
		name          string
		values        map[string]interface{}
		expectError   bool
		expectedFiles []string
		checkContent  map[string]string // file -> content to check
	}{
		{
			name: "successful generation with valid data",
			values: map[string]interface{}{
				"automation_name":          "home-automation",
				"image_repositories":       `[{"name":"myapp","image":"myregistry/myapp","interval":"6h"}]`,
				"image_policies":           `[{"name":"myapp","repository":"myapp","policyType":"semver","range":"*"}]`,
				"git_repository_name":      DefaultFluxNamespace,
				"git_repository_namespace": DefaultFluxNamespace,
				"update_path":              "./apps/test",
				"git_branch":               "main",
				"author_name":              "Test Author",
				"author_email":             "test@example.com",
				"automation_interval":      "10m",
				"update_strategy":          "Setters",
				"commit_message_template":  "chore: update container versions",
			},
			expectError: false,
			expectedFiles: []string{
				"image-repository.yaml",
				"image-policy.yaml",
				"image-update-automation.yaml",
			},
			checkContent: map[string]string{
				"image-repository.yaml": "name: myapp",
				"image-policy.yaml":     "semver:",
			},
		},
		{
			name: "generation with empty arrays",
			values: map[string]interface{}{
				"automation_name":          "test-automation",
				"image_repositories":       `[]`,
				"image_policies":           `[]`,
				"git_repository_name":      DefaultFluxNamespace,
				"git_repository_namespace": DefaultFluxNamespace,
				"update_path":              "./apps/test",
				"git_branch":               "main",
				"author_name":              "Test Author",
				"author_email":             "test@example.com",
				"automation_interval":      "10m",
			},
			expectError: false,
			expectedFiles: []string{
				"image-repository.yaml",
				"image-policy.yaml",
				"image-update-automation.yaml",
			},
		},
		{
			name: "invalid repositories JSON",
			values: map[string]interface{}{
				"automation_name":    "test",
				"image_repositories": `invalid json`,
				"image_policies":     `[]`,
			},
			expectError: true,
		},
		{
			name: "invalid policies JSON",
			values: map[string]interface{}{
				"automation_name":    "test",
				"image_repositories": `[]`,
				"image_policies":     `invalid json`,
			},
			expectError: true,
		},
		{
			name: "multiple repositories and policies",
			values: map[string]interface{}{
				"automation_name": "multi-app",
				"image_repositories": `[
					{"name":"app1","image":"registry/app1","interval":"5m"},
					{"name":"app2","image":"registry/app2","interval":"10m"}
				]`,
				"image_policies": `[
					{"name":"app1","repository":"app1","policyType":"semver","range":">=1.0.0"},
					{"name":"app2","repository":"app2","policyType":"numerical","pattern":"^v[0-9]+","extract":"$ts","order":"desc"}
				]`,
				"git_repository_name":      DefaultFluxNamespace,
				"git_repository_namespace": DefaultFluxNamespace,
				"update_path":              "./apps",
				"git_branch":               "main",
				"author_name":              "Test Author",
				"author_email":             "test@example.com",
				"automation_interval":      "15m",
			},
			expectError: false,
			expectedFiles: []string{
				"image-repository.yaml",
				"image-policy.yaml",
				"image-update-automation.yaml",
			},
			checkContent: map[string]string{
				"image-repository.yaml": "app1",
				"image-policy.yaml":     "app2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			namespace := DefaultFluxNamespace

			err := plugin.GenerateFile(tt.values, tempDir, namespace)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("GenerateFile failed: %v", err)
			}

			// Check if expected files were created
			for _, fileName := range tt.expectedFiles {
				filePath := filepath.Join(tempDir, fileName)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("expected file %s was not created", fileName)
				}
			}

			// Check content if specified
			for fileName, expectedContent := range tt.checkContent {
				filePath := filepath.Join(tempDir, fileName)
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("failed to read %s: %v", fileName, err)
					continue
				}

				if !strings.Contains(string(content), expectedContent) {
					t.Errorf("%s should contain '%s'", fileName, expectedContent)
				}
			}
		})
	}
}

// Test constants
func TestImageUpdatePlugin_Constants(t *testing.T) {
	if PolicyTypeSemver != "semver" {
		t.Errorf("PolicyTypeSemver should be 'semver', got '%s'", PolicyTypeSemver)
	}
	if PolicyTypeTimestamp != "timestamp" {
		t.Errorf("PolicyTypeTimestamp should be 'timestamp', got '%s'", PolicyTypeTimestamp)
	}
	if PolicyTypeNumerical != "numerical" {
		t.Errorf("PolicyTypeNumerical should be 'numerical', got '%s'", PolicyTypeNumerical)
	}
	if DefaultFluxNamespace != "flux-system" {
		t.Errorf("DefaultFluxNamespace should be 'flux-system', got '%s'", DefaultFluxNamespace)
	}
}
