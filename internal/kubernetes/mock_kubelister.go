package kubernetes

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockKubeLister is a mock implementation of KubeLister for testing
// Embed mock.Mock so it can be used with testify's On/Called methods.
type MockKubeLister struct {
	mock.Mock
}

// GetNamespaces returns a list of mock Kubernetes namespaces.
func (m *MockKubeLister) GetNamespaces(_ context.Context) ([]string, error) {
	return []string{"default", "kube-system", "kube-public"}, nil
}

// GetServices returns a list of mock Kubernetes services in the specified namespace.
func (m *MockKubeLister) GetServices(_ context.Context, _ string) ([]string, error) {
	return []string{"kubernetes", "nginx-service"}, nil
}

// GetConfigMaps returns a list of mock Kubernetes configmaps in the specified namespace.
func (m *MockKubeLister) GetConfigMaps(_ context.Context, _ string) ([]string, error) {
	return []string{"kube-root-ca.crt", "my-config"}, nil
}

// GetSecrets returns a list of mock Kubernetes secrets in the specified namespace.
func (m *MockKubeLister) GetSecrets(_ context.Context, _ string) ([]string, error) {
	return []string{"default-token-abc123", "my-secret"}, nil
}

// GetPods returns a list of mock Kubernetes pods in the specified namespace.
func (m *MockKubeLister) GetPods(_ context.Context, _ string) ([]string, error) {
	return []string{"nginx-pod", "app-pod"}, nil
}

// GetDeployments returns a list of mock Kubernetes deployments in the specified namespace.
func (m *MockKubeLister) GetDeployments(_ context.Context, _ string) ([]string, error) {
	return []string{"nginx-deployment", "app-deployment"}, nil
}

// GetStatefulSets returns a list of mock Kubernetes statefulsets in the specified namespace.
func (m *MockKubeLister) GetStatefulSets(_ context.Context, _ string) ([]string, error) {
	return []string{"redis-statefulset", "mysql-statefulset"}, nil
}

// GetDaemonSets returns a list of mock Kubernetes daemonsets in the specified namespace.
func (m *MockKubeLister) GetDaemonSets(_ context.Context, _ string) ([]string, error) {
	return []string{"fluentd-daemonset", "node-exporter"}, nil
}

// GetPersistentVolumeClaims returns a list of mock Kubernetes persistent volume claims in the specified namespace.
func (m *MockKubeLister) GetPersistentVolumeClaims(_ context.Context, _ string) ([]string, error) {
	return []string{"data-pvc", "backup-pvc"}, nil
}

// GetClusterSecretStores returns a list of mock External Secrets Operator ClusterSecretStores.
func (m *MockKubeLister) GetClusterSecretStores(_ context.Context) ([]string, error) {
	return []string{"vault-backend", "aws-secrets-manager", "azure-key-vault"}, nil
}

// GetSecretStores returns a list of mock External Secrets Operator SecretStores in the specified namespace.
func (m *MockKubeLister) GetSecretStores(_ context.Context, _ string) ([]string, error) {
	return []string{"local-vault", "namespace-secrets"}, nil
}

// TestConnection tests the mock Kubernetes connection.
func (m *MockKubeLister) TestConnection(_ context.Context) error {
	return nil
}
