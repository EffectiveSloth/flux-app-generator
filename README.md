# Flux App Generator

A modern, extensible CLI tool to generate Flux GitOps manifests for Helm-based applications with plugin support. This tool simplifies the process of setting up Flux resources by providing an intuitive terminal UI and a powerful plugin architecture for additional integrations.

## ✨ Features

- **Interactive Terminal UI** - Beautiful, user-friendly interface built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Plugin Architecture** - Extensible plugin system for additional integrations (External Secrets, etc.)
- **Helm Repository Integration** - Automatically fetches available charts and versions from Helm repositories
- **Smart Chart Selection** - Browse and select charts with descriptions and version information
- **Flux v2 Resource Generation** - Creates all necessary Flux resources:
  - `dependencies/helm-repository.yaml` - HelmRepository resource
  - `release/helm-release.yaml` - HelmRelease resource  
  - `release/helm-values.yaml` - Helm values configuration
  - `kustomization.yaml` - Kustomize configuration
- **Plugin-Generated Resources** - Additional resources based on configured plugins
- **Values Prefilling** - Option to download default values from Helm charts
- **Embedded Templates** - Uses Go's embed functionality for reliable template distribution
- **Comprehensive Testing** - High test coverage with mocked network calls for CI reliability

## 🔌 Plugin System

The tool now features a powerful plugin architecture that allows extending functionality:

### Available Plugins

- **ExternalSecret Plugin** - Generates External Secrets Operator resources for managing secrets from external secret stores
  - Supports ClusterSecretStore and SecretStore references
  - Configurable refresh intervals
  - Automatic secret creation and management

### Plugin Features

- **Interactive Configuration** - Each plugin provides its own configuration interface
- **Template-Based Generation** - Plugins use Go templates for resource generation
- **Validation** - Built-in validation for plugin configurations
- **Multiple Instances** - Support for multiple instances of the same plugin type

## 🚀 Quick Start

### Prerequisites
- Go 1.24.0 or later
- A Helm repository URL (e.g., `https://helm.datadoghq.com`)

### Installation & Usage

```bash
# Build the application
make build

# Run the interactive generator
make run
```

The CLI will guide you through:
1. **Application Configuration** - Name, namespace, and Helm repository details
2. **Chart Selection** - Browse and select from available charts
3. **Version Selection** - Choose the chart version to deploy
4. **Configuration** - Set sync interval and values prefill options
5. **Plugin Management** - Configure optional plugins for additional functionality

## 📁 Project Structure

```
flux-app-generator/
├── cmd/
│   └── flux-app-generator/
│       ├── main.go                    # CLI entrypoint with Bubble Tea UI
│       └── templates/                 # Embedded YAML templates
│           ├── helm-repository.yaml.tmpl
│           ├── helm-release.yaml.tmpl
│           └── kustomization.yaml.tmpl
├── internal/
│   ├── generator/
│   │   ├── generator.go               # Flux resource generation logic
│   │   └── generator_test.go          # Comprehensive tests
│   ├── helm/
│   │   ├── version_fetcher.go         # Helm repository integration
│   │   ├── version_fetcher_test.go    # Mocked network tests
│   │   ├── chart_downloader.go        # Chart downloading functionality
│   │   └── chart_downloader_test.go   # Chart downloader tests
│   ├── plugins/                       # Plugin system
│   │   ├── types.go                   # Plugin interfaces and types
│   │   ├── types_test.go              # Plugin type tests
│   │   ├── registry.go                # Plugin registry management
│   │   ├── registry_test.go           # Registry tests
│   │   ├── externalsecret.go          # External Secrets plugin
│   │   └── externalsecret_test.go     # External Secrets tests
│   └── types/
│       ├── types.go                   # Application configuration types
│       └── types_test.go              # Type validation tests
├── .github/workflows/                 # CI/CD pipelines
│   ├── build.yml                      # Build and release
│   ├── test.yml                       # Test with coverage
│   └── lint.yml                       # Code quality checks
├── Makefile                           # Build and development tasks
├── .golangci.yml                      # Linting configuration
└── README.md
```

## 🛠️ Development

### Available Commands

```bash
make build    # Build the CLI binary in dist/
make run      # Build and run the CLI from dist/
make test     # Run all tests with coverage
make lint     # Lint the codebase with golangci-lint
make clean    # Remove build artifacts in dist/
make help     # Show available commands
```

### Testing

The project includes comprehensive tests with high coverage:

- **Unit Tests** - All packages have thorough unit tests
- **Plugin Tests** - Comprehensive testing of the plugin system
- **Mocked Network Calls** - Helm repository tests use mocks for reliability
- **Template Testing** - Generator tests verify YAML output
- **CI Integration** - Tests run automatically on every PR

```bash
# Run tests with coverage
go test -v -coverprofile=coverage.txt -covermode=atomic ./...

# View coverage report
go tool cover -func=coverage.txt
```

