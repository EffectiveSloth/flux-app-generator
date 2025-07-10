package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/huh"

	"github.com/EffectiveSloth/flux-app-generator/internal/generator"
	"github.com/EffectiveSloth/flux-app-generator/internal/helm"
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
)

func main() {
	// Load and set templates in the generator package
	if err := loadTemplates(); err != nil {
		log.Fatal(err)
	}

	// Set default values
	namespace = "default"
	interval = "5m"
	valuesPrefill = "empty"

	// Create the form with multiple groups
	form := huh.NewForm(
		// Group 1: Basic Application Info
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
						namespace = "default"
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

		// Group 2: Chart Selection (Dynamic)
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

		// Group 3: Final Configuration
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

	// Run the form
	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	// Validate all required fields are filled
	if appName == "" || helmRepoName == "" || helmRepoURL == "" || selectedChart == "" || selectedVersion == "" {
		fmt.Println("❌ Missing required information. Please run the application again.")
		os.Exit(1)
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
