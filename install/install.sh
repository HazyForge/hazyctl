#!/bin/bash

set -e

REPO="hazyforge/hazyctl"
BINARY_NAME="hazyctl"
INSTALL_DIR="$HOME/.local/bin"

# Determine OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Normalize architecture names
case $ARCH in
  x86_64)
    ARCH="amd64"
    ;;
  aarch64)
    ARCH="arm64"
    ;;
esac

# Map OS names to GitHub release names
case $OS in
  darwin)
    OS="darwin"
    ;;
  linux)
    OS="linux"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Check if architecture is supported
if [[ "$ARCH" != "amd64" && "$ARCH" != "arm64" ]]; then
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

# Get latest release tag
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=${LATEST_TAG#v}
FILENAME="hazyctl_${VERSION}_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$FILENAME"

if [[ "$OS" != "windows" ]]; then
  ARCHIVE_URL="${DOWNLOAD_URL}.tar.gz"
else
  ARCHIVE_URL="${DOWNLOAD_URL}.zip"
fi

SHA256_URL="${ARCHIVE_URL}.sha256"

echo "Downloading hazyctl $LATEST_TAG for $OS/$ARCH..."

# Download the archive
TMP_ARCHIVE=$(mktemp)
curl -sSL --fail "$ARCHIVE_URL" -o "$TMP_ARCHIVE"

# Verify the SHA256 checksum
echo "Verifying checksum..."
TMP_SHA256=$(mktemp)
curl -sSL --fail "$SHA256_URL" -o "$TMP_SHA256"
EXPECTED_CHECKSUM=$(cat "$TMP_SHA256" | awk '{print $1}')
ACTUAL_CHECKSUM=$(sha256sum "$TMP_ARCHIVE" | awk '{print $1}')
if [ "$EXPECTED_CHECKSUM" != "$ACTUAL_CHECKSUM" ]; then
  echo "Checksum verification failed!"
  exit 1
fi

echo "Checksum verification passed."

# Extract the archive
TMP_DIR=$(mktemp -d)
if [[ "$OS" != "windows" ]]; then
  tar -xzf "$TMP_ARCHIVE" -C "$TMP_DIR"
else
  unzip -q "$TMP_ARCHIVE" -d "$TMP_DIR"
fi

# Move the binary to the install directory
echo "Installing to $INSTALL_DIR (may require sudo)"
mv "$TMP_DIR/$FILENAME/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

echo "Installation complete! Verify with:"
echo "hazyctl version"
