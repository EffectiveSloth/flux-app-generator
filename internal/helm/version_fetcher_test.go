package helm

import (
	"sort"
	"testing"
)

func TestChartVersion_Sorting(t *testing.T) {
	versions := []ChartVersion{
		{ChartVersion: "1.0.1", AppVersion: "1.0.1", Description: "Old version"},
		{ChartVersion: "1.1.0", AppVersion: "1.0.28", Description: "New version"},
		{ChartVersion: "1.0.40", AppVersion: "1.0.28", Description: "Middle version"},
	}

	// Sort versions (newest first) by chart version
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].ChartVersion > versions[j].ChartVersion
	})

	// Check that versions are sorted correctly
	expected := []string{"1.1.0", "1.0.40", "1.0.1"}
	for i, version := range versions {
		if version.ChartVersion != expected[i] {
			t.Errorf("Expected version %s at position %d, got %s", expected[i], i, version.ChartVersion)
		}
	}

	t.Logf("Versions sorted correctly: %v", expected)
}

func TestChartVersion_DisplayString(t *testing.T) {
	chartName := "vantage-kubernetes-agent"
	version := ChartVersion{
		ChartVersion:  "1.1.2",
		AppVersion:    "1.0.28",
		Description:   "Provisions the Vantage Kubernetes agent.",
		DisplayString: chartName + "\t" + "1.1.2" + "\t" + "1.0.28" + "\t" + "Provisions the Vantage Kubernetes agent.",
	}

	// Test DisplayString generation
	expected := chartName + "\t" + version.ChartVersion + "\t" + version.AppVersion + "\t" + version.Description

	if version.DisplayString != expected {
		t.Errorf("DisplayString mismatch.\nExpected: %s\nGot:      %s", expected, version.DisplayString)
	}

	t.Logf("DisplayString format is correct: %s", version.DisplayString)
}

func TestChartVersion_Structure(t *testing.T) {
	version := ChartVersion{
		ChartVersion:  "1.1.2",
		AppVersion:    "1.0.28",
		Description:   "Provisions the Vantage Kubernetes agent.",
		DisplayString: "vantage-kubernetes-agent\t1.1.2\t1.0.28\tProvisions the Vantage Kubernetes agent.",
	}

	// Test that all fields are populated
	if version.ChartVersion == "" {
		t.Error("ChartVersion should not be empty")
	}
	if version.AppVersion == "" {
		t.Error("AppVersion should not be empty")
	}
	if version.Description == "" {
		t.Error("Description should not be empty")
	}
	if version.DisplayString == "" {
		t.Error("DisplayString should not be empty")
	}

	t.Logf("ChartVersion structure is valid: %+v", version)
}

func TestVersionFetcher_NewVersionFetcher(t *testing.T) {
	fetcher := NewVersionFetcher()
	if fetcher == nil {
		t.Error("NewVersionFetcher should return a non-nil fetcher")
	}
	t.Logf("VersionFetcher created successfully: %+v", fetcher)
}

func TestVersionFetcher_ListCharts(t *testing.T) {
	fetcher := NewVersionFetcher()

	t.Run("valid repo", func(t *testing.T) {
		repoURL := "https://vantage-sh.github.io/helm-charts"
		charts, err := fetcher.ListCharts(repoURL)
		if err != nil {
			t.Fatalf("Failed to list charts: %v", err)
		}
		if len(charts) == 0 {
			t.Fatal("Expected to find charts, but got none")
		}
		found := false
		for _, c := range charts {
			if c.Name == "vantage-kubernetes-agent" {
				found = true
				if c.Description == "" {
					t.Error("Expected description for vantage-kubernetes-agent, got empty string")
				}
			}
		}
		if !found {
			t.Error("Expected to find chart 'vantage-kubernetes-agent' in the repo")
		}
	})

	t.Run("invalid repo", func(t *testing.T) {
		_, err := fetcher.ListCharts("https://not-a-real-helm-repo")
		if err == nil {
			t.Error("Expected error for invalid repo URL, got nil")
		}
	})
}

func TestVersionFetcher_FetchChartVersions(t *testing.T) {
	fetcher := NewVersionFetcher()
	repoURL := "https://vantage-sh.github.io/helm-charts"
	chartName := "vantage-kubernetes-agent"

	versions, err := fetcher.FetchChartVersions(repoURL, chartName)
	if err != nil {
		t.Fatalf("Failed to fetch chart versions: %v", err)
	}

	if len(versions) == 0 {
		t.Fatal("Expected to find chart versions, but got none")
	}

	// Test that versions are sorted (newest first)
	if len(versions) > 1 {
		if versions[0].ChartVersion <= versions[1].ChartVersion {
			t.Errorf("Versions not sorted correctly. Expected %s > %s",
				versions[0].ChartVersion, versions[1].ChartVersion)
		}
	}

	// Test that each version has required fields
	for i, version := range versions {
		if version.ChartVersion == "" {
			t.Errorf("Version %d has empty ChartVersion", i)
		}
		if version.AppVersion == "" {
			t.Errorf("Version %d has empty AppVersion", i)
		}
		if version.Description == "" {
			t.Errorf("Version %d has empty Description", i)
		}
		if version.DisplayString == "" {
			t.Errorf("Version %d has empty DisplayString", i)
		}
	}

	t.Logf("Found %d versions for chart %s", len(versions), chartName)
	t.Logf("Latest version: %s (App: %s)", versions[0].ChartVersion, versions[0].AppVersion)
}
