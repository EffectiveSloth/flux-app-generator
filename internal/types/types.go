package types

type AppConfig struct {
	AppName        string
	Namespace      string
	HelmRepoName   string
	HelmRepoURL    string
	ChartName      string
	ChartVersion   string
	Interval       string
	Values         map[string]interface{}
} 