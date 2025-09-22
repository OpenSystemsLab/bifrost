# Bifrost - OpenSystemsLab Fork

This is a fork of [maximhq/bifrost](https://github.com/maximhq/bifrost) maintained by OpenSystemsLab for our Open WebUI GitOps deployment.

## What is Bifrost?

Bifrost is "The Fastest LLM Gateway" with built-in OpenTelemetry observability and MCP (Model Context Protocol) gateway capabilities.

### Key Features:
- 🚀 High-performance LLM Gateway written in Go
- 🌐 Next.js Web UI for management and monitoring  
- 🔌 MCP (Model Context Protocol) integration
- 📊 OpenTelemetry observability and metrics
- 🐳 Multi-architecture Docker images (amd64, arm64)
- 🔧 Flexible configuration and plugin system

## Our Modifications

- ✅ Custom Docker image builds via GitHub Actions
- ✅ Integration with our EKS cluster and GitOps workflow
- ✅ IRSA (IAM Roles for Service Accounts) integration for AWS services
- ✅ Tailscale networking support
- ✅ Proper understanding of Makefile build process

## Docker Images

Our custom builds are available at:
```
ghcr.io/opensystemslab/bifrost:latest
ghcr.io/opensystemslab/bifrost:sha-<commit>
```

## Usage in Open WebUI GitOps

This fork is specifically configured for our Open WebUI deployment with:
- AWS Secrets Manager integration for API keys
- Kubernetes service discovery
- MCP server routing to internal services
- Enterprise security practices
- Zero local secrets architecture

### Deployment Configuration

```yaml
image: ghcr.io/opensystemslab/bifrost:latest
ports:
  - containerPort: 8080
env:
  - name: APP_PORT
    value: "8080"
  - name: LOG_LEVEL
    value: "info"
  - name: LOG_STYLE
    value: "json"
```

## Development

To build locally:
```bash
# Install dependencies
make install-ui

# Build UI and binary
make build

# Run with hot reload
make dev

# Build Docker image
make docker-build
```

## Original Project

All credit goes to the original [maximhq/bifrost](https://github.com/maximhq/bifrost) project.
We maintain this fork to ensure stability and add specific integrations for our use case.

## Sync Status

This fork is regularly synchronized with the upstream repository to incorporate
latest features and security updates.

Last sync: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
