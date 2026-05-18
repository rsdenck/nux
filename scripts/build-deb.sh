#!/bin/bash
# scripts/build-deb.sh
# Build Debian package for NUX

set -e

VERSION="0.3.0"
ARCH="amd64"
PKG_DIR="packaging/deb"
BUILD_DIR="dist/deb/nux_${VERSION}_${ARCH}"

echo "Building Debian package for NUX v${VERSION}..."

# Clean previous build
rm -rf "dist/deb"
mkdir -p "$BUILD_DIR/usr/local/bin"
mkdir -p "$BUILD_DIR/usr/local/share/nux/man"
mkdir -p "$BUILD_DIR/DEBIAN"

# Compile binary
echo "Compiling binary..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$VERSION" -o "$BUILD_DIR/usr/local/bin/nux" cmd/nux/main.go

# Copy man page
echo "Copying documentation..."
cp docs/man/nux.1 "$BUILD_DIR/usr/local/share/nux/man/"
gzip "$BUILD_DIR/usr/local/share/nux/man/nux.1"

# Copy control files
echo "Copying control files..."
cp "$PKG_DIR/DEBIAN/control" "$BUILD_DIR/DEBIAN/"
cp "$PKG_DIR/DEBIAN/postinst" "$BUILD_DIR/DEBIAN/"
cp "$PKG_DIR/DEBIAN/prerm" "$BUILD_DIR/DEBIAN/"
chmod 755 "$BUILD_DIR/DEBIAN/postinst"
chmod 755 "$BUILD_DIR/DEBIAN/prerm"

# Build package
echo "Creating .deb package..."
dpkg-deb --build "$BUILD_DIR" "dist/deb/nux_${VERSION}_${ARCH}.deb"

echo "Debian package built successfully at dist/deb/nux_${VERSION}_${ARCH}.deb"
