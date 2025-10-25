#!/bin/bash

# Build release binaries for multiple platforms

set -e

APP_NAME="whatsapp-web-api"
VERSION=${1:-"latest"}
BUILD_DIR="releases"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Clean and create releases directory
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

print_status "Building release binaries..."

# Build for different platforms
platforms=(
    "linux/amd64"
    "linux/arm64"
    "linux/386"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/386"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"

    # Set Windows executable extension
    if [[ $GOOS == "windows" ]]; then
        OUTPUT_NAME="${APP_NAME}-${GOOS}-${GOARCH}.exe"
    else
        OUTPUT_NAME="${APP_NAME}-${GOOS}-${GOARCH}"
    fi

    print_status "Building for $GOOS/$GOARCH..."

    GOOS=$GOOS GOARCH=$GOARCH go build -o "$BUILD_DIR/$OUTPUT_NAME" .

    # Create zip archive
    cd $BUILD_DIR
    zip "${OUTPUT_NAME}.zip" "$OUTPUT_NAME"
    rm "$OUTPUT_NAME"
    cd ..

    print_success "Created $BUILD_DIR/${OUTPUT_NAME}.zip"
done

# Copy documentation files
print_status "Copying documentation..."
cp README.md $BUILD_DIR/
cp swagger.yaml $BUILD_DIR/
cp .env.example $BUILD_DIR/

# Create checksums
print_status "Generating checksums..."
cd $BUILD_DIR
sha256sum *.zip > checksums.txt
cd ..

print_success "Release build completed!"
print_status "Files in $BUILD_DIR:"
ls -la $BUILD_DIR/