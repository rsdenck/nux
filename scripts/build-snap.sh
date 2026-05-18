#!/bin/bash
# scripts/build-snap.sh
# Build Snap package for NUX

set -e

echo "Building Snap package for NUX..."

# Check for snapcraft
if ! command -v snapcraft >/dev/null 2>&1; then
    echo "Error: snapcraft not found. Please install snapcraft."
    exit 1
fi

# Build snap
snapcraft clean
snapcraft

echo "Snap package built successfully."
