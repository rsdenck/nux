#!/bin/bash
# scripts/build-aur.sh
# Build AUR package for NUX

set -e

VERSION="0.3.0"
AUR_DIR="dist/aur"

echo "Building AUR package for NUX v${VERSION}..."

# Clean previous build
rm -rf "$AUR_DIR"
mkdir -p "$AUR_DIR"

# Copy PKGBUILD
cp packaging/aur/PKGBUILD "$AUR_DIR/"
cp LICENSE "$AUR_DIR/"
cp README.md "$AUR_DIR/"

# Update checksums (requires makepkg)
if command -v updpkgsums >/dev/null 2>&1; then
    cd "$AUR_DIR"
    echo "Updating checksums..."
    updpkgsums
else
    echo "Warning: updpkgsums not found. Checksums in PKGBUILD might be outdated."
fi

# Build package (requires makepkg)
if command -v makepkg >/dev/null 2>&1; then
    cd "$AUR_DIR"
    echo "Running makepkg..."
    makepkg -sfc
    echo "AUR package built successfully at $AUR_DIR"
else
    echo "Warning: makepkg not found. Skipping build."
    echo "PKGBUILD is ready at $AUR_DIR/PKGBUILD"
fi
