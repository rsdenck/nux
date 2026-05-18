#!/bin/bash
set -e

APP_NAME="nux"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR="bin"

echo "Building $APP_NAME version $VERSION..."

platforms=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm/v7"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH GOARM <<< "$platform"
    
    output_name=$APP_NAME'_'$GOOS'_'$GOARCH
    if [ "$GOARCH" = "arm" ]; then
        output_name=$output_name'v'$GOARM
    fi

    echo "Building for $GOOS/$GOARCH..."
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    export CGO_ENABLED=0
    if [ "$GOARCH" = "arm" ]; then
        export GOARM=$GOARM
    fi

    go build -ldflags "-s -w -X main.version=$VERSION" -trimpath -o "$BUILD_DIR/$output_name" ./cmd/nux
    
    if [ $? -eq 0 ]; then
        echo "Built $output_name"
    else
        echo "Failed to build for $GOOS/$GOARCH"
        exit 1
    fi
done

echo "Build complete! Artifacts in $BUILD_DIR/"
ls -lh $BUILD_DIR/
