package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

// TUIProvider provides TUI integration for Kubernetes auto-completion.
type TUIProvider struct {
	autoComplete *AutoCompleteService
}

// NewTUIProvider creates a new TUI provider for Kubernetes auto-completion.
func NewTUIProvider(autoComplete *AutoCompleteService) *TUIProvider {
	return &TUIProvider{
		autoComplete: autoComplete,
	}
}

// createCustomInput creates an input field with custom key bindings for auto-completion.
func (tp *TUIProvider) createCustomInput(title, description, placeholder string, value *string, suggestionsFunc func() []string) *huh.Input {
	// Create custom key map
	keyMap := huh.NewDefaultKeyMap()

	// Set AcceptSuggestion to default (ctrl+e), Tab to next, Enter to submit
	// Do not override AcceptSuggestion, use default
	keyMap.Input.Submit = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	)
	keyMap.Input.Next = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	)

	// Handle nil value by creating a pointer to an empty string
	if value == nil {
		emptyValue := ""
		value = &emptyValue
	}

	input := huh.NewInput().
		Title(title).
		Description(description).
		Placeholder(placeholder).
		Value(value).
		SuggestionsFunc(suggestionsFunc, value).
		WithKeyMap(keyMap)

	// Type assertion to get back to *Input
	return input.(*huh.Input)
}

// TextInput creates a simple text input field without auto-completion.
func (tp *TUIProvider) TextInput(title, description, placeholder string, value *string) *huh.Input {
	return huh.NewInput().
		Title(title).
		Description(description).
		Placeholder(placeholder).
		Value(value)
}

// NamespaceInput creates an input field with namespace auto-completion.
func (tp *TUIProvider) NamespaceInput(title, description, placeholder string, value *string) *huh.Input {
	return tp.createCustomInput(
		title,
		description,
		placeholder,
		value,
		func() []string {
			ctx := context.Background()
			query := ""
			if value != nil {
				query = *value
			}
			suggestions, err := tp.autoComplete.GetNamespaceSuggestions(ctx, query)
			if err != nil {
				return []string{}
			}
			return suggestions
		},
	).Validate(func(s string) error {
		if s == "" {
			return fmt.Errorf("namespace is required")
		}
		return nil
	})
}

// ServiceInput creates an input field with service auto-completion for a specific namespace.
func (tp *TUIProvider) ServiceInput(title, description, placeholder, namespace string, value *string) *huh.Input {
	return tp.createCustomInput(
		title,
		description,
		placeholder,
		value,
		func() []string {
			if namespace == "" {
				return []string{}
			}
			ctx := context.Background()
			query := ""
			if value != nil {
				query = *value
			}
			suggestions, err := tp.autoComplete.GetServiceSuggestions(ctx, namespace, query)
			if err != nil {
				return []string{}
			}
			return suggestions
		},
	)
}

// ConfigMapInput creates an input field with configmap auto-completion for a specific namespace.
func (tp *TUIProvider) ConfigMapInput(title, description, placeholder, namespace string, value *string) *huh.Input {
	return tp.createCustomInput(
		title,
		description,
		placeholder,
		value,
		func() []string {
			if namespace == "" {
				return []string{}
			}
			ctx := context.Background()
			query := ""
			if value != nil {
				query = *value
			}
			suggestions, err := tp.autoComplete.GetConfigMapSuggestions(ctx, namespace, query)
			if err != nil {
				return []string{}
			}
			return suggestions
		},
	)
}

// SecretInput creates an input field with secret auto-completion for a specific namespace.
func (tp *TUIProvider) SecretInput(title, description, placeholder, namespace string, value *string) *huh.Input {
	return tp.createCustomInput(
		title,
		description,
		placeholder,
		value,
		func() []string {
			if namespace == "" {
				return []string{}
			}
			ctx := context.Background()
			query := ""
			if value != nil {
				query = *value
			}
			suggestions, err := tp.autoComplete.GetSecretSuggestions(ctx, namespace, query)
			if err != nil {
				return []string{}
			}
			return suggestions
		},
	)
}

// DeploymentInput creates an input field with deployment auto-completion for a specific namespace.
func (tp *TUIProvider) DeploymentInput(title, description, placeholder, namespace string, value *string) *huh.Input {
	return tp.createCustomInput(
		title,
		description,
		placeholder,
		value,
		func() []string {
			if namespace == "" {
				return []string{}
			}
			ctx := context.Background()
			query := ""
			if value != nil {
				query = *value
			}
			suggestions, err := tp.autoComplete.GetDeploymentSuggestions(ctx, namespace, query)
			if err != nil {
				return []string{}
			}
			return suggestions
		},
	)
}

