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

# Get version
if [[ "$VERSION" == "latest" ]]; then
  VERSION=$(curl -s "$GH_API/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi
STRIPPED_VERSION="${VERSION#v}"

echo "Installing $APP version $VERSION for $OS/$ARCH..."

# Determine packaging format based on available tools
EXT=""
INSTALL_CMD=""

if [[ "$OS" == "linux" ]]; then
  if command -v dpkg &>/dev/null; then
    EXT="deb"
    INSTALL_CMD="sudo dpkg -i"
  elif command -v rpm &>/dev/null; then
    EXT="rpm"
    INSTALL_CMD="sudo rpm -Uvh"
  elif command -v apk &>/dev/null; then
    EXT="apk"
    INSTALL_CMD="sudo apk add --allow-untrusted"
  elif [[ -f /etc/arch-release || -f /etc/artix-release ]]; then
    EXT="pkg.tar.zst"
    INSTALL_CMD="sudo pacman -U"
  else
    echo "❌ No supported package manager found (dpkg, rpm, apk, pacman)."
    exit 1
  fi
elif [[ "$OS" == "darwin" ]]; then
  EXT="tar.gz"
  INSTALL_CMD="sudo tar -xzf - -C /usr/local/bin"
fi

# Get asset URL from GitHub API (instead of guessing name)
ASSET_URL=$(curl -s "$GH_API/tags/$VERSION" |
  grep "browser_download_url" |
  grep -i "${EXT}" |
  grep -i "${ARCH}" |
  cut -d '"' -f 4)

if [[ -z "$ASSET_URL" ]]; then
  echo "❌ Could not find a $EXT package for $OS/$ARCH in release $VERSION"
  exit 1
fi

FILENAME=$(basename "$ASSET_URL")

echo "Downloading $FILENAME..."
curl -LO "$ASSET_URL"

echo "Installing $FILENAME..."
if [[ "$EXT" == "tar.gz" ]]; then
  mkdir -p /tmp/$APP-install
  tar -xzf "$FILENAME" -C /tmp/$APP-install
  sudo mv /tmp/$APP-install/$APP /usr/local/bin/
  rm -rf /tmp/$APP-install
else
  $INSTALL_CMD "$FILENAME"
fi

rm -f "$FILENAME"
echo "$APP $VERSION installed successfully ✅"
