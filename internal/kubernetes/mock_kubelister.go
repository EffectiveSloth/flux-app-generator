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

func (m *MockKubeLister) GetNamespaces(ctx context.Context) ([]string, error) {
	return []string{"default", "kube-system", "kube-public"}, nil
}

func (m *MockKubeLister) GetServices(ctx context.Context, namespace string) ([]string, error) {
	return []string{"kubernetes", "nginx-service"}, nil
}

func (m *MockKubeLister) GetConfigMaps(ctx context.Context, namespace string) ([]string, error) {
	return []string{"kube-root-ca.crt", "my-config"}, nil
}

func (m *MockKubeLister) GetSecrets(ctx context.Context, namespace string) ([]string, error) {
	return []string{"default-token-abc123", "my-secret"}, nil
}

func (m *MockKubeLister) GetPods(ctx context.Context, namespace string) ([]string, error) {
	return []string{"nginx-pod", "app-pod"}, nil
}

func (m *MockKubeLister) GetDeployments(ctx context.Context, namespace string) ([]string, error) {
	return []string{"nginx-deployment", "app-deployment"}, nil
}

func (m *MockKubeLister) GetStatefulSets(ctx context.Context, namespace string) ([]string, error) {
	return []string{"redis-statefulset", "mysql-statefulset"}, nil
}

func (m *MockKubeLister) GetDaemonSets(ctx context.Context, namespace string) ([]string, error) {
	return []string{"fluentd-daemonset", "node-exporter"}, nil
}

func (m *MockKubeLister) GetPersistentVolumeClaims(ctx context.Context, namespace string) ([]string, error) {
	return []string{"data-pvc", "backup-pvc"}, nil
}

func (m *MockKubeLister) GetClusterSecretStores(ctx context.Context) ([]string, error) {
	return []string{"vault-backend", "aws-secrets-manager", "azure-key-vault"}, nil
}

func (m *MockKubeLister) GetSecretStores(ctx context.Context, namespace string) ([]string, error) {
	return []string{"local-vault", "namespace-secrets"}, nil
}

func (m *MockKubeLister) TestConnection(ctx context.Context) error {
	return nil
}
