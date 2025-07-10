package helm

import (
	"fmt"
	"io"
	"net/http"
	"sort"

	"gopkg.in/yaml.v3"
)

type IndexYAML struct {
	Entries map[string][]struct {
		Version     string `yaml:"version"`
		AppVersion  string `yaml:"appVersion"`
		Description string `yaml:"description"`
	} `yaml:"entries"`
}

// ChartVersion represents a chart version with metadata
type ChartVersion struct {
	ChartVersion   string
	AppVersion     string
	Description    string
	DisplayString  string
}

// VersionFetcher handles fetching chart versions from Helm repositories
type VersionFetcher struct{}

// NewVersionFetcher creates a new version fetcher
func NewVersionFetcher() *VersionFetcher {
	return &VersionFetcher{}
}

func fetchIndexYAML(repoURL string) (*IndexYAML, error) {
	indexURL := repoURL
	if indexURL[len(indexURL)-1] != '/' {
		indexURL += "/"
	}
	indexURL += "index.yaml"
	resp, err := http.Get(indexURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch index.yaml: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch index.yaml: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read index.yaml: %w", err)
	}
	var idx IndexYAML
	if err := yaml.Unmarshal(body, &idx); err != nil {
		return nil, fmt.Errorf("failed to parse index.yaml: %w", err)
	}
	return &idx, nil
}

// ListCharts fetches all chart names and their descriptions from a Helm repository
func (vf *VersionFetcher) ListCharts(repoURL string) ([]struct{ Name, Description string }, error) {
	idx, err := fetchIndexYAML(repoURL)
	if err != nil {
		return nil, err
	}
	var charts []struct{ Name, Description string }
	for name, entries := range idx.Entries {
		desc := ""
		if len(entries) > 0 {
			desc = entries[0].Description
		}
		charts = append(charts, struct{ Name, Description string }{ Name: name, Description: desc })
	}
	// Sort charts by name
	sort.Slice(charts, func(i, j int) bool { return charts[i].Name < charts[j].Name })
	return charts, nil
}

// FetchChartVersions fetches available versions for a chart from a repository
func (vf *VersionFetcher) FetchChartVersions(repoURL, chartName string) ([]ChartVersion, error) {
	idx, err := fetchIndexYAML(repoURL)
	if err != nil {
		return nil, err
	}
	chart, exists := idx.Entries[chartName]
	if !exists {
		return nil, fmt.Errorf("chart '%s' not found in repository", chartName)
	}
	var versions []ChartVersion
	for _, version := range chart {
		if version.Version != "" {
			displayString := fmt.Sprintf("%s\t%s\t%s\t%s", chartName, version.Version, version.AppVersion, version.Description)
			versions = append(versions, ChartVersion{
				ChartVersion:  version.Version,
				AppVersion:    version.AppVersion,
				Description:   version.Description,
				DisplayString: displayString,
			})
		}
	}
	// Sort versions (newest first, lexicographically)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].ChartVersion > versions[j].ChartVersion
	})
	return versions, nil
}

// FetchLatestVersion fetches the latest version for a chart
func (vf *VersionFetcher) FetchLatestVersion(repoURL, chartName string) (ChartVersion, error) {
	versions, err := vf.FetchChartVersions(repoURL, chartName)
	if err != nil {
		return ChartVersion{}, err
	}
	if len(versions) == 0 {
		return ChartVersion{}, fmt.Errorf("no versions found for chart '%s'", chartName)
	}
	return versions[0], nil
}

// ValidateChartExists checks if a chart exists in the repository
func (vf *VersionFetcher) ValidateChartExists(repoURL, chartName string) error {
	idx, err := fetchIndexYAML(repoURL)
	if err != nil {
		return err
	}
	if _, exists := idx.Entries[chartName]; !exists {
		return fmt.Errorf("chart '%s' not found in repository", chartName)
	}
	return nil
} 