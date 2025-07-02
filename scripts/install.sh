#!/usr/bin/env bash
set -euo pipefail

REPO="crumbyte/noxdir"
APP="noxdir"
VERSION="${1:-latest}"

GH_API="https://api.github.com/repos/$REPO/releases"
GH_DL="https://github.com/$REPO/releases/download"

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

# Normalize OS
case "$OS" in
  Linux*)   OS="linux" ;;
  Darwin*)  OS="darwin" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Normalize ARCH
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  i386|i686) ARCH="i386" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Resolve version
if [[ "$VERSION" == "latest" ]]; then
  VERSION=$(curl -s "$GH_API/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

echo "Installing $APP version $VERSION for $OS/$ARCH..."

# Determine package file name
PKG=""
EXT=""
INSTALL_CMD=""

if [[ "$OS" == "linux" ]]; then
  if command -v dpkg &>/dev/null; then
    EXT="deb"
    PKG="${APP}_${VERSION}_${ARCH}.${EXT}"
    INSTALL_CMD="sudo dpkg -i $PKG"
  elif command -v rpm &>/dev/null; then
    EXT="rpm"
    PKG="${APP}-${VERSION}-1.${ARCH}.${EXT}"
    INSTALL_CMD="sudo rpm -Uvh $PKG"
  elif command -v apk &>/dev/null; then
    EXT="apk"
    PKG="${APP}_${VERSION}_${ARCH}.${EXT}"
    INSTALL_CMD="sudo apk add --allow-untrusted $PKG"
  elif [[ -f /etc/arch-release || -f /etc/artix-release ]]; then
    EXT="pkg.tar.zst"
    PKG="${APP}-${VERSION}-1-${ARCH}.${EXT}"
    INSTALL_CMD="sudo pacman -U $PKG"
  else
    echo "No supported package manager found (dpkg, rpm, apk, pacman)."
    exit 1
  fi
elif [[ "$OS" == "darwin" ]]; then
  EXT="tar.gz"
  PKG="${APP}_${OS^}_${ARCH}.${EXT}"
  INSTALL_CMD="sudo tar -xzf $PKG -C /usr/local/bin"
fi

if [[ -z "$PKG" ]]; then
  echo "Unable to determine package file name for $OS/$ARCH"
  exit 1
fi

echo "Downloading $PKG..."
curl -LO "$GH_DL/$VERSION/$PKG"

echo "Installing..."
eval "$INSTALL_CMD"

rm -f "$PKG"

echo "$APP $VERSION installed successfully ✅"
