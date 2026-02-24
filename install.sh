#!/usr/bin/env sh
# cooledit installer — Linux and macOS
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sh
#
# To install a specific version:
#   COOLEDIT_VERSION=v0.8.0 curl -fsSL \
#     https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sh
#
# To install as root (to /usr/local/bin):
#   curl -fsSL https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sudo sh
set -eu

REPO="tomcoolpxl/cooledit"
BINARY="cooledit"
INSTALL_DIR=""
VERSION=""

# ── helpers ───────────────────────────────────────────────────────────────────

die() { echo "error: $*" >&2; exit 1; }

need() {
    command -v "$1" >/dev/null 2>&1 || die "required tool not found: $1"
}

# ── detect OS and architecture ────────────────────────────────────────────────

detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux)  OS="linux" ;;
        Darwin) OS="darwin" ;;
        *)      die "unsupported OS: $OS (only Linux and macOS are supported)" ;;
    esac

    case "$ARCH" in
        x86_64)         ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        *)              die "unsupported architecture: $ARCH" ;;
    esac
}

# ── resolve version ───────────────────────────────────────────────────────────

resolve_version() {
    if [ -n "${COOLEDIT_VERSION:-}" ]; then
        VERSION="$COOLEDIT_VERSION"
        return
    fi

    need curl

    VERSION="$(curl -fsSL \
        "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' \
        | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"

    [ -n "$VERSION" ] || die "could not determine latest version from GitHub API"
}

# ── resolve install directory ─────────────────────────────────────────────────

resolve_install_dir() {
    if [ "$(id -u)" = "0" ]; then
        INSTALL_DIR="/usr/local/bin"
    else
        INSTALL_DIR="${HOME}/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi
}

# ── download, verify, and install ────────────────────────────────────────────

install_binary() {
    need curl
    need sha256sum

    ARCHIVE_NAME="cooledit_${VERSION#v}_${OS}_${ARCH}.tar.gz"
    BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
    TMPDIR="$(mktemp -d)"
    trap 'rm -rf "$TMPDIR"' EXIT

    echo "Downloading cooledit ${VERSION} for ${OS}/${ARCH}..."
    curl -fsSL --progress-bar \
        "${BASE_URL}/${ARCHIVE_NAME}" -o "${TMPDIR}/${ARCHIVE_NAME}"
    curl -fsSL \
        "${BASE_URL}/checksums.txt" -o "${TMPDIR}/checksums.txt"

    echo "Verifying checksum..."
    (
        cd "$TMPDIR"
        grep "${ARCHIVE_NAME}" checksums.txt | sha256sum -c -
    ) || die "checksum verification failed — the downloaded archive may be corrupted"

    echo "Extracting..."
    tar -xzf "${TMPDIR}/${ARCHIVE_NAME}" -C "$TMPDIR"

    echo "Installing to ${INSTALL_DIR}/${BINARY}..."
    install -m 755 "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

    echo ""
    echo "cooledit ${VERSION} installed successfully."
    echo "  Binary: ${INSTALL_DIR}/${BINARY}"

    case ":${PATH}:" in
        *":${INSTALL_DIR}:"*) ;;
        *) echo ""
           echo "NOTE: ${INSTALL_DIR} is not in your PATH."
           echo "      Add this to your shell profile:"
           echo "        export PATH=\"\$PATH:${INSTALL_DIR}\""
           ;;
    esac
}

# ── main ──────────────────────────────────────────────────────────────────────

detect_platform
resolve_version
resolve_install_dir
install_binary
