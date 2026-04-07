# Kontractor

**An open standard for containers to declare their infrastructure requirements, enabling automatic provisioning and deploy-time validation in Kubernetes.**

## The Problem

Containers are opaque binaries. They need environment variables, TLS certificates, persistent storage, config files, secrets, and RBAC permissions -- but there is no standardized way for a container to *declare* those requirements. Instead, platform engineers manually author Helm values, Kustomize overlays, or raw Kubernetes manifests to wire everything together.

When a container image is upgraded, the deployment config may not match. A new env var, a changed mount path, or an additional TLS cert file silently breaks the deployment. The image and its deployment config are independently versioned artifacts with no enforced coupling.

## The Solution

**Kontractor** inverts the deployment model: the container image itself carries a machine-readable *contract* of what it expects the platform to provision. Tools read this contract and automatically validate or inject the required resources.

```
Container Image --> Contract (YAML) --> kontractor-post-render --> Mutated Manifests --> Running Pod
```

## Key Features

- **Container Contract Spec** -- A YAML schema for declaring env vars, volume mounts, secrets, configmaps, ports, probes, RBAC, and security context
- **Conditional Requirements** -- Feature flags (`tls`, `backup`, `ha`) that gate groups of requirements using `when` clauses
- **Helm Post-Renderer** -- Mutate any Helm chart's output by injecting contract requirements, with no chart modifications needed
- **CLI Validator** -- `kontractor-validate` parses and displays contract contents
- **Variable Substitution** -- `${RELEASE}` and custom variables resolved at render time
- **Dry-run Mode** -- Preview exactly what will be injected before applying

## Quick Start

```bash
# Build both tools
make build

# Validate a contract
./bin/kontractor-validate examples/echo-server.contract.yaml

# Mutate a chart (pipe through post-renderer)
helm template myrelease charts/echo-server-bare | \
  ./bin/kontractor-post-render \
    --contract=examples/echo-server.contract.yaml \
    --features=tls \
    --set-vars=RELEASE=myrelease

# Use Helm's --post-renderer flag
KONTRACTOR_CONTRACT=examples/echo-server.contract.yaml \
KONTRACTOR_FEATURES=tls \
  helm install myrelease charts/echo-server-bare \
    --post-renderer ./scripts/kontractor-post-render.sh

# Run the full demo
bash test/demo.sh
```

## Contract Example

```yaml
apiVersion: kontractor.io/v1
kind: ContainerContract
metadata:
  name: echo-server
  version: "1.0.0"
spec:
  features:
    tls:
      description: "Enable HTTPS with TLS certificates"
      default: false
  env:
    required:
      - name: APP_PORT
        default: "8080"
    conditional:
      - when: { feature: tls }
        env:
          - name: TLS_ENABLED
            value: "true"
  mounts:
    required:
      - path: /tmp
        type: ephemeral
        maxSize: 64Mi
    conditional:
      - when: { feature: tls }
        mounts:
          - path: /etc/tls
            type: secret
            secretName: "${RELEASE}-tls"
  ports:
    - name: http
      containerPort: 8080
      protocol: TCP
```

## Project Structure

```
kontractor/
├── spec/
│   └── contract-schema.yaml         # Contract JSON Schema
├── examples/
│   └── echo-server.contract.yaml    # Full example contract
├── pkg/
│   ├── contract/                    # Shared types, loader, variable substitution
│   ├── parser/                      # Multi-document Kubernetes YAML parser
│   └── mutator/                     # Mutation engine (env, volume, mount injection)
├── cmd/
│   ├── kontractor-validate/         # Phase 1: CLI validator
│   └── kontractor-post-render/      # Phase 2: Helm post-renderer
├── charts/
│   ├── echo-server/                 # Full sample chart (contract-aware)
│   └── echo-server-bare/            # Bare chart for testing post-renderer
├── scripts/
│   └── kontractor-post-render.sh    # Helm --post-renderer wrapper
├── docs/
│   ├── concept.html                 # Concept & benefits write-up
│   ├── project-architecture.html    # Architecture deep dive
│   └── helm-plugin.html             # Post-renderer report
├── test/
│   └── demo.sh                      # Full pipeline demo
├── plugin.yaml                      # Helm plugin manifest
├── Makefile                         # Builds both binaries
├── LICENSE                          # Apache 2.0
├── CONTRIBUTORS.md
└── README.md
```

## Roadmap

| Phase | Component | Status |
|-------|-----------|--------|
| 1 | CLI Validator (`kontractor-validate`) | Done |
| 2 | Helm Post-Renderer (`kontractor-post-render`) | Done |
| 3 | Contract Discovery (OCI annotations, in-chart contracts) | Planned |
| 4 | Validation Mode (verify existing charts against contracts) | Planned |
| 5 | Mutating Admission Webhook | Future |
| 6 | Contract Registry | Future |

## Contributing

See [CONTRIBUTORS.md](CONTRIBUTORS.md). This project uses the Apache 2.0 license.

## License

Apache License 2.0 -- see [LICENSE](LICENSE).
