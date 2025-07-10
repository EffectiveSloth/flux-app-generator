package plugins

// ExternalSecretPlugin creates ExternalSecret resources for Kubernetes.
type ExternalSecretPlugin struct {
	BasePlugin
}

// NewExternalSecretPlugin creates a new ExternalSecret plugin instance.
func NewExternalSecretPlugin() *ExternalSecretPlugin {
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
	}
}
