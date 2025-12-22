#!/bin/bash
# Build release binaries for afk
# Usage: ./build_releases.sh [version]
# Example: ./build_releases.sh 1.0.0

set -e

VERSION="${1:-dev}"
OUTPUT_DIR="releases"
BINARY_NAME="afk"

echo "Building afk ${VERSION}..."
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Platforms to build
PLATFORMS=(
    "darwin/arm64"
    "darwin/amd64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
)

# Build each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"
    OUTPUT_NAME="${BINARY_NAME}-${GOOS}-${GOARCH}"

    # Add .exe for Windows
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    echo "Building ${OUTPUT_NAME}..."

    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.version=${VERSION}" \
        -o "${OUTPUT_DIR}/${OUTPUT_NAME}" \
        ./cmd/main.go
done

echo ""
echo "Build complete!"
echo ""
ls -lh "$OUTPUT_DIR"
echo ""
echo "To create a GitHub release:"
echo "  1. git tag v${VERSION}"
echo "  2. git push origin v${VERSION}"
echo "  3. Go to https://github.com/packetpipe/afk/releases"
echo "  4. Create new release from tag v${VERSION}"
echo "  5. Upload the binaries from ${OUTPUT_DIR}/"
