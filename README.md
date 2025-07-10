# Flux App Generator

A CLI tool to interactively generate the necessary YAML files in a Flux Git repository to deploy an app using Helm and Kustomize.

## Features
- Interactive CLI for Helm/Flux deployment setup
- Generates:
  - `dependencies/helm-repository.yaml`
  - `release/helm-release.yaml`
  - `release/helm-values.yaml`
  - `kustomization.yaml`
- Uses Go templates for easy customization

## Project Structure
```
flux-app-generator/
├── cmd/
│   └── flux-app-generator/
│       └── main.go         # CLI entrypoint
├── internal/
│   ├── generator/
│   │   └── generator.go    # Generation logic
│   └── types/
│       └── types.go        # Types (AppConfig)
├── templates/              # YAML templates
├── Makefile                # Build/test/lint tasks
├── go.mod, go.sum          # Go modules
└── README.md
```

## Usage
```sh
make build
./flux-app-generator
```

## Development
- `make run` — build and run the CLI
- `make test` — run tests
- `make lint` — lint the codebase (requires `golangci-lint`)
- `make clean` — remove build artifacts

## Customizing Templates
Edit files in `templates/` to change the generated YAML structure.

---
MIT License 