// StatefulSetInput creates an input field with statefulset auto-completion for a specific namespace.
func (tp *TUIProvider) StatefulSetInput(title, description, placeholder, namespace string, value *string) *huh.Input {
	return tp.createCustomInput(
		title,
		description,
		placeholder,
		value,
		func() []string {
			if namespace == "" {
				return []string{}
			}
			ctx := context.Background()
			query := ""
			if value != nil {
				query = *value
			}
			suggestions, err := tp.autoComplete.GetStatefulSetSuggestions(ctx, namespace, query)
			if err != nil {
				return []string{}
			}
			return suggestions
		},
	)
}

// DaemonSetInput creates an input field with daemonset auto-completion for a specific namespace.
func (tp *TUIProvider) DaemonSetInput(title, description, placeholder, namespace string, value *string) *huh.Input {
	return tp.createCustomInput(
		title,
		description,
		placeholder,
		value,
		func() []string {
			if namespace == "" {
				return []string{}
			}
			ctx := context.Background()
			query := ""
			if value != nil {
				query = *value
			}
			suggestions, err := tp.autoComplete.GetDaemonSetSuggestions(ctx, namespace, query)
			if err != nil {
				return []string{}
			}
			return suggestions
		},
	)
}

// PVCInput creates an input field with PVC auto-completion for a specific namespace.
func (tp *TUIProvider) PVCInput(title, description, placeholder, namespace string, value *string) *huh.Input {
	return tp.createCustomInput(
		title,
		description,
		placeholder,
		value,
		func() []string {
			if namespace == "" {
				return []string{}
			}
			ctx := context.Background()
			query := ""
			if value != nil {
				query = *value
			}
			suggestions, err := tp.autoComplete.GetPVCSuggestions(ctx, namespace, query)
			if err != nil {
				return []string{}
			}
			return suggestions
		},
	)
}

// ResourceSelect creates a select field with resource auto-completion for a specific namespace.
func (tp *TUIProvider) ResourceSelect(title, description, namespace string, resourceType ResourceType, value *string) *huh.Select[string] {
	return huh.NewSelect[string]().
		Title(title).
		Description(description).
		OptionsFunc(func() []huh.Option[string] {
			if namespace == "" {
				return []huh.Option[string]{huh.NewOption("Please select a namespace first", "")}
			}

			ctx := context.Background()
			suggestions, err := tp.autoComplete.GetSuggestions(ctx, resourceType, namespace, "")
			if err != nil {
				return []huh.Option[string]{huh.NewOption(fmt.Sprintf("Error: %s", err.Error()), "")}
			}

			options := make([]huh.Option[string], len(suggestions))
			for i, suggestion := range suggestions {
				options[i] = huh.NewOption(suggestion, suggestion)
			}
			return options
		}, &namespace).
		Value(value)
}

// NamespaceSelect creates a select field with namespace options.
func (tp *TUIProvider) NamespaceSelect(title, description string, value *string) *huh.Select[string] {
	return huh.NewSelect[string]().
		Title(title).
		Description(description).
		OptionsFunc(func() []huh.Option[string] {
			ctx := context.Background()
			suggestions, err := tp.autoComplete.GetNamespaceSuggestions(ctx, "")
			if err != nil {
				return []huh.Option[string]{huh.NewOption(fmt.Sprintf("Error: %s", err.Error()), "")}
			}

			options := make([]huh.Option[string], len(suggestions))
			for i, suggestion := range suggestions {
				options[i] = huh.NewOption(suggestion, suggestion)
			}
			return options
		}, nil).
		Value(value)
}

// GetResourceTypeFromString converts a string to ResourceType.
func GetResourceTypeFromString(s string) ResourceType {
	switch strings.ToLower(s) {
	case "namespace":
		return ResourceTypeNamespace
	case "service":
		return ResourceTypeService
	case "configmap":
		return ResourceTypeConfigMap
	case "secret":
		return ResourceTypeSecret
	case "pod":
		return ResourceTypePod
	case "deployment":
		return ResourceTypeDeployment
	case "statefulset":
		return ResourceTypeStatefulSet
	case "daemonset":
		return ResourceTypeDaemonSet
	case "pvc":
		return ResourceTypePersistentVolumeClaim
	case "clustersecretstore":
		return ResourceTypeClusterSecretStore
	case "secretstore":
		return ResourceTypeSecretStore
	default:
		return ResourceTypeNamespace
	}
}
