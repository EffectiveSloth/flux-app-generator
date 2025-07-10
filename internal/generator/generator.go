package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/EffectiveSloth/flux-app-generator/internal/kubernetes"
	"github.com/EffectiveSloth/flux-app-generator/internal/plugins"
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

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	// Ensure file ends with a newline
	_, err = file.WriteString("\n")
	return err
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
		// Write raw YAML directly, ensuring it ends with a newline
		content := raw.(string)
		if content != "" && content[len(content)-1] != '\n' {
			content += "\n"
		}
		return os.WriteFile(outputPath, []byte(content), 0600)
	}
	// Create an empty file with just a newline
	return os.WriteFile(outputPath, []byte("\n"), 0600)
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

	// Generate plugin files first
	pluginFiles, err := generatePluginFiles(config, appDir)
	if err != nil {
		return err
	}
	config.PluginFiles = pluginFiles

	// Generate kustomization.yaml after plugin files are generated
	if err := generateKustomization(config, appDir); err != nil {
		return err
	}

	fmt.Printf("\n✅ Generated Flux structure for '%s' in namespace '%s'\n", config.AppName, config.Namespace)
	fmt.Printf("📁 Files created in directory: %s/\n", appDir)

	// Print plugin summary if any plugins were generated
	if len(config.Plugins) > 0 {
		fmt.Printf("🔌 Generated %d plugin file(s):\n", len(config.Plugins))
		for _, pluginConfig := range config.Plugins {
			fmt.Printf("   - %s\n", pluginConfig.PluginName)
		}
	}

	return nil
}

// generatePluginFiles generates files for all configured plugins and returns their paths.
func generatePluginFiles(config *types.AppConfig, appDir string) ([]string, error) {
	if len(config.Plugins) == 0 {
		return nil, nil // No plugins to generate
	}

	// Create plugin registry to access plugin definitions
	pluginRegistry := plugins.NewRegistry(&kubernetes.MockKubeLister{})
	var pluginFiles []string

	for _, pluginConfig := range config.Plugins {
		plugin, exists := pluginRegistry.Get(pluginConfig.PluginName)
		if !exists {
			return nil, fmt.Errorf("plugin '%s' not found in registry", pluginConfig.PluginName)
		}

		// Validate plugin configuration
		if err := plugin.Validate(pluginConfig.Values); err != nil {
			return nil, fmt.Errorf("validation failed for plugin '%s': %w", pluginConfig.PluginName, err)
		}

		// Special handling for imageupdate plugin which generates multiple files
		if pluginConfig.PluginName == "imageupdate" {
			// Generate the plugin files
			if err := plugin.GenerateFile(pluginConfig.Values, appDir, config.Namespace); err != nil {
				return nil, fmt.Errorf("failed to generate file for plugin '%s': %w", pluginConfig.PluginName, err)
			}

			// Add all three imageupdate files to kustomization
			imageUpdateFiles := []string{
				"image-repository.yaml",
				"image-policy.yaml",
				"image-update-automation.yaml",
			}
			pluginFiles = append(pluginFiles, imageUpdateFiles...)

			fmt.Printf("✅ Generated %s plugin files\n", pluginConfig.PluginName)
			continue
		}

		// Regular plugin handling
		// Get the file path that will be generated
		templateData := make(map[string]interface{})
		for k, v := range pluginConfig.Values {
			templateData[k] = v
		}
		templateData["Namespace"] = config.Namespace

		// Parse the file path template to get the actual path
		pathTmpl, err := template.New("filepath").Parse(plugin.FilePath())
		if err != nil {
			return nil, fmt.Errorf("failed to parse file path template for plugin '%s': %w", pluginConfig.PluginName, err)
		}

		var pathBuf strings.Builder
		if err := pathTmpl.Execute(&pathBuf, templateData); err != nil {
			return nil, fmt.Errorf("failed to execute file path template for plugin '%s': %w", pluginConfig.PluginName, err)
		}

		filePath := pathBuf.String()
		pluginFiles = append(pluginFiles, filePath)

		// Generate the plugin file
		if err := plugin.GenerateFile(pluginConfig.Values, appDir, config.Namespace); err != nil {
			return nil, fmt.Errorf("failed to generate file for plugin '%s': %w", pluginConfig.PluginName, err)
		}

		fmt.Printf("✅ Generated %s plugin file\n", pluginConfig.PluginName)
	}

	return pluginFiles, nil
}
