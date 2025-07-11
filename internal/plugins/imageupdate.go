package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/charmbracelet/huh"
)

const (
	// PolicyTypeSemver represents semantic version policy type.
	PolicyTypeSemver = "semver"
	// PolicyTypeTimestamp represents timestamp policy type.
	PolicyTypeTimestamp = "timestamp"
	// PolicyTypeNumerical represents numerical policy type.
	PolicyTypeNumerical = "numerical"

	// DefaultFluxNamespace is the default namespace where Flux is installed.
	DefaultFluxNamespace = "flux-system"
)

// ImageUpdatePlugin creates Flux image update automation resources.
type ImageUpdatePlugin struct {
	BasePlugin
}

// Ensure ImageUpdatePlugin implements CustomConfigPlugin.
var _ CustomConfigPlugin = (*ImageUpdatePlugin)(nil)

// ImageRepository represents a single image repository configuration.
type ImageRepository struct {
	Name      string `json:"name" yaml:"name"`
	Image     string `json:"image" yaml:"image"`
	Interval  string `json:"interval" yaml:"interval"`
	SecretRef string `json:"secretRef,omitempty" yaml:"secretRef,omitempty"`
}

// ImagePolicy represents a single image policy configuration.
type ImagePolicy struct {
	Name       string `json:"name" yaml:"name"`
	Repository string `json:"repository" yaml:"repository"`
	PolicyType string `json:"policyType" yaml:"policyType"`
	Range      string `json:"range,omitempty" yaml:"range,omitempty"`
	Pattern    string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Extract    string `json:"extract,omitempty" yaml:"extract,omitempty"`
	Order      string `json:"order,omitempty" yaml:"order,omitempty"`
}

// NewImageUpdatePlugin creates a new image update automation plugin instance.
func NewImageUpdatePlugin() *ImageUpdatePlugin {
	variables := []Variable{
		{
			Name:        "automation_name",
			Type:        VariableTypeText,
			Description: "Name for the ImageUpdateAutomation resource",
			Required:    true,
		},
	}

	// This plugin will generate multiple files directly in the main directory
	filePath := "image-update-automation.yaml"

	return &ImageUpdatePlugin{
		BasePlugin: BasePlugin{
			name:        "imageupdate",
			description: "Generates Flux image update automation resources for automatic container image updates",
			variables:   variables,
			template:    "", // We'll override GenerateFile method
			filePath:    filePath,
		},
	}
}

// Validate performs validation specific to the image update plugin.
func (p *ImageUpdatePlugin) Validate(values map[string]interface{}) error {
	// First, perform base validation
	if err := p.BasePlugin.Validate(values); err != nil {
		return err
	}

	// Validate image_repositories JSON
	if err := p.validateImageRepositories(values); err != nil {
		return err
	}

	// Validate image_policies JSON
	if err := p.validateImagePolicies(values); err != nil {
		return err
	}

	return nil
}

// validateImageRepositories validates the image_repositories JSON field.
func (p *ImageUpdatePlugin) validateImageRepositories(values map[string]interface{}) error {
	return p.validateJSONField(values, "image_repositories", func(data []byte) error {
		var repos []ImageRepository
		if err := json.Unmarshal(data, &repos); err != nil {
			return err
		}

		for i, repo := range repos {
			if err := p.validateSingleRepository(repo, i); err != nil {
				return err
			}
		}
		return nil
	})
}

// validateImagePolicies validates the image_policies JSON field.
func (p *ImageUpdatePlugin) validateImagePolicies(values map[string]interface{}) error {
	return p.validateJSONField(values, "image_policies", func(data []byte) error {
		var policies []ImagePolicy
		if err := json.Unmarshal(data, &policies); err != nil {
			return err
		}

		for i, policy := range policies {
			if err := p.validateSinglePolicy(&policy, i); err != nil {
				return err
			}
		}
		return nil
	})
}

// validateJSONField provides common JSON field validation logic.
func (p *ImageUpdatePlugin) validateJSONField(values map[string]interface{}, fieldName string, validator func([]byte) error) error {
	data, exists := values[fieldName]
	if !exists {
		return nil
	}

	dataStr, ok := data.(string)
	if !ok {
		return nil
	}

	if err := validator([]byte(dataStr)); err != nil {
		return &ValidationError{
			Variable: fieldName,
			Message:  fmt.Sprintf("invalid JSON format: %v", err),
		}
	}

	return nil
}

