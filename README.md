# Flux App Generator

A CLI tool to generate the necessary YAML files in a Flux Git repository to deploy applications using GitOps principles.

## Features

- Generate Flux-compatible Kubernetes manifests
- Support for deployments, services, and kustomizations
- Configurable application settings
- GitOps-ready output

## Installation

```bash
go install github.com/EffectiveSloth/flux-app-generator@latest
```

## Usage

### Basic Usage

```bash
# Generate manifests for a simple app
flux-app-generator generate --name my-app --image nginx:latest

# Generate with custom settings
flux-app-generator generate \
  --name my-app \
  --image my-registry/my-app:v1.0.0 \
  --port 3000 \
  --namespace production \
  --output ./manifests
```

### Global Flags

- `--config`: Config file path (default: `$HOME/.flux-app-generator.yaml`)
- `--verbose, -v`: Enable verbose output

### Generate Command Flags

- `--name, -n`: Application name (required)
- `--image, -i`: Container image (required)
- `--port, -p`: Application port (default: 8080)
- `--namespace, -s`: Kubernetes namespace (default: default)
- `--output, -o`: Output directory (default: current directory)

## Development

### Prerequisites

- Go 1.21 or later

### Building

```bash
go build -o flux-app-generator
```

### Running Tests

```bash
go test ./...
```

## License

MIT 