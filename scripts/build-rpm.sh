#!/bin/bash
# scripts/build-rpm.sh
# Build RPM package for NUX

set -e

VERSION="0.3.0"
RPM_ROOT="dist/rpm"

echo "Building RPM package for NUX v${VERSION}..."

# Clean previous build
rm -rf "$RPM_ROOT"
mkdir -p "$RPM_ROOT"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

# Create source tarball
echo "Creating source tarball..."
tar --exclude='.git' --exclude='dist' -czf "$RPM_ROOT/SOURCES/nux-${VERSION}.tar.gz" .

# Copy spec file
cp packaging/rpm/nux.spec "$RPM_ROOT/SPECS/"

# Build RPM
echo "Running rpmbuild..."
rpmbuild --define "_topdir $(pwd)/$RPM_ROOT" --nodeps -ba "$RPM_ROOT/SPECS/nux.spec"

echo "RPM package built successfully at $RPM_ROOT/RPMS/"
