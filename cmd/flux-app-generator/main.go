package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/EffectiveSloth/flux-app-generator/internal/generator"
	"github.com/EffectiveSloth/flux-app-generator/internal/helm"
	"github.com/EffectiveSloth/flux-app-generator/internal/plugins"
	"github.com/EffectiveSloth/flux-app-generator/internal/types"
)

//go:embed templates
var templatesFS embed.FS

// loadTemplate loads a template from the embedded filesystem.
func loadTemplate(name string) (string, error) {
	data, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %w", name, err)
	}
	return string(data), nil
}

// Form data variables - these will store the user's responses.
var (
	appName         string
	namespace       string
	helmRepoName    string
	helmRepoURL     string
	selectedChart   string
	selectedVersion string
	interval        string
	valuesPrefill   string
	versionFetcher  = helm.NewVersionFetcher()

	// Plugin-related variables.
	pluginRegistry  = plugins.NewRegistry()
	pluginInstances []plugins.PluginConfig // List of configured plugin instances.
)

func main() {
	// Load and set templates in the generator package
	if err := loadTemplates(); err != nil {
		log.Fatal(err)
	}

	// Set default values
	namespace = ""
	interval = "5m"
	valuesPrefill = "empty"

	// Step 1: Basic Application Info
	appInfoForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Application Name").
				Description("Enter a name for your Flux application").
				Placeholder("my-app").
				Value(&appName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("application name is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Namespace").
				Description("Kubernetes namespace for the application").
				Placeholder("default").
				Value(&namespace).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("namespace is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Helm Repository Name").
				Description("Name for the Helm repository resource").
				Placeholder("my-helm-repo").
				Value(&helmRepoName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("helm repository name is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Helm Repository URL").
				Description("URL of the Helm repository").
				Placeholder("https://helm.example.com").
				Value(&helmRepoURL).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("helm repository URL is required")
					}
					return nil
				}),
		).Title("📝 Application Configuration"),
	).WithTheme(huh.ThemeCharm())

	if err := appInfoForm.Run(); err != nil {
		log.Fatal(err)
	}

	// Step 2: Chart Selection
	chartForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Chart").
				Description("Choose a chart from the Helm repository").
				OptionsFunc(func() []huh.Option[string] {
					if helmRepoURL == "" {
						return []huh.Option[string]{huh.NewOption("Please enter repository URL first", "")}
					}

					// Fetch charts from repository
					charts, err := versionFetcher.ListCharts(helmRepoURL)
					if err != nil {
						return []huh.Option[string]{huh.NewOption(fmt.Sprintf("Error: %s", err.Error()), "")}
					}

					options := make([]huh.Option[string], len(charts))
					for i, chart := range charts {
						displayName := chart.Name
						if chart.Description != "" {
							displayName = fmt.Sprintf("%s - %s", chart.Name, chart.Description)
						}
						options[i] = huh.NewOption(displayName, chart.Name)
					}
					return options
				}, &helmRepoURL).
				Value(&selectedChart),

			huh.NewSelect[string]().
				Title("Select Version").
				Description("Choose a version of the selected chart").
				OptionsFunc(func() []huh.Option[string] {
					if selectedChart == "" || helmRepoURL == "" {
						return []huh.Option[string]{huh.NewOption("Please select a chart first", "")}
					}

					// Fetch versions for the selected chart
					versions, err := versionFetcher.FetchChartVersions(helmRepoURL, selectedChart)
					if err != nil {
						return []huh.Option[string]{huh.NewOption(fmt.Sprintf("Error: %s", err.Error()), "")}
					}

					options := make([]huh.Option[string], len(versions))
					for i, version := range versions {
						displayName := fmt.Sprintf("%s (App: %s)", version.ChartVersion, version.AppVersion)
						if version.Description != "" {
							displayName = fmt.Sprintf("%s - %s", displayName, version.Description)
						}
						options[i] = huh.NewOption(displayName, version.ChartVersion)
					}
					return options
				}, &selectedChart).
				Value(&selectedVersion),
		).Title("📦 Chart Selection"),
	).WithTheme(huh.ThemeCharm())

	if err := chartForm.Run(); err != nil {
		log.Fatal(err)
	}

	// Step 3: Final Configuration
	finalForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Sync Interval").
				Description("How often Flux should check for changes").
				Options(
					huh.NewOption("1 minute", "1m"),
					huh.NewOption("5 minutes", "5m"),
					huh.NewOption("10 minutes", "10m"),
					huh.NewOption("30 minutes", "30m"),
					huh.NewOption("1 hour", "1h"),
				).
				Value(&interval),

			huh.NewSelect[string]().
				Title("Values Configuration").
				Description("How to initialize the Helm values file").
				Options(
					huh.NewOption("Use default values from chart", "default"),
					huh.NewOption("Create empty values file", "empty"),
				).
				Value(&valuesPrefill),
		).Title("⚙️  Configuration"),
	).WithTheme(huh.ThemeCharm())

	if err := finalForm.Run(); err != nil {
		log.Fatal(err)
	}

	// Validate all required fields are filled
	if appName == "" || helmRepoName == "" || helmRepoURL == "" || selectedChart == "" || selectedVersion == "" {
		fmt.Println("❌ Missing required information. Please run the application again.")
		os.Exit(1)
	}

	// Step 4: Interactive Plugin Menu
	if err := runInteractivePluginMenu(); err != nil {
		log.Fatal(err)
	}

	// Create configuration
	config := &types.AppConfig{
		AppName:      appName,
		Namespace:    namespace,
		HelmRepoName: helmRepoName,
		HelmRepoURL:  helmRepoURL,
		ChartName:    selectedChart,
		ChartVersion: selectedVersion,
		Interval:     interval,
		Values:       make(map[string]interface{}),
		Plugins:      pluginInstances, // Use the new plugin instances list
		PluginFiles:  []string{},      // Will be populated by generatePluginFiles
	}

	// Handle values prefill
	if valuesPrefill == "default" {
		// Download and extract default values.yaml from the chart tarball
		fmt.Println("📦 Downloading chart and extracting default values...")
		values, err := helm.DownloadAndExtractValuesYAML(helmRepoURL, selectedChart, selectedVersion)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to download default values: %s\n", err.Error())
			fmt.Println("📝 Creating empty values file instead...")
			config.Values["__raw_yaml__"] = "# Failed to download default values for " + selectedChart + "\n# Error: " + err.Error() + "\n"
		} else {
			fmt.Println("✅ Successfully extracted default values from chart")
			config.Values["__raw_yaml__"] = values
		}
	}

	// Generate the Flux structure
	if err := generator.GenerateFluxStructure(config); err != nil {
		log.Fatal(err)
	}

	// Success message
	fmt.Printf("\n🎉 Successfully generated Flux GitOps structure!\n")
	fmt.Printf("📁 Application: %s\n", appName)
	fmt.Printf("🏷️  Namespace: %s\n", namespace)
	fmt.Printf("📦 Chart: %s@%s\n", selectedChart, selectedVersion)
	fmt.Printf("🔄 Sync Interval: %s\n", interval)

	if len(pluginInstances) > 0 {
		fmt.Printf("🔌 Plugin Instances: %d\n", len(pluginInstances))
		for i, instance := range pluginInstances {
			fmt.Printf("   %d. %s\n", i+1, instance.PluginName)
		}
	}

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   1. Review the generated files in the '%s/' directory\n", appName)
	fmt.Printf("   2. Customize the values in '%s/release/helm-values.yaml'\n", appName)
	fmt.Printf("   3. Commit to your Git repository\n")
	fmt.Printf("   4. Apply to your cluster: kubectl apply -k %s/\n", appName)
}

