package plugins

import (
	"context"
	"fmt"

	"github.com/EffectiveSloth/flux-app-generator/internal/kubernetes"
	"github.com/charmbracelet/huh"
)

// ExternalSecretPlugin creates ExternalSecret resources for Kubernetes.
type ExternalSecretPlugin struct {
	BasePlugin
	kubeClient kubernetes.KubeLister
}

// NewExternalSecretPlugin creates a new ExternalSecret plugin instance.
func NewExternalSecretPlugin(kubeClient kubernetes.KubeLister) *ExternalSecretPlugin {
	variables := []Variable{
		{
			Name:        "name",
			Type:        VariableTypeText,
			Description: "Name for the ExternalSecret resource",
			Required:    true,
		},
		{
			Name:        "secret_store_type",
			Type:        VariableTypeSelect,
			Description: "Type of secret store to reference",
			Required:    true,
			Default:     "ClusterSecretStore",
			Options: []Option{
				{Label: "Cluster Secret Store", Value: "ClusterSecretStore"},
				{Label: "Secret Store", Value: "SecretStore"},
			},
		},
		{
			Name:        "secret_store_name",
			Type:        VariableTypeText,
			Description: "Name of the secret store resource",
			Required:    true,
		},
		{
			Name:        "secret_key",
			Type:        VariableTypeText,
			Description: "Key name in the external secret store",
			Required:    true,
		},
		{
			Name:        "target_secret_name",
			Type:        VariableTypeText,
			Description: "Name of the Kubernetes secret to create",
			Required:    true,
		},
		{
			Name:        "refresh_interval",
			Type:        VariableTypeSelect,
			Description: "How often to refresh the secret",
			Required:    false,
			Default:     "60m",
			Options: []Option{
				{Label: "15 minutes", Value: "15m"},
				{Label: "30 minutes", Value: "30m"},
				{Label: "1 hour", Value: "60m"},
				{Label: "2 hours", Value: "120m"},
				{Label: "6 hours", Value: "6h"},
				{Label: "12 hours", Value: "12h"},
				{Label: "24 hours", Value: "24h"},
			},
		},
	}

	template := `apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: {{.name}}
  namespace: {{.Namespace}}
spec:
  secretStoreRef:
    kind: {{.secret_store_type}}
    name: {{.secret_store_name}}
  dataFrom:
    - extract:
        key: {{.secret_key}}
  refreshInterval: {{.refresh_interval}}
  target:
    creationPolicy: Owner
    name: {{.target_secret_name}}`

	filePath := "dependencies/external-secret-{{.target_secret_name}}.yaml"

	return &ExternalSecretPlugin{
		BasePlugin: BasePlugin{
			name:        "externalsecret",
			description: "Generates ExternalSecret resources for managing secrets from external secret stores",
			variables:   variables,
			template:    template,
			filePath:    filePath,
		},
		kubeClient: kubeClient,
	}
}

