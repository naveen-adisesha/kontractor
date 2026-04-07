#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BINARY="$ROOT/bin/kontractor-post-render"
CONTRACT="$ROOT/examples/echo-server.contract.yaml"
CHART="$ROOT/charts/echo-server-bare"
RELEASE="kontractor-demo"

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${CYAN}=====================================================${NC}"
echo -e "${CYAN}  Kontractor Helm Post-Renderer - Live Demo${NC}"
echo -e "${CYAN}=====================================================${NC}"

if [ ! -x "$BINARY" ]; then
    echo -e "${YELLOW}Building binary...${NC}"
    make -C "$ROOT" build
fi

echo ""
echo -e "${YELLOW}┌─────────────────────────────────────────────────┐${NC}"
echo -e "${YELLOW}│  STEP 1: Bare chart (BEFORE mutation)           │${NC}"
echo -e "${YELLOW}└─────────────────────────────────────────────────┘${NC}"
echo ""
helm template "$RELEASE" "$CHART"

echo ""
echo -e "${YELLOW}┌─────────────────────────────────────────────────┐${NC}"
echo -e "${YELLOW}│  STEP 2: Dry-run - what WOULD be injected       │${NC}"
echo -e "${YELLOW}└─────────────────────────────────────────────────┘${NC}"
echo ""
helm template "$RELEASE" "$CHART" | "$BINARY" \
    --contract="$CONTRACT" \
    --set-vars="RELEASE=$RELEASE" \
    --dry-run 2>&1 || true

echo ""
echo -e "${YELLOW}┌─────────────────────────────────────────────────┐${NC}"
echo -e "${YELLOW}│  STEP 3: Mutated output (no TLS)                │${NC}"
echo -e "${YELLOW}└─────────────────────────────────────────────────┘${NC}"
echo ""
helm template "$RELEASE" "$CHART" | "$BINARY" \
    --contract="$CONTRACT" \
    --set-vars="RELEASE=$RELEASE" \
    --quiet

echo ""
echo -e "${YELLOW}┌─────────────────────────────────────────────────┐${NC}"
echo -e "${YELLOW}│  STEP 4: Mutated output (TLS enabled)           │${NC}"
echo -e "${YELLOW}└─────────────────────────────────────────────────┘${NC}"
echo ""
helm template "$RELEASE" "$CHART" | "$BINARY" \
    --contract="$CONTRACT" \
    --features=tls \
    --set-vars="RELEASE=$RELEASE" \
    --quiet

echo ""
echo -e "${YELLOW}┌─────────────────────────────────────────────────┐${NC}"
echo -e "${YELLOW}│  STEP 5: Using Helm --post-renderer flag        │${NC}"
echo -e "${YELLOW}└─────────────────────────────────────────────────┘${NC}"
echo ""
KONTRACTOR_CONTRACT="$CONTRACT" KONTRACTOR_FEATURES="tls" \
    helm template "$RELEASE" "$CHART" \
    --post-renderer "$ROOT/scripts/kontractor-post-render.sh" 2>&1

echo ""
echo -e "${GREEN}=====================================================${NC}"
echo -e "${GREEN}  Demo complete!${NC}"
echo -e "${GREEN}=====================================================${NC}"
