#!/bin/bash
set -e

# obsidian-go-mcp installer
# Usage: curl -sSL https://raw.githubusercontent.com/zach-snell/obsidian-go-mcp/main/install.sh | bash

REPO="zach-snell/obsidian-go-mcp"
BINARY_NAME="obsidian-mcp"

# Default install dir, can be overridden with --user or INSTALL_DIR env var
if [ "$1" = "--user" ]; then
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
else
    INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() { echo -e "${BLUE}==>${NC} $1"; }
success() { echo -e "${GREEN}==>${NC} $1"; }
warn() { echo -e "${YELLOW}==>${NC} $1"; }
error() { echo -e "${RED}==>${NC} $1"; exit 1; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported OS: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest version
get_latest_version() {
    curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

# Check if command exists
has_command() {
    command -v "$1" >/dev/null 2>&1
}

# Main install
main() {
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     obsidian-go-mcp installer         ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════╝${NC}"
    echo ""

    # Check for curl
    if ! has_command curl; then
        error "curl is required but not installed"
    fi

    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)
    
    if [ -z "$VERSION" ]; then
        error "Failed to get latest version"
    fi

    info "Detected: ${OS}/${ARCH}"
    info "Latest version: ${VERSION}"

    # Build download URL
    FILENAME="${BINARY_NAME}-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        FILENAME="${FILENAME}.exe"
    fi
    
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
    
    info "Downloading ${FILENAME}..."
    
    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT
    
    # Download
    if ! curl -sSL "$DOWNLOAD_URL" -o "$TMP_DIR/$BINARY_NAME"; then
        error "Download failed"
    fi
    
    # Make executable
    chmod +x "$TMP_DIR/$BINARY_NAME"
    
    # Check if we can write to install dir
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        success "Installed to ${INSTALL_DIR}/${BINARY_NAME}"
    else
        info "Need sudo to install to ${INSTALL_DIR}"
        sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        success "Installed to ${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Verify installation
    if has_command "$BINARY_NAME"; then
        INSTALLED_PATH=$(which "$BINARY_NAME")
        success "Installation complete!"
        echo ""
        echo -e "  Binary: ${GREEN}${INSTALLED_PATH}${NC}"
        echo -e "  Version: ${GREEN}${VERSION}${NC}"
        echo ""
        echo "  Next steps:"
        echo "  1. Configure your MCP client (Claude Desktop, etc.)"
        echo "  2. Point it to: ${INSTALLED_PATH} /path/to/vault"
        echo ""
        echo "  Docs: https://github.com/${REPO}"
        echo ""
    else
        warn "Installed but ${BINARY_NAME} not in PATH"
        echo "  Add ${INSTALL_DIR} to your PATH, or run directly:"
        echo "  ${INSTALL_DIR}/${BINARY_NAME} /path/to/vault"
    fi
}

# Run with version flag
if [ "$1" = "--version" ] || [ "$1" = "-v" ]; then
    get_latest_version
    exit 0
fi

# Run with help flag
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "obsidian-go-mcp installer"
    echo ""
    echo "Usage:"
    echo "  curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash"
    echo "  curl -sSL ... | bash -s -- --user   # Install to ~/.local/bin (no sudo)"
    echo ""
    echo "Options:"
    echo "  --user           Install to ~/.local/bin (no sudo required)"
    echo "  --version, -v    Print latest version"
    echo "  --help, -h       Show this help"
    echo ""
    echo "Environment variables:"
    echo "  INSTALL_DIR      Installation directory (default: /usr/local/bin)"
    echo ""
    exit 0
fi

main
