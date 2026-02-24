#!/bin/bash
# Build NanoClaw images: nanoclaw (main app) + nanoclaw-agent (agent container)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TAG="$(date +%Y%m%d-%H%M%S)"
RUNTIME="${CONTAINER_RUNTIME:-docker}"

echo "=== Building nanoclaw-agent ==="
echo "  Tags: ${TAG}, latest"
${RUNTIME} build \
  -t "nanoclaw-agent:${TAG}" \
  -t "nanoclaw-agent:latest" \
  "$SCRIPT_DIR/container"

echo ""
echo "=== Building nanoclaw ==="
echo "  Tags: ${TAG}, latest"
${RUNTIME} build \
  -t "nanoclaw:${TAG}" \
  -t "nanoclaw:latest" \
  "$SCRIPT_DIR"

echo ""
echo "Build complete!"
echo "  nanoclaw-agent:${TAG}"
echo "  nanoclaw-agent:latest"
echo "  nanoclaw:${TAG}"
echo "  nanoclaw:latest"