### Adding New Plugins

To create a new plugin:

1. Implement the `Plugin` interface in `internal/plugins/`
2. Define your plugin variables, template, and file path
3. Register the plugin in the registry
4. Add comprehensive tests

Example plugin structure:
```go
type MyPlugin struct {
    BasePlugin
}

func NewMyPlugin() *MyPlugin {
    return &MyPlugin{
        BasePlugin: BasePlugin{
            name:        "myplugin",
            description: "Description of what this plugin does",
            variables:   []Variable{...},
            template:    "...",
            filePath:    "path/to/output.yaml",
        },
    }
}
```

### Code Quality

- **golangci-lint** - Comprehensive linting with multiple linters
- **Go 1.24.5** - Latest stable Go version
- **Embedded Templates** - No external file dependencies
- **Error Handling** - Robust error handling throughout

## 📋 Generated Resources

The tool generates a complete Flux GitOps structure with optional plugin resources:

```
your-app/
├── dependencies/
│   ├── helm-repository.yaml           # Flux HelmRepository
│   └── external-secret-*.yaml         # External Secrets (if configured)
├── release/
│   ├── helm-release.yaml              # Flux HelmRelease
│   └── helm-values.yaml               # Helm values
└── kustomization.yaml                 # Kustomize configuration
```

### Example Output

**helm-repository.yaml:**
```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: my-repo
  namespace: default
spec:
  interval: 5m
  url: https://helm.datadoghq.com
```

**helm-release.yaml:**
```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: my-app
  namespace: default
spec:
  interval: 5m
  chart:
    spec:
      chart: my-chart
      version: '1.0.0'
      sourceRef:
        kind: HelmRepository
        name: my-repo
      interval: 5m
  valuesFrom:
    - kind: ConfigMap
      name: my-app-values
      valuesKey: values.yaml
```

**external-secret-example.yaml** (if External Secrets plugin configured):
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: my-secret
  namespace: default
spec:
  secretStoreRef:
    kind: ClusterSecretStore
    name: vault-backend
  dataFrom:
    - extract:
        key: myapp/secrets
  refreshInterval: 60m
  target:
    creationPolicy: Owner
    name: my-app-secrets
```

## 🔧 Plugin Configuration

The plugin system allows for flexible configuration of additional resources:

### External Secrets Plugin

Configure external secret management with the following options:
- **Secret Store Type**: ClusterSecretStore or SecretStore
- **Secret Store Name**: Name of the secret store resource
- **Secret Key**: Key name in the external secret store
- **Target Secret**: Name of the Kubernetes secret to create
- **Refresh Interval**: How often to refresh the secret (15m to 24h)

### Multiple Plugin Instances

You can configure multiple instances of the same plugin type for different secrets or configurations.

## 🚀 Releases

This project uses automated releases with [Release Please](https://github.com/googleapis/release-please) based on [Conventional Commits](https://www.conventionalcommits.org/).

### How Releases Work

- **Automatic Changelog Generation** - Changes are automatically categorized and documented
- **Semantic Versioning** - Version numbers follow [SemVer](https://semver.org/) based on commit types
- **Release PRs** - Release Please creates and maintains release pull requests
- **Cross-Platform Binaries** - Releases include compiled binaries for Linux, macOS, and Windows

### Commit Message Format

Use conventional commit messages to trigger appropriate version bumps:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types that trigger releases:**
- `feat:` - New feature (minor version bump)
- `fix:` - Bug fix (patch version bump)
- `feat!:` or `fix!:` - Breaking change (major version bump)

**Other useful types:**
- `docs:` - Documentation changes
- `chore:` - Maintenance tasks
- `test:` - Test improvements
- `refactor:` - Code refactoring

**Examples:**
```bash
feat: add support for Kubernetes secrets plugin
fix: resolve chart version fetching timeout
feat!: change plugin configuration API (breaking change)
docs: update installation instructions
chore: bump dependencies to latest versions
```

### Release Process

1. **Development** - Make changes using conventional commit messages
2. **Release PR** - Release Please automatically creates/updates a release PR
3. **Review & Merge** - Review the generated changelog and merge the release PR
4. **Automated Release** - GitHub Actions automatically:
   - Creates a GitHub release with changelog
   - Builds and attaches cross-platform binaries
   - Generates checksums for security verification

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with conventional commit messages
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Development Guidelines

- Follow Go best practices and conventions
- Use conventional commit messages for automatic release management
- Add tests for new features and plugins
- Use mocks for network-dependent code
- Update documentation as needed
- Ensure CI checks pass

## 📄 License

GNU General Public License v3.0 - see LICENSE file for details.

---

**Built with ❤️ using [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the terminal UI, [Flux](https://fluxcd.io/) for GitOps, and an extensible plugin architecture for enhanced functionality.** 