// loadTemplates loads all template files and sets them in the generator package.
func loadTemplates() error {
	templates := map[string]*string{
		"helm-repository.yaml.tmpl": &generator.HelmRepositoryTemplate,
		"helm-release.yaml.tmpl":    &generator.HelmReleaseTemplate,
		"kustomization.yaml.tmpl":   &generator.KustomizationTemplate,
	}

	for filename, target := range templates {
		content, err := loadTemplate(filename)
		if err != nil {
			return err
		}
		*target = content
	}

	return nil
}

// runInteractivePluginMenu provides an interactive menu for managing plugin instances.
func runInteractivePluginMenu() error {
	for {
		// Build menu options
		var options []huh.Option[string]

		// Add available plugins
		availablePlugins := pluginRegistry.List()
		for _, plugin := range availablePlugins {
			options = append(options, huh.NewOption(
				fmt.Sprintf("➕ Add %s - %s", plugin.Name(), plugin.Description()),
				fmt.Sprintf("add_%s", plugin.Name()),
			))
		}

		// Add done option
		options = append(options, huh.NewOption("✅ Done with plugins", "done"))

		// Build description with current plugin instances
		var description string
		if len(pluginInstances) == 0 {
			description = "No plugin instances configured yet. Select a plugin to add."
		} else {
			description = fmt.Sprintf("Currently configured: %d plugin instance(s):\n", len(pluginInstances))
			for i, instance := range pluginInstances {
				// Get a brief description of this instance
				instanceDesc := getPluginInstanceDescription(instance)
				description += fmt.Sprintf("  %d. %s - %s\n", i+1, instance.PluginName, instanceDesc)
			}
			description += "\nSelect a plugin to add another instance, or choose Done."
		}

		var choice string
		pluginMenuForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Plugin Management").
					Description(description).
					Options(options...).
					Value(&choice),
			).Title("🔌 Plugin Manager"),
		).WithTheme(huh.ThemeCharm())

		if err := pluginMenuForm.Run(); err != nil {
			return err
		}

		switch choice {
		case "done":
			return nil

		default:
			if strings.HasPrefix(choice, "add_") {
				pluginName := strings.TrimPrefix(choice, "add_")
				if err := configurePluginInstance(pluginName); err != nil {
					return fmt.Errorf("error configuring plugin '%s': %w", pluginName, err)
				}
			}
		}
	}
}

