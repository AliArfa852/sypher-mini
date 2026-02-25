#!/bin/sh
# Sypher-mini install script (Linux/macOS)
# Usage: curl -fsSL https://raw.githubusercontent.com/sypherexx/sypher-mini/main/scripts/install.sh | sh
# Or: curl -fsSL ... | sh -s -- /path/to/install

set -e
INSTALL_DIR="${1:-$HOME/.local/bin}"
REPO="${SYPHER_REPO:-https://github.com/sypherexx/sypher-mini}"
BRANCH="${SYPHER_BRANCH:-main}"

echo "Sypher-mini installer"
echo "--------------------"

if ! command -v go >/dev/null 2>&1; then
  echo "Go 1.22+ required. Install from https://go.dev/doc/install"
  exit 1
fi

TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT
cd "$TMP"
echo "Cloning $REPO..."
git clone --depth 1 -b "$BRANCH" "$REPO" .
echo "Building..."
go build -o sypher ./cmd/sypher
mkdir -p "$INSTALL_DIR"
mv sypher "$INSTALL_DIR/sypher"
echo "Installed to $INSTALL_DIR/sypher"
echo ""
echo "Next steps:"
echo "  1. $INSTALL_DIR/sypher onboard"
echo "  2. Set API key: sypher config set providers.gemini.api_key YOUR_KEY"
echo "  3. $INSTALL_DIR/sypher agent -m 'Hello'"
echo ""
echo "Add to PATH: export PATH=\"$INSTALL_DIR:\$PATH\""
