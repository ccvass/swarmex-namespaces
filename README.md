<p align="center"><img src="https://raw.githubusercontent.com/ccvass/swarmex/main/docs/assets/logo.svg" alt="Swarmex" width="400"></p>

[![Test, Build & Deploy](https://github.com/ccvass/swarmex-namespaces/actions/workflows/publish.yml/badge.svg)](https://github.com/ccvass/swarmex-namespaces/actions/workflows/publish.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

# Swarmex Namespaces

Namespace isolation via overlay networks for Docker Swarm services.

Part of [Swarmex](https://github.com/ccvass/swarmex) — enterprise-grade orchestration for Docker Swarm.

## What It Does

Creates isolated overlay networks per namespace label, providing network-level isolation between groups of services. Each namespace gets its own dedicated network, and services are automatically attached to their namespace network.

## Labels

```yaml
deploy:
  labels:
    swarmex.namespace: "frontend"  # Namespace to assign this service to
```

## How It Works

1. Watches for services with the `swarmex.namespace` label.
2. Creates an overlay network named `ns-<namespace>` if it doesn't exist.
3. Attaches the service to its namespace network.
4. Services in different namespaces are network-isolated by default.

## Quick Start

```bash
docker service update \
  --label-add swarmex.namespace=frontend \
  my-app
```

## Verified

Namespace networks `ns-frontend`, `ns-backend`, and `ns-production` created and services correctly isolated.

## License

Apache-2.0