// getPluginInstanceDescription returns a brief description of a plugin instance.
func getPluginInstanceDescription(instance plugins.PluginConfig) string {
	// For external secret, show the target secret name if available
	if instance.PluginName == "externalsecret" {
		if targetName, ok := instance.Values["target_secret_name"].(string); ok && targetName != "" {
			return targetName
		}
		if name, ok := instance.Values["name"].(string); ok && name != "" {
			return name
		}
	}

	// For other plugins, try to find a meaningful identifier
	if name, ok := instance.Values["name"].(string); ok && name != "" {
		return name
	}

	return "configured"
}

// configurePluginInstance handles configuration of a single plugin instance.
func configurePluginInstance(pluginName string) error {
	plugin, exists := pluginRegistry.Get(pluginName)
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	variables := plugin.Variables()
	if len(variables) == 0 {
		// Plugin has no variables, just add it
		pluginInstances = append(pluginInstances, plugins.PluginConfig{
			PluginName: pluginName,
			Values:     make(map[string]interface{}),
		})
		return nil
	}

	// Create storage for this plugin instance's values
	pluginValues := make(map[string]interface{})

	// Create form fields for each variable
	var fields []huh.Field

	for _, variable := range variables {
		switch variable.Type {
		case plugins.VariableTypeText:
			var value string
			if variable.Default != nil {
				if defaultStr, ok := variable.Default.(string); ok {
					value = defaultStr
				}
			}

			field := huh.NewInput().
				Title(variable.Name).
				Description(variable.Description).
				Value(&value)

			if variable.Required {
				field = field.Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("%s is required", variable.Name)
					}
					pluginValues[variable.Name] = s
					return nil
				})
			} else {
				field = field.Validate(func(s string) error {
					pluginValues[variable.Name] = s
					return nil
				})
			}

			fields = append(fields, field)

		case plugins.VariableTypeBool, plugins.VariableTypeCheckbox:
			var value bool
			if variable.Default != nil {
				if defaultBool, ok := variable.Default.(bool); ok {
					value = defaultBool
				}
			}

			field := huh.NewConfirm().
				Title(variable.Name).
				Description(variable.Description).
				Value(&value).
				Validate(func(b bool) error {
					pluginValues[variable.Name] = b
					return nil
				})

			fields = append(fields, field)

		case plugins.VariableTypeSelect:
			var value string
			if variable.Default != nil {
				if defaultStr, ok := variable.Default.(string); ok {
					value = defaultStr
				}
			}

			options := make([]huh.Option[string], len(variable.Options))
			for i, option := range variable.Options {
				optionValue := ""
				if str, ok := option.Value.(string); ok {
					optionValue = str
				}
				options[i] = huh.NewOption(option.Label, optionValue)
			}

			field := huh.NewSelect[string]().
				Title(variable.Name).
				Description(variable.Description).
				Options(options...).
				Value(&value).
				Validate(func(s string) error {
					pluginValues[variable.Name] = s
					return nil
				})

			if variable.Required {
				field = field.Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("%s is required", variable.Name)
					}
					pluginValues[variable.Name] = s
					return nil
				})
			}

			fields = append(fields, field)
		}
	}

	if len(fields) > 0 {
		// Create and run form for this plugin instance
		configForm := huh.NewForm(
			huh.NewGroup(fields...).Title(fmt.Sprintf("🔧 Configure %s Plugin Instance", plugin.Name())),
		).WithTheme(huh.ThemeCharm())

		if err := configForm.Run(); err != nil {
			return fmt.Errorf("error collecting configuration: %w", err)
		}
	}

	// Add the configured instance
	pluginInstances = append(pluginInstances, plugins.PluginConfig{
		PluginName: pluginName,
		Values:     pluginValues,
	})

	return nil
}