// validateSingleRepository validates a single image repository configuration.
func (p *ImageUpdatePlugin) validateSingleRepository(repo ImageRepository, index int) error {
	if repo.Name == "" {
		return &ValidationError{
			Variable: "image_repositories",
			Message:  fmt.Sprintf("repository %d: name is required", index),
		}
	}
	if repo.Image == "" {
		return &ValidationError{
			Variable: "image_repositories",
			Message:  fmt.Sprintf("repository %d: image is required", index),
		}
	}
	if repo.Interval == "" {
		return &ValidationError{
			Variable: "image_repositories",
			Message:  fmt.Sprintf("repository %d: interval is required", index),
		}
	}
	return nil
}

// validateSinglePolicy validates a single image policy configuration.
func (p *ImageUpdatePlugin) validateSinglePolicy(policy *ImagePolicy, index int) error {
	if policy.Name == "" {
		return &ValidationError{
			Variable: "image_policies",
			Message:  fmt.Sprintf("policy %d: name is required", index),
		}
	}
	if policy.Repository == "" {
		return &ValidationError{
			Variable: "image_policies",
			Message:  fmt.Sprintf("policy %d: repository is required", index),
		}
	}
	if policy.PolicyType == "" {
		return &ValidationError{
			Variable: "image_policies",
			Message:  fmt.Sprintf("policy %d: policyType is required", index),
		}
	}

	return p.validatePolicyTypeSpecificFields(policy, index)
}

// validatePolicyTypeSpecificFields validates fields specific to each policy type.
func (p *ImageUpdatePlugin) validatePolicyTypeSpecificFields(policy *ImagePolicy, index int) error {
	switch policy.PolicyType {
	case PolicyTypeSemver:
		return p.validateSemverPolicy(policy, index)
	case PolicyTypeNumerical:
		return p.validateNumericalPolicy(policy, index)
	}
	return nil
}

// validateSemverPolicy validates semver-specific policy fields.
func (p *ImageUpdatePlugin) validateSemverPolicy(policy *ImagePolicy, index int) error {
	if policy.Range == "" {
		return &ValidationError{
			Variable: "image_policies",
			Message:  fmt.Sprintf("policy %d: range is required for semver policy", index),
		}
	}
	return nil
}

// validateNumericalPolicy validates numerical-specific policy fields.
func (p *ImageUpdatePlugin) validateNumericalPolicy(policy *ImagePolicy, index int) error {
	if policy.Pattern == "" {
		return &ValidationError{
			Variable: "image_policies",
			Message:  fmt.Sprintf("policy %d: pattern is required for numerical policy", index),
		}
	}
	if policy.Extract == "" {
		return &ValidationError{
			Variable: "image_policies",
			Message:  fmt.Sprintf("policy %d: extract is required for numerical policy", index),
		}
	}
	if policy.Order == "" {
		return &ValidationError{
			Variable: "image_policies",
			Message:  fmt.Sprintf("policy %d: order is required for numerical policy", index),
		}
	}
	return nil
}

// CollectCustomConfig handles the multi-step configuration for image update automation.
func (p *ImageUpdatePlugin) CollectCustomConfig(values map[string]interface{}) error {
	// Step 1: Configure ImageRepository
	repo, err := p.configureImageRepository()
	if err != nil {
		return fmt.Errorf("failed to configure image repository: %w", err)
	}

	// Step 2: Configure ImagePolicy
	policy, err := p.configureImagePolicy(repo.Name)
	if err != nil {
		return fmt.Errorf("failed to configure image policy: %w", err)
	}

	// Step 3: Configure ImageUpdateAutomation
	automation, err := p.configureImageUpdateAutomation()
	if err != nil {
		return fmt.Errorf("failed to configure image update automation: %w", err)
	}

	// Convert to JSON arrays (single items)
	repos := []ImageRepository{repo}
	policies := []ImagePolicy{policy}

	repoJSON, _ := json.Marshal(repos)
	policyJSON, _ := json.Marshal(policies)

	values["image_repositories"] = string(repoJSON)
	values["image_policies"] = string(policyJSON)

	// Set automation values
	values["git_repository_name"] = automation.GitRepositoryName
	values["git_repository_namespace"] = automation.GitRepositoryNamespace
	values["update_path"] = automation.UpdatePath
	values["git_branch"] = automation.GitBranch
	values["author_name"] = automation.AuthorName
	values["author_email"] = automation.AuthorEmail
	values["automation_interval"] = automation.Interval
	values["update_strategy"] = "Setters"
	values["commit_message_template"] = "chore: update container versions"

	return nil
}

