# Flux App Generator

A modern, interactive CLI tool to generate Flux GitOps manifests for Helm-based applications. This tool simplifies the process of setting up Flux resources by providing an intuitive terminal UI that guides you through the configuration process.

## ✨ Features

- **Interactive Terminal UI** - Beautiful, user-friendly interface built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Helm Repository Integration** - Automatically fetches available charts and versions from Helm repositories
- **Smart Chart Selection** - Browse and select charts with descriptions and version information
- **Flux v2 Resource Generation** - Creates all necessary Flux resources:
  - `dependencies/helm-repository.yaml` - HelmRepository resource
  - `release/helm-release.yaml` - HelmRelease resource  
  - `release/helm-values.yaml` - Helm values configuration
  - `kustomization.yaml` - Kustomize configuration
- **Embedded Templates** - Uses Go's embed functionality for reliable template distribution
- **Comprehensive Testing** - High test coverage with mocked network calls for CI reliability

## 🚀 Quick Start

### Prerequisites
- Go 1.24.0 or later
- A Helm repository URL (e.g., `https://helm.airbyte.io`)

### Installation & Usage

```bash
# Build the application
make build

# Run the interactive generator
make run
```

The CLI will guide you through:
1. **Application Name** - Name for your Flux application
2. **Namespace** - Kubernetes namespace (defaults to "default")
3. **Helm Repository** - Name and URL for the Helm repository
4. **Chart Selection** - Browse and select from available charts
5. **Version Selection** - Choose the chart version to deploy
6. **Sync Interval** - Flux sync interval (1m, 5m, 10m, 30m, 1h)

## 📁 Project Structure

```
flux-app-generator/
├── cmd/
│   └── flux-app-generator/
│       ├── main.go                    # CLI entrypoint with Bubble Tea UI
│       └── templates/                 # Embedded YAML templates
│           ├── helm-repository.yaml.tmpl
│           ├── helm-release.yaml.tmpl
│           ├── helm-values.yaml.tmpl
│           └── kustomization.yaml.tmpl
├── internal/
│   ├── generator/
│   │   ├── generator.go               # Flux resource generation logic
│   │   └── generator_test.go          # Comprehensive tests
│   ├── helm/
│   │   ├── version_fetcher.go         # Helm repository integration
│   │   └── version_fetcher_test.go    # Mocked network tests
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
- **Mocked Network Calls** - Helm repository tests use mocks for reliability
- **Template Testing** - Generator tests verify YAML output
- **CI Integration** - Tests run automatically on every PR

```bash
# Run tests with coverage
go test -v -coverprofile=coverage.txt -covermode=atomic ./...

# View coverage report
go tool cover -func=coverage.txt
```

### Code Quality

- **golangci-lint** - Comprehensive linting with multiple linters
- **Go 1.24.5** - Latest stable Go version
- **Embedded Templates** - No external file dependencies
- **Error Handling** - Robust error handling throughout

## 📋 Generated Resources

The tool generates a complete Flux GitOps structure:

```
your-app/
├── dependencies/
│   └── helm-repository.yaml    # Flux HelmRepository
├── release/
│   ├── helm-release.yaml       # Flux HelmRelease
│   └── helm-values.yaml        # Helm values
└── kustomization.yaml          # Kustomize configuration
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
  url: https://helm.example.com
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

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

### Development Guidelines

- Follow Go best practices and conventions
- Add tests for new features
- Use mocks for network-dependent code
- Update documentation as needed
- Ensure CI checks pass

## 📄 License

MIT License - see LICENSE file for details.

---

**Built with ❤️ using [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the terminal UI and [Flux](https://fluxcd.io/) for GitOps.** 