#!/usr/bin/env bash
set -e

REPO="Naveen-M99/conn"
BINARY_NAME="conn"

echo "==> Detecting system architecture..."
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64)
    BIN_ARCH="amd64"
    ;;
  i386|i686|x86)
    BIN_ARCH="386"
    ;;
  aarch64|arm64)
    echo "Error: ARM64 builds are not currently provided. Please compile from source."
    exit 1
    ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/conn-linux-${BIN_ARCH}"

# Select target directory based on user privileges
if [[ $EUID -eq 0 ]]; then
  TARGET_DIR="/usr/local/bin"
  SUDO=""
elif command -v sudo >/dev/null 2>&1 && sudo -v 2>/dev/null; then
  TARGET_DIR="/usr/local/bin"
  SUDO="sudo"
else
  TARGET_DIR="${HOME}/.local/bin"
  SUDO=""
fi

$SUDO mkdir -p "${TARGET_DIR}"

echo "==> Downloading ${BINARY_NAME} for linux/${BIN_ARCH}..."
$SUDO curl -fsSL "${DOWNLOAD_URL}" -o "${TARGET_DIR}/${BINARY_NAME}"
$SUDO chmod +x "${TARGET_DIR}/${BINARY_NAME}"

echo "==> Verification..."
if ! command -v "${BINARY_NAME}" >/dev/null 2>&1; then
  echo ""
  echo "Notice: '${TARGET_DIR}' is not in your current PATH."
  echo "Run this command or add it to ~/.bashrc / ~/.zshrc:"
  echo "  export PATH=\"${TARGET_DIR}:\$PATH\""
  echo ""
fi

echo "==> Successfully installed ${BINARY_NAME} to ${TARGET_DIR}/${BINARY_NAME}!"
echo "Run '${BINARY_NAME} --help' to get started."