// configureImageRepository handles the first step: ImageRepository configuration.
func (p *ImageUpdatePlugin) configureImageRepository() (ImageRepository, error) {
	var repo ImageRepository

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Repository Name").
				Description("Unique name for this image repository").
				Value(&repo.Name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("repository name is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Container Image").
				Description("Full image name (e.g., nginx, myregistry/myapp)").
				Value(&repo.Image).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("image name is required")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Check Interval").
				Description("How often to check for new image versions").
				Options(
					huh.NewOption("1 hour", "60m"),
					huh.NewOption("6 hours", "6h"),
					huh.NewOption("12 hours", "12h"),
					huh.NewOption("24 hours", "24h"),
				).
				Value(&repo.Interval),

			huh.NewInput().
				Title("Secret Reference (Optional)").
				Description("Name of secret for private registry (leave empty for public)").
				Value(&repo.SecretRef),
		).Title("📦 Step 1: Configure Image Repository"),
	).WithTheme(huh.ThemeCharm())

	// Set default
	repo.Interval = "6h"

	return repo, form.Run()
}

// configureImagePolicy handles the second step: ImagePolicy configuration.
func (p *ImageUpdatePlugin) configureImagePolicy(repositoryName string) (ImagePolicy, error) {
	var policy ImagePolicy
	var policyType string

	// Basic policy configuration
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Policy Name").
				Description("Name for this image policy").
				Value(&policy.Name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("policy name is required")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Version Policy").
				Description("How should image versions be evaluated?").
				Options(
					huh.NewOption("Semantic Versioning (1.2.3)", PolicyTypeSemver),
					huh.NewOption("Timestamp-based (main-abc123-1234567890)", PolicyTypeTimestamp),
				).
				Value(&policyType),
		).Title("🏷️ Step 2: Configure Image Policy"),
	).WithTheme(huh.ThemeCharm())

	// Set defaults
	policy.Repository = repositoryName
	policyType = PolicyTypeSemver

	if err := form.Run(); err != nil {
		return policy, err
	}

	// Configure policy-specific settings
	policy.PolicyType = policyType
	if policyType == PolicyTypeTimestamp {
		policy.Pattern = "^main-[a-f0-9]+-(?P<ts>[0-9]+)"
		policy.Extract = "$ts"
		policy.Order = "asc"
		policy.PolicyType = PolicyTypeNumerical
	} else {
		// Semver policy - ask for range
		var semverRange string
		rangeForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Version Range").
					Description("Which semantic versions should be considered?").
					Options(
						huh.NewOption("Any version (*)", "*"),
						huh.NewOption("Major version (^1.0.0)", "^1.0.0"),
						huh.NewOption("Minor version (~1.2.0)", "~1.2.0"),
					).
					Value(&semverRange),
			).Title("🏷️ Semantic Version Range"),
		).WithTheme(huh.ThemeCharm())

		semverRange = "*"
		if err := rangeForm.Run(); err != nil {
			return policy, err
		}
		policy.Range = semverRange
	}

	return policy, nil
}

// ImageUpdateAutomationConfig holds the automation configuration.
type ImageUpdateAutomationConfig struct {
	GitRepositoryName      string
	GitRepositoryNamespace string
	UpdatePath             string
	GitBranch              string
	AuthorName             string
	AuthorEmail            string
	Interval               string
}

