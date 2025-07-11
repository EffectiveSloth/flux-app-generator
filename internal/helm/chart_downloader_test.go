package helm

import (
	"strings"
	"testing"
)

func TestDownloadAndExtractValuesYAML(t *testing.T) {
	tests := []struct {
		name         string
		repoURL      string
		chartName    string
		chartVersion string
		expectError  bool
		expectEmpty  bool
	}{
		{
			name:         "valid chart and version",
			repoURL:      "https://vantage-sh.github.io/helm-charts",
			chartName:    "vantage-kubernetes-agent",
			chartVersion: "1.1.2",
			expectError:  false,
			expectEmpty:  false,
		},
		{
			name:         "invalid chart name",
			repoURL:      "https://vantage-sh.github.io/helm-charts",
			chartName:    "non-existent-chart",
			chartVersion: "1.0.0",
			expectError:  true,
			expectEmpty:  true,
		},
		{
			name:         "invalid version",
			repoURL:      "https://vantage-sh.github.io/helm-charts",
			chartName:    "vantage-kubernetes-agent",
			chartVersion: "999.999.999",
			expectError:  true,
			expectEmpty:  true,
		},
		{
			name:         "invalid repo URL",
			repoURL:      "https://invalid-repo.example.com",
			chartName:    "some-chart",
			chartVersion: "1.0.0",
			expectError:  true,
			expectEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := DownloadAndExtractValuesYAML(tt.repoURL, tt.chartName, tt.chartVersion)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectEmpty && values != "" {
				t.Errorf("Expected empty values but got: %s", values)
			}
			if !tt.expectEmpty && values == "" {
				t.Errorf("Expected non-empty values but got empty string")
			}

			// If we got values successfully, check that it looks like a valid YAML
			if !tt.expectError && !tt.expectEmpty {
				if !strings.Contains(values, "name:") && !strings.Contains(values, "image:") && !strings.Contains(values, "#") {
					previewLen := 100
					if len(values) < previewLen {
						previewLen = len(values)
					}
					t.Errorf("Values don't look like a valid values.yaml file: %s", values[:previewLen])
				}
				t.Logf("Successfully extracted values.yaml (%d bytes)", len(values))
			}
		})
	}
}

func TestDownloadAndExtractValuesYAML_ContentValidation(t *testing.T) {
	// Test with a known working chart to validate content
	values, err := DownloadAndExtractValuesYAML(
		"https://vantage-sh.github.io/helm-charts",
		"vantage-kubernetes-agent",
		"1.1.2",
	)
	if err != nil {
		t.Skipf("Skipping content validation test due to network error: %v", err)
		return
	}

	// Check that the values.yaml contains expected content structure
	expectedFields := []string{
		"image:",
		"resources:",
		"nodeSelector:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(values, field) {
			t.Errorf("Expected field '%s' not found in values.yaml", field)
		}
	}

	// Check that it's valid YAML structure (basic check)
	if !strings.HasPrefix(strings.TrimSpace(values), "#") && !strings.Contains(values, ":") {
		t.Errorf("Values don't appear to be valid YAML format")
	}

	previewLen := 200
	if len(values) < previewLen {
		previewLen = len(values)
	}
	t.Logf("Values content preview: %s...", values[:previewLen])
}
