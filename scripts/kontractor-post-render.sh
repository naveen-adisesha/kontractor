#!/bin/bash
# Wrapper script for Helm --post-renderer
# Usage: helm install myrelease ./mychart --post-renderer ./scripts/kontractor-post-render.sh
#
# Configure via environment variables:
#   KONTRACTOR_CONTRACT  - path to contract YAML
#   KONTRACTOR_FEATURES  - comma-separated feature flags

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BINARY="${ROOT_DIR}/bin/kontractor-post-render"

if [ ! -x "$BINARY" ]; then
    echo "Error: kontractor-post-render binary not found. Run 'make build' first." >&2
    exit 1
fi

exec "$BINARY" --contract="${KONTRACTOR_CONTRACT}" --features="${KONTRACTOR_FEATURES}" "$@"
