#!/bin/bash

# Local build script to test release builds
# This script mimics what the GitHub Actions release workflow does

set -e

VERSION=${1:-"dev"}
COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "Building sortTF version $VERSION"
echo "Commit: $COMMIT"
echo "Date: $DATE"

# Build for different platforms
platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

mkdir -p dist

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    
    output="dist/sorttf-${GOOS}-${GOARCH}"
    if [[ $GOOS == "windows" ]]; then
        output+=".exe"
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
        -ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
        -o "$output" .
done

echo ""
echo "Build complete! Binaries created in dist/ directory:"
ls -la dist/

echo ""
echo "Creating checksums..."
cd dist
sha256sum sorttf-* > checksums.txt
echo "Checksums:"
cat checksums.txt