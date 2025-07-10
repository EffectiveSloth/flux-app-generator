package helm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DownloadAndExtractValuesYAML downloads the chart tarball and extracts values.yaml as a string.
func DownloadAndExtractValuesYAML(repoURL, chartName, chartVersion string) (string, error) {
	idx, err := fetchIndexYAML(repoURL)
	if err != nil {
		return "", err
	}
	chartEntries, ok := idx.Entries[chartName]
	if !ok {
		return "", fmt.Errorf("chart '%s' not found in repository", chartName)
	}
	var chartURL string
	for _, entry := range chartEntries {
		if entry.Version == chartVersion {
			if len(entry.URLs) == 0 {
				return "", fmt.Errorf("no tarball URL found for chart %s version %s", chartName, chartVersion)
			}
			chartURL = entry.URLs[0]
			break
		}
	}
	if chartURL == "" {
		return "", fmt.Errorf("version %s not found for chart %s", chartVersion, chartName)
	}
	resp, err := http.NewRequestWithContext(context.Background(), http.MethodGet, chartURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request for chart: %w", err)
	}
	client := &http.Client{}
	resp2, err := client.Do(resp)
	if err != nil {
		return "", fmt.Errorf("failed to download chart: %w", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		return "", fmt.Errorf("failed to download chart: status %d", resp2.StatusCode)
	}
	gzr, err := gzip.NewReader(resp2.Body)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		if strings.HasSuffix(hdr.Name, "values.yaml") {
			data, err := io.ReadAll(tr)
			if err != nil {
				return "", fmt.Errorf("failed to read values.yaml: %w", err)
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("values.yaml not found in chart")
}