// ConfigureWithAutoComplete provides a custom configuration flow with select dropdowns for secret stores.
func (p *ExternalSecretPlugin) ConfigureWithAutoComplete(namespace string) (map[string]interface{}, error) {
	// Create auto-complete service
	autoComplete := kubernetes.NewAutoCompleteService(p.kubeClient)
	tuiProvider := kubernetes.NewTUIProvider(autoComplete)

	// Variables to store form values
	var name, secretStoreType, secretStoreName, secretKey, targetSecretName, refreshInterval string

	// Step 1: Secret store type selection
	storeTypeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Secret Store Type").
				Description("Type of secret store to reference").
				Options(
					huh.NewOption("Cluster Secret Store", "ClusterSecretStore"),
					huh.NewOption("Secret Store", "SecretStore"),
				).
				Value(&secretStoreType),
		).Title("🔐 Secret Store Selection"),
	)

	if err := storeTypeForm.Run(); err != nil {
		return nil, err
	}

	// Step 2: Secret store name selection with fallback to manual input
	var storeNameInput huh.Field
	if secretStoreType == "ClusterSecretStore" {
		// Try to get ClusterSecretStores from Kubernetes
		if p.kubeClient != nil {
			stores, err := p.kubeClient.GetClusterSecretStores(context.Background())
			if err != nil || len(stores) == 0 {
				// Fallback to manual input if no stores found or error
				storeNameInput = tuiProvider.TextInput(
					"Secret Store Name",
					"Name of the ClusterSecretStore resource (no stores found in cluster)",
					"my-cluster-store",
					&secretStoreName,
				)
			} else {
				// Create select dropdown with available stores
				options := make([]huh.Option[string], len(stores))
				for i, store := range stores {
					options[i] = huh.NewOption(store, store)
				}
				storeNameInput = huh.NewSelect[string]().
					Title("Secret Store Name").
					Description("Select a ClusterSecretStore from the cluster").
					Options(options...).
					Value(&secretStoreName)
			}
		} else {
			// Kubernetes client not available, use manual input
			storeNameInput = tuiProvider.TextInput(
				"Secret Store Name",
				"Name of the ClusterSecretStore resource (Kubernetes client not available)",
				"my-cluster-store",
				&secretStoreName,
			)
		}
	} else {
		// Try to get SecretStores from the selected namespace
		if p.kubeClient != nil {
			stores, err := p.kubeClient.GetSecretStores(context.Background(), namespace)
			if err != nil || len(stores) == 0 {
				// Fallback to manual input if no stores found or error
				storeNameInput = tuiProvider.TextInput(
					"Secret Store Name",
					fmt.Sprintf("Name of the SecretStore resource in namespace %s (no stores found)", namespace),
					"my-secret-store",
					&secretStoreName,
				)
			} else {
				// Create select dropdown with available stores
				options := make([]huh.Option[string], len(stores))
				for i, store := range stores {
					options[i] = huh.NewOption(store, store)
				}
				storeNameInput = huh.NewSelect[string]().
					Title("Secret Store Name").
					Description(fmt.Sprintf("Select a SecretStore from namespace %s", namespace)).
					Options(options...).
					Value(&secretStoreName)
			}
		} else {
			// Kubernetes client not available, use manual input
			storeNameInput = tuiProvider.TextInput(
				"Secret Store Name",
				fmt.Sprintf("Name of the SecretStore resource in namespace %s (Kubernetes client not available)", namespace),
				"my-secret-store",
				&secretStoreName,
			)
		}
	}

	storeNameForm := huh.NewForm(
		huh.NewGroup(
			storeNameInput,
		).Title("🔐 Secret Store Selection"),
	)

	if err := storeNameForm.Run(); err != nil {
		return nil, err
	}

	// Step 3: Secret configuration
	secretForm := huh.NewForm(
		huh.NewGroup(
			tuiProvider.TextInput("Name", "Name for the ExternalSecret resource", "my-external-secret", &name),
			tuiProvider.TextInput("Secret Key", "Key name in the external secret store", "my-secret-key", &secretKey),
			tuiProvider.TextInput("Target Secret Name", "Name of the Kubernetes secret to create", "my-secret", &targetSecretName),
			huh.NewSelect[string]().
				Title("Refresh Interval").
				Description("How often to refresh the secret").
				Options(
					huh.NewOption("15 minutes", "15m"),
					huh.NewOption("30 minutes", "30m"),
					huh.NewOption("1 hour", "60m"),
					huh.NewOption("2 hours", "120m"),
					huh.NewOption("6 hours", "6h"),
					huh.NewOption("12 hours", "12h"),
					huh.NewOption("24 hours", "24h"),
				).
				Value(&refreshInterval),
		).Title("🔑 Secret Configuration"),
	)

	if err := secretForm.Run(); err != nil {
		return nil, err
	}

	// Return the configuration
	return map[string]interface{}{
		"name":               name,
		"secret_store_type":  secretStoreType,
		"secret_store_name":  secretStoreName,
		"secret_key":         secretKey,
		"target_secret_name": targetSecretName,
		"refresh_interval":   refreshInterval,
		"Namespace":          namespace,
	}, nil
}
