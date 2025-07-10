package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/EffectiveSloth/flux-app-generator/internal/types"
)

// Import embedded templates from main package.
var (
	HelmRepositoryTemplate string
	HelmReleaseTemplate    string
	HelmValuesTemplate     string
	KustomizationTemplate  string
)

func generateFromTemplateString(templateStr, outputPath string, data interface{}) error {
	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", outputPath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log the error but don't return it as the main operation succeeded
			fmt.Printf("Warning: failed to close file %s: %v\n", outputPath, closeErr)
		}
	}()

	return tmpl.Execute(file, data)
}

func generateHelmRepository(config *types.AppConfig, appDir string) error {
	return generateFromTemplateString(
		HelmRepositoryTemplate,
		filepath.Join(appDir, "dependencies", "helm-repository.yaml"),
		config,
	)
}

func generateHelmRelease(config *types.AppConfig, appDir string) error {
	return generateFromTemplateString(
		HelmReleaseTemplate,
		filepath.Join(appDir, "release", "helm-release.yaml"),
		config,
	)
}

func generateHelmValues(config *types.AppConfig, appDir string) error {
	outputPath := filepath.Join(appDir, "release", "helm-values.yaml")
	if raw, ok := config.Values["__raw_yaml__"]; ok {
		// Write raw YAML directly
		return os.WriteFile(outputPath, []byte(raw.(string)), 0600)
	}
	// Just create an empty file
	return os.WriteFile(outputPath, []byte{}, 0600)
}

func generateKustomization(config *types.AppConfig, appDir string) error {
	return generateFromTemplateString(
		KustomizationTemplate,
		filepath.Join(appDir, "kustomization.yaml"),
		config,
	)
}

// GenerateFluxStructure is the main entrypoint for generating the Flux structure.
func GenerateFluxStructure(config *types.AppConfig) error {
	// Create app directory
	appDir := config.AppName
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("failed to create app directory %s: %w", appDir, err)
	}

	// Create subdirectories
	dirs := []string{filepath.Join(appDir, "dependencies"), filepath.Join(appDir, "release")}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	if err := generateHelmRepository(config, appDir); err != nil {
		return err
	}
	if err := generateHelmRelease(config, appDir); err != nil {
		return err
	}
	if err := generateHelmValues(config, appDir); err != nil {
		return err
	}
	if err := generateKustomization(config, appDir); err != nil {
		return err
	}

	fmt.Printf("\n✅ Generated Flux structure for '%s' in namespace '%s'\n", config.AppName, config.Namespace)
	fmt.Printf("📁 Files created in directory: %s/\n", appDir)
	return nil
}
