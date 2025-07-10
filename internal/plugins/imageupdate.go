package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
)

// ImageUpdatePlugin creates Flux image update automation resources.
type ImageUpdatePlugin struct {
	BasePlugin
}

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
		{
			Name:        "git_repository_name",
			Type:        VariableTypeText,
			Description: "Name of the GitRepository resource to reference",
			Required:    true,
			Default:     "flux-system",
		},
		{
			Name:        "git_repository_namespace",
			Type:        VariableTypeText,
			Description: "Namespace of the GitRepository resource",
			Required:    true,
			Default:     "flux-system",
		},
		{
			Name:        "update_path",
			Type:        VariableTypeText,
			Description: "Path to update in the repository",
			Required:    true,
		},
		{
			Name:        "update_strategy",
			Type:        VariableTypeSelect,
			Description: "Update strategy to use",
			Required:    true,
			Default:     "Setters",
			Options: []Option{
				{Label: "Setters", Value: "Setters"},
			},
		},
		{
			Name:        "git_branch",
			Type:        VariableTypeText,
			Description: "Git branch to push updates to",
			Required:    true,
			Default:     "main",
		},
		{
			Name:        "author_name",
			Type:        VariableTypeText,
			Description: "Author name for git commits",
			Required:    true,
		},
		{
			Name:        "author_email",
			Type:        VariableTypeText,
			Description: "Author email for git commits",
			Required:    true,
		},
		{
			Name:        "commit_message_template",
			Type:        VariableTypeText,
			Description: "Template for commit messages",
			Required:    true,
			Default:     "chore: update container versions",
		},
		{
			Name:        "automation_interval",
			Type:        VariableTypeSelect,
			Description: "How often to check for updates",
			Required:    true,
			Default:     "1m",
			Options: []Option{
				{Label: "1 minute", Value: "1m"},
				{Label: "5 minutes", Value: "5m"},
				{Label: "10 minutes", Value: "10m"},
				{Label: "30 minutes", Value: "30m"},
				{Label: "1 hour", Value: "60m"},
			},
		},
		{
			Name:        "image_repositories",
			Type:        VariableTypeText,
			Description: "JSON array of image repository configurations",
			Required:    true,
		},
		{
			Name:        "image_policies",
			Type:        VariableTypeText,
			Description: "JSON array of image policy configurations",
			Required:    true,
		},
	}

	// This plugin will generate multiple files, so we'll use a special approach
	filePath := "update/"

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
	if repoData, exists := values["image_repositories"]; exists {
		if repoStr, ok := repoData.(string); ok {
			var repos []ImageRepository
			if err := json.Unmarshal([]byte(repoStr), &repos); err != nil {
				return &ValidationError{
					Variable: "image_repositories",
					Message:  fmt.Sprintf("invalid JSON format: %v", err),
				}
			}
			// Validate each repository
			for i, repo := range repos {
				if repo.Name == "" {
					return &ValidationError{
						Variable: "image_repositories",
						Message:  fmt.Sprintf("repository %d: name is required", i),
					}
				}
				if repo.Image == "" {
					return &ValidationError{
						Variable: "image_repositories",
						Message:  fmt.Sprintf("repository %d: image is required", i),
					}
				}
				if repo.Interval == "" {
					return &ValidationError{
						Variable: "image_repositories",
						Message:  fmt.Sprintf("repository %d: interval is required", i),
					}
				}
			}
		}
	}

	// Validate image_policies JSON
	if policyData, exists := values["image_policies"]; exists {
		if policyStr, ok := policyData.(string); ok {
			var policies []ImagePolicy
			if err := json.Unmarshal([]byte(policyStr), &policies); err != nil {
				return &ValidationError{
					Variable: "image_policies",
					Message:  fmt.Sprintf("invalid JSON format: %v", err),
				}
			}
			// Validate each policy
			for i, policy := range policies {
				if policy.Name == "" {
					return &ValidationError{
						Variable: "image_policies",
						Message:  fmt.Sprintf("policy %d: name is required", i),
					}
				}
				if policy.Repository == "" {
					return &ValidationError{
						Variable: "image_policies",
						Message:  fmt.Sprintf("policy %d: repository is required", i),
					}
				}
				if policy.PolicyType == "" {
					return &ValidationError{
						Variable: "image_policies",
						Message:  fmt.Sprintf("policy %d: policyType is required", i),
					}
				}
				if policy.PolicyType == "semver" && policy.Range == "" {
					return &ValidationError{
						Variable: "image_policies",
						Message:  fmt.Sprintf("policy %d: range is required for semver policy", i),
					}
				}
				if policy.PolicyType == "numerical" {
					if policy.Pattern == "" {
						return &ValidationError{
							Variable: "image_policies",
							Message:  fmt.Sprintf("policy %d: pattern is required for numerical policy", i),
						}
					}
					if policy.Extract == "" {
						return &ValidationError{
							Variable: "image_policies",
							Message:  fmt.Sprintf("policy %d: extract is required for numerical policy", i),
						}
					}
					if policy.Order == "" {
						return &ValidationError{
							Variable: "image_policies",
							Message:  fmt.Sprintf("policy %d: order is required for numerical policy", i),
						}
					}
				}
			}
		}
	}

	return nil
}

// GenerateFile creates the three image update automation files.
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

	// Create update directory
	updateDir := fmt.Sprintf("%s/update", appDir)
	if err := os.MkdirAll(updateDir, 0755); err != nil {
		return fmt.Errorf("failed to create update directory: %v", err)
	}

	// Generate the three files
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
		if err := p.generateSingleFile(templateStr, fmt.Sprintf("%s/%s", updateDir, filename), templateData); err != nil {
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