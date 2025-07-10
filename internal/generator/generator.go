package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig holds the configuration for generating Flux manifests
type AppConfig struct {
	Name      string
	Image     string
	Port      int
	Namespace string
}

// Generator handles the creation of Flux manifests
type Generator struct {
	config AppConfig
}

// NewGenerator creates a new Generator instance
func NewGenerator(config AppConfig) *Generator {
	return &Generator{
		config: config,
	}
}

// Generate creates all necessary Flux manifests
func (g *Generator) Generate(outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// TODO: Implement manifest generation
	// This will include:
	// - Kustomization.yaml
	// - deployment.yaml
	// - service.yaml
	// - namespace.yaml (if needed)
	// - ingress.yaml (if needed)

	fmt.Printf("Generated manifests for %s in %s\n", g.config.Name, outputDir)
	return nil
}

// GenerateKustomization creates the kustomization.yaml file
func (g *Generator) GenerateKustomization(outputDir string) error {
	// TODO: Implement kustomization.yaml generation
	return nil
}

// GenerateDeployment creates the deployment.yaml file
func (g *Generator) GenerateDeployment(outputDir string) error {
	// TODO: Implement deployment.yaml generation
	return nil
}

// GenerateService creates the service.yaml file
func (g *Generator) GenerateService(outputDir string) error {
	// TODO: Implement service.yaml generation
	return nil
} 