// configureImageUpdateAutomation handles the third step: ImageUpdateAutomation configuration.
func (p *ImageUpdatePlugin) configureImageUpdateAutomation() (ImageUpdateAutomationConfig, error) {
	var config ImageUpdateAutomationConfig

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("GitRepository Name").
				Description("Name of the GitRepository resource to reference").
				Value(&config.GitRepositoryName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("git repository name is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("GitRepository Namespace").
				Description("Namespace of the GitRepository resource").
				Value(&config.GitRepositoryNamespace).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("git repository namespace is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Update Path").
				Description("Path to update in the repository (e.g., ./apps/myapp)").
				Value(&config.UpdatePath).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("update path is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Git Branch").
				Description("Git branch to push updates to").
				Value(&config.GitBranch).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("git branch is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Author Name").
				Description("Author name for git commits").
				Value(&config.AuthorName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("author name is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Author Email").
				Description("Author email for git commits").
				Value(&config.AuthorEmail).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("author email is required")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Automation Interval").
				Description("How often to check for updates").
				Options(
					huh.NewOption("5 minutes", "5m"),
					huh.NewOption("10 minutes", "10m"),
					huh.NewOption("30 minutes", "30m"),
					huh.NewOption("1 hour", "60m"),
				).
				Value(&config.Interval),
		).Title("⚙️ Step 3: Configure Update Automation"),
	).WithTheme(huh.ThemeCharm())

	// Set defaults
	config.GitRepositoryName = DefaultFluxNamespace
	config.GitRepositoryNamespace = DefaultFluxNamespace
	config.GitBranch = "main"
	config.Interval = "10m"

	return config, form.Run()
}

// GenerateFile creates the three image update automation files directly in the main directory.
func (p *ImageUpdatePlugin) GenerateFile(values map[string]interface{}, appDir, namespace string) error {
	// Parse image repositories and policies from JSON
	var imageRepositories []ImageRepository
	var imagePolicies []ImagePolicy

	if repoData, exists := values["image_repositories"]; exists {
		if repoStr, ok := repoData.(string); ok {
			if err := json.Unmarshal([]byte(repoStr), &imageRepositories); err != nil {
				return fmt.Errorf("failed to parse image repositories: %v", err)
			}
		}
	}

	if policyData, exists := values["image_policies"]; exists {
		if policyStr, ok := policyData.(string); ok {
			if err := json.Unmarshal([]byte(policyStr), &imagePolicies); err != nil {
				return fmt.Errorf("failed to parse image policies: %v", err)
			}
		}
	}

	// Create template data
	templateData := make(map[string]interface{})
	for k, v := range values {
		templateData[k] = v
	}
	templateData["Namespace"] = namespace
	templateData["ImageRepositories"] = imageRepositories
	templateData["ImagePolicies"] = imagePolicies

	// Generate the three files directly in the main directory
	files := map[string]string{
		"image-repository.yaml": `{{- range .ImageRepositories }}
---
apiVersion: image.toolkit.fluxcd.io/v1beta2
kind: ImageRepository
metadata:
  name: {{.Name}}
spec:
  image: {{.Image}}
  interval: {{.Interval}}{{- if .SecretRef }}
  secretRef:
    name: {{.SecretRef}}{{- end }}
{{- end }}`,
		"image-policy.yaml": `{{- range .ImagePolicies }}
---
apiVersion: image.toolkit.fluxcd.io/v1beta2
kind: ImagePolicy
metadata:
  name: {{.Name}}
spec:
  imageRepositoryRef:
    name: {{.Repository}}{{- if eq .PolicyType "semver" }}
  policy:
    semver:
      range: '{{.Range}}'{{- else if eq .PolicyType "numerical" }}
  filterTags:
    pattern: '{{.Pattern}}'
    extract: '{{.Extract}}'
  policy:
    numerical:
      order: {{.Order}}{{- end }}
{{- end }}`,
		"image-update-automation.yaml": `---
apiVersion: image.toolkit.fluxcd.io/v1beta1
kind: ImageUpdateAutomation
metadata:
  name: {{.automation_name}}
  namespace: {{.Namespace}}
spec:
  interval: {{.automation_interval}}
  sourceRef:
    kind: GitRepository
    name: {{.git_repository_name}}
    namespace: {{.git_repository_namespace}}
  git:
    commit:
      author:
        email: {{.author_email}}
        name: {{.author_name}}
      messageTemplate: "{{.commit_message_template}}"
    push:
      branch: {{.git_branch}}
  update:
    path: {{.update_path}}
    strategy: {{.update_strategy}}`,
	}

	for filename, templateStr := range files {
		outputPath := filepath.Join(appDir, filename)
		if err := p.generateSingleFile(templateStr, outputPath, templateData); err != nil {
			return fmt.Errorf("failed to generate %s: %v", filename, err)
		}
	}

	return nil
}

// generateSingleFile is a helper method to generate a single file from a template.
func (p *ImageUpdatePlugin) generateSingleFile(templateStr, outputPath string, data interface{}) error {
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
			fmt.Printf("Warning: failed to close file %s: %v\n", outputPath, closeErr)
		}
	}()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Ensure file ends with a newline
	if _, err := file.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}
