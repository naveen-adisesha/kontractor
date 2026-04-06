# Kontractor

**An open standard for containers to declare their infrastructure requirements, enabling automatic provisioning and deploy-time validation in Kubernetes.**

## The Problem

Containers are opaque binaries. They need environment variables, TLS certificates, persistent storage, config files, secrets, and RBAC permissions -- but there is no standardized way for a container to *declare* those requirements. Instead, platform engineers manually author Helm values, Kustomize overlays, or raw Kubernetes manifests to wire everything together.

When a container image is upgraded, the deployment config may not match. A new env var, a changed mount path, or an additional TLS cert file silently breaks the deployment. The image and its deployment config are independently versioned artifacts with no enforced coupling.

## The Solution

**Kontractor** inverts the deployment model: the container image itself carries a machine-readable *contract* of what it expects the platform to provision. A controller or admission webhook reads this contract at deploy time and automatically validates or provisions the required resources.

```
Container Image --> Contract (embedded/OCI artifact) --> Webhook --> Auto-validate/provision --> Running Pod
```

## Key Features

- **Container Contract Spec** -- A YAML schema for declaring env vars, volume mounts, secrets, configmaps, ports, probes, RBAC, and security context
- **Conditional Requirements** -- Feature flags (`tls`, `backup`, `ha`) that gate groups of requirements using `when` clauses
- **Deploy-time Validation** -- Mutating admission webhook that checks contracts against actual pod specs
- **Self-documenting Images** -- `kontractor inspect <image>` shows exactly what an image needs
- **Helm Integration** -- Contract-aware Helm library chart that auto-generates volumes, env, and init containers from contracts

## Quick Start

```bash
# Inspect a container's contract
kontractor inspect myregistry/echo-server:1.0.0

# Validate a pod spec against its contract
kontractor validate -f pod.yaml

# Deploy the sample echo-server with Kontractor
helm install echo-server charts/echo-server -n kontractor --create-namespace
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
            keys: ["tls.crt", "tls.key"]
  ports:
    - name: http
      containerPort: 8080
      protocol: TCP
  health:
    readiness:
      httpGet:
        path: /healthz
        port: 8080
```

## Project Structure

```
kontractor/
├── README.md
├── LICENSE                          # Apache 2.0
├── CONTRIBUTORS.md
├── spec/
│   └── contract-schema.yaml        # Contract JSON Schema
├── examples/
│   └── echo-server.contract.yaml   # Full example contract
├── charts/
│   └── echo-server/                # Sample Helm chart
├── cmd/
│   └── kontractor-validate/        # CLI validator (Go)
├── deploy/
│   └── webhook/                    # Admission webhook Helm chart
└── docs/
    └── concept.html                # Concept & benefits write-up
```

## Contributing

See [CONTRIBUTORS.md](CONTRIBUTORS.md). This project uses the Apache 2.0 license.

## License

Apache License 2.0 -- see [LICENSE](LICENSE).
