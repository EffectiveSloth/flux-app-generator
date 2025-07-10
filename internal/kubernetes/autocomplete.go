package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ResourceType represents different types of Kubernetes resources.
type ResourceType string

const (
	ResourceTypeNamespace             ResourceType = "namespace"
	ResourceTypeService               ResourceType = "service"
	ResourceTypeConfigMap             ResourceType = "configmap"
	ResourceTypeSecret                ResourceType = "secret"
	ResourceTypePod                   ResourceType = "pod"
	ResourceTypeDeployment            ResourceType = "deployment"
	ResourceTypeStatefulSet           ResourceType = "statefulset"
	ResourceTypeDaemonSet             ResourceType = "daemonset"
	ResourceTypePersistentVolumeClaim ResourceType = "pvc"
	ResourceTypeClusterSecretStore    ResourceType = "clustersecretstore"
	ResourceTypeSecretStore           ResourceType = "secretstore"
)

// AutoCompleteService provides auto-completion functionality for Kubernetes resources.
type AutoCompleteService struct {
	kubeLister KubeLister
	cache      map[string]cacheEntry
}

type cacheEntry struct {
	items     []string
	timestamp time.Time
}

// NewAutoCompleteService creates a new AutoCompleteService instance.
func NewAutoCompleteService(client KubeLister) *AutoCompleteService {
	return &AutoCompleteService{
		kubeLister: client,
		cache:      make(map[string]cacheEntry),
	}
}

// GetSuggestions returns suggestions for the given resource type and namespace.
func (acs *AutoCompleteService) GetSuggestions(ctx context.Context, resourceType ResourceType, namespace, query string) ([]string, error) {
	// Nil check for kubeLister
	if acs.kubeLister == nil {
		return []string{}, nil
	}
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", resourceType, namespace)
	if entry, exists := acs.cache[cacheKey]; exists && time.Since(entry.timestamp) < 30*time.Second {
		return acs.filterSuggestions(entry.items, query), nil
	}

	// Fetch fresh data
	items, err := acs.fetchResourceItems(ctx, resourceType, namespace)
	if err != nil {
		return nil, err
	}

	// Cache the results
	acs.cache[cacheKey] = cacheEntry{
		items:     items,
		timestamp: time.Now(),
	}

	return acs.filterSuggestions(items, query), nil
}

// fetchResourceItems fetches items for the given resource type and namespace.
func (acs *AutoCompleteService) fetchResourceItems(ctx context.Context, resourceType ResourceType, namespace string) ([]string, error) {
	switch resourceType {
	case ResourceTypeNamespace:
		return acs.kubeLister.GetNamespaces(ctx)
	case ResourceTypeService:
		return acs.kubeLister.GetServices(ctx, namespace)
	case ResourceTypeConfigMap:
		return acs.kubeLister.GetConfigMaps(ctx, namespace)
	case ResourceTypeSecret:
		return acs.kubeLister.GetSecrets(ctx, namespace)
	case ResourceTypePod:
		return acs.kubeLister.GetPods(ctx, namespace)
	case ResourceTypeDeployment:
		return acs.kubeLister.GetDeployments(ctx, namespace)
	case ResourceTypeStatefulSet:
		return acs.kubeLister.GetStatefulSets(ctx, namespace)
	case ResourceTypeDaemonSet:
		return acs.kubeLister.GetDaemonSets(ctx, namespace)
	case ResourceTypePersistentVolumeClaim:
		return acs.kubeLister.GetPersistentVolumeClaims(ctx, namespace)
	case ResourceTypeClusterSecretStore:
		return acs.kubeLister.GetClusterSecretStores(ctx)
	case ResourceTypeSecretStore:
		return acs.kubeLister.GetSecretStores(ctx, namespace)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// filterSuggestions filters the suggestions based on the query string.
func (acs *AutoCompleteService) filterSuggestions(items []string, query string) []string {
	if query == "" {
		return items
	}

	var filtered []string
	queryLower := strings.ToLower(query)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), queryLower) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// ClearCache clears the auto-completion cache.
func (acs *AutoCompleteService) ClearCache() {
	acs.cache = make(map[string]cacheEntry)
}

// GetNamespaceSuggestions returns namespace suggestions for auto-completion.
func (acs *AutoCompleteService) GetNamespaceSuggestions(ctx context.Context, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypeNamespace, "", query)
}

// GetServiceSuggestions returns service suggestions for auto-completion in a specific namespace.
func (acs *AutoCompleteService) GetServiceSuggestions(ctx context.Context, namespace, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypeService, namespace, query)
}

// GetConfigMapSuggestions returns configmap suggestions for auto-completion in a specific namespace.
func (acs *AutoCompleteService) GetConfigMapSuggestions(ctx context.Context, namespace, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypeConfigMap, namespace, query)
}

// GetSecretSuggestions returns secret suggestions for auto-completion in a specific namespace.
func (acs *AutoCompleteService) GetSecretSuggestions(ctx context.Context, namespace, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypeSecret, namespace, query)
}

// GetDeploymentSuggestions returns deployment suggestions for auto-completion in a specific namespace.
func (acs *AutoCompleteService) GetDeploymentSuggestions(ctx context.Context, namespace, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypeDeployment, namespace, query)
}

// GetStatefulSetSuggestions returns statefulset suggestions for auto-completion in a specific namespace.
func (acs *AutoCompleteService) GetStatefulSetSuggestions(ctx context.Context, namespace, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypeStatefulSet, namespace, query)
}

// GetDaemonSetSuggestions returns daemonset suggestions for auto-completion in a specific namespace.
func (acs *AutoCompleteService) GetDaemonSetSuggestions(ctx context.Context, namespace, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypeDaemonSet, namespace, query)
}

// GetPVCSuggestions returns PVC suggestions for auto-completion in a specific namespace.
func (acs *AutoCompleteService) GetPVCSuggestions(ctx context.Context, namespace, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypePersistentVolumeClaim, namespace, query)
}

// GetClusterSecretStoreSuggestions returns ClusterSecretStore suggestions for auto-completion.
func (acs *AutoCompleteService) GetClusterSecretStoreSuggestions(ctx context.Context, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypeClusterSecretStore, "", query)
}

// GetSecretStoreSuggestions returns SecretStore suggestions for auto-completion in a specific namespace.
func (acs *AutoCompleteService) GetSecretStoreSuggestions(ctx context.Context, namespace, query string) ([]string, error) {
	return acs.GetSuggestions(ctx, ResourceTypeSecretStore, namespace, query)
}
