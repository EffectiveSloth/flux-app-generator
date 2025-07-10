package types

import (
	"github.com/EffectiveSloth/flux-app-generator/internal/plugins"
)

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
