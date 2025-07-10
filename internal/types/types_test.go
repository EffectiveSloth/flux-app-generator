package types

import (
	"testing"
)

func TestAppConfig_Structure(t *testing.T) {
	config := AppConfig{
		AppName:      "test-app",
		Namespace:    "default",
		HelmRepoName: "test-repo",
		HelmRepoURL:  "https://example.com/repo",
		ChartName:    "test-chart",
		ChartVersion: "1.0.0",
		Interval:     "5m",
		Values: map[string]interface{}{
			"replicaCount": 3,
			"image": map[string]string{
				"repository": "nginx",
				"tag":        "latest",
			},
		},
	}

	// Test that all fields are properly set
	if config.AppName != "test-app" {
		t.Errorf("Expected AppName to be 'test-app', got '%s'", config.AppName)
	}

	if config.Namespace != "default" {
		t.Errorf("Expected Namespace to be 'default', got '%s'", config.Namespace)
	}

	if config.HelmRepoName != "test-repo" {
		t.Errorf("Expected HelmRepoName to be 'test-repo', got '%s'", config.HelmRepoName)
	}

	if config.HelmRepoURL != "https://example.com/repo" {
		t.Errorf("Expected HelmRepoURL to be 'https://example.com/repo', got '%s'", config.HelmRepoURL)
	}

	if config.ChartName != "test-chart" {
		t.Errorf("Expected ChartName to be 'test-chart', got '%s'", config.ChartName)
	}

	if config.ChartVersion != "1.0.0" {
		t.Errorf("Expected ChartVersion to be '1.0.0', got '%s'", config.ChartVersion)
	}

	if config.Interval != "5m" {
		t.Errorf("Expected Interval to be '5m', got '%s'", config.Interval)
	}

	if config.Values == nil {
		t.Error("Expected Values to be non-nil")
	}

	if replicaCount, ok := config.Values["replicaCount"].(int); !ok || replicaCount != 3 {
		t.Errorf("Expected replicaCount to be 3, got %v", config.Values["replicaCount"])
	}

	if image, ok := config.Values["image"].(map[string]string); !ok {
		t.Errorf("Expected image to be map[string]string, got %T", config.Values["image"])
	} else {
		if image["repository"] != "nginx" {
			t.Errorf("Expected image.repository to be 'nginx', got '%s'", image["repository"])
		}
		if image["tag"] != "latest" {
			t.Errorf("Expected image.tag to be 'latest', got '%s'", image["tag"])
		}
	}
}

func TestAppConfig_EmptyValues(t *testing.T) {
	config := AppConfig{
		AppName:   "test-app",
		Namespace: "default",
	}

	// Test that Values can be nil
	if config.Values != nil {
		t.Error("Expected Values to be nil for empty config")
	}
}

func TestAppConfig_ZeroValues(t *testing.T) {
	config := AppConfig{}

	// Test default zero values
	if config.AppName != "" {
		t.Errorf("Expected AppName to be empty string, got '%s'", config.AppName)
	}

	if config.Namespace != "" {
		t.Errorf("Expected Namespace to be empty string, got '%s'", config.Namespace)
	}

	if config.HelmRepoName != "" {
		t.Errorf("Expected HelmRepoName to be empty string, got '%s'", config.HelmRepoName)
	}

	if config.HelmRepoURL != "" {
		t.Errorf("Expected HelmRepoURL to be empty string, got '%s'", config.HelmRepoURL)
	}

	if config.ChartName != "" {
		t.Errorf("Expected ChartName to be empty string, got '%s'", config.ChartName)
	}

	if config.ChartVersion != "" {
		t.Errorf("Expected ChartVersion to be empty string, got '%s'", config.ChartVersion)
	}

	if config.Interval != "" {
		t.Errorf("Expected Interval to be empty string, got '%s'", config.Interval)
	}

	if config.Values != nil {
		t.Error("Expected Values to be nil for zero config")
	}
}
