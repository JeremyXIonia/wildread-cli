#!/bin/sh
set -eu

REPO="JeremyXIonia/wildread-cli"
BIN="wildread-cli"
VERSION="${1:-latest}"
INSTALL_DIR="${WILDREAD_INSTALL_DIR:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$os" in
  darwin) os="darwin" ;;
  linux) echo "Linux release artifacts are not published yet. Use: go install github.com/JeremyXIonia/wildread-cli@latest" >&2; exit 1 ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

asset="$BIN-$os-$arch.tar.gz"
base="https://github.com/$REPO/releases"
if [ "$VERSION" = "latest" ]; then
  url="$base/latest/download/$asset"
else
  url="$base/download/$VERSION/$asset"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

mkdir -p "$INSTALL_DIR"
echo "Downloading $url"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$tmp/$asset"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$url" -O "$tmp/$asset"
else
  echo "curl or wget is required" >&2
  exit 1
fi

tar -xzf "$tmp/$asset" -C "$tmp"
install -m 0755 "$tmp/$BIN" "$INSTALL_DIR/$BIN"

echo "Installed $BIN to $INSTALL_DIR/$BIN"
case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    echo "You can now run: $BIN"
    ;;
  *)
    echo ""
    echo "To run '$BIN' from anywhere, add $INSTALL_DIR to PATH."
    echo "For modern macOS zsh, run:"
    echo ""
    echo "  mkdir -p \"$INSTALL_DIR\""
    echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
    echo "  source ~/.zshrc"
    echo ""
    echo "Or reopen your terminal after updating ~/.zshrc."
    ;;
esac
