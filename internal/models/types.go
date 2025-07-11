// Package models provides data structure definitions for Flux application configuration and management.
package models

import (
	"github.com/EffectiveSloth/flux-app-generator/internal/plugins"
)

// AppConfig represents the complete configuration for generating a Flux application.
type AppConfig struct {
	AppName      string
	Namespace    string
	HelmRepoName string
	HelmRepoURL  string
	ChartName    string
	ChartVersion string
	Interval     string
	Values       map[string]interface{}
	Plugins      []plugins.PluginConfig
	PluginFiles  []string // Relative paths to plugin-generated files
}
