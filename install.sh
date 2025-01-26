#!/bin/bash

set -e

REPO="hazyforge/hazyctl"
BINARY_NAME="hazyctl"
INSTALL_DIR="/usr/local/bin"

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

# Build download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/hazyctl-$OS-$ARCH"

echo "Downloading hazyctl $LATEST_TAG for $OS/$ARCH..."

# Download and install
TMP_FILE=$(mktemp)
curl -sSL --fail "$DOWNLOAD_URL" -o "$TMP_FILE"
chmod +x "$TMP_FILE"

# Move to install directory
echo "Installing to $INSTALL_DIR (may require sudo)"
sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"

echo "Installation complete! Verify with:"
echo "  hazyctl --version"