#!/bin/sh

set -eu

REPOSITORY=${SCRIPTGO_REPOSITORY:-pilotworks/scriptgo}
VERSION=${SCRIPTGO_VERSION:-latest}
INSTALL_DIR=${SCRIPTGO_INSTALL_DIR:-"${HOME}/.scriptgo"}

usage() {
    cat <<'EOF'
Install ScriptGo from a GitHub release.

Usage:
  install.sh [--version VERSION] [--install-dir DIR]

Environment variables:
  SCRIPTGO_VERSION       Release version, with or without a leading v
  SCRIPTGO_INSTALL_DIR   Destination directory (default: $HOME/.scriptgo)
  SCRIPTGO_REPOSITORY    GitHub repository (default: pilotworks/scriptgo)
EOF
}

fail() {
    printf 'scriptgo installer: %s\n' "$*" >&2
    exit 1
}

while [ "$#" -gt 0 ]; do
    case "$1" in
        --version)
            [ "$#" -ge 2 ] || fail "--version requires a value"
            VERSION=$2
            shift 2
            ;;
        --install-dir)
            [ "$#" -ge 2 ] || fail "--install-dir requires a value"
            INSTALL_DIR=$2
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            fail "unknown argument: $1"
            ;;
    esac
done

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

case "$(uname -s)" in
    Darwin) OS=darwin ;;
    Linux) OS=linux ;;
    *) fail "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
    x86_64|amd64) ARCH=amd64 ;;
    arm64|aarch64) ARCH=arm64 ;;
    *) fail "unsupported architecture: $(uname -m)" ;;
esac

if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -fsSL \
        -H 'Accept: application/vnd.github+json' \
        "https://api.github.com/repos/${REPOSITORY}/releases?per_page=1" \
        | sed -n 's/.*"tag_name": "\([^"]*\)".*/\1/p' \
        | head -n 1)
    [ -n "$VERSION" ] || fail "could not determine the latest GitHub release"
fi

TAG=$VERSION
case "$TAG" in
    v*) ;;
    *) TAG="v${TAG}" ;;
esac
VERSION=${TAG#v}

ASSET="scriptgo_${VERSION}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPOSITORY}/releases/download/${TAG}"

TEMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t scriptgo)
STAGED_PATH=
cleanup() {
    rm -rf "$TEMP_DIR"
    if [ -n "$STAGED_PATH" ]; then
        rm -f "$STAGED_PATH"
    fi
}
trap cleanup EXIT HUP INT TERM

printf 'Downloading ScriptGo %s for %s/%s...\n' "$TAG" "$OS" "$ARCH"
curl -fL --retry 3 --output "$TEMP_DIR/$ASSET" "$BASE_URL/$ASSET"
curl -fL --retry 3 --output "$TEMP_DIR/SHA256SUMS" "$BASE_URL/SHA256SUMS"

EXPECTED=$(awk -v asset="$ASSET" '$2 == asset || $2 == "*" asset { print $1; exit }' "$TEMP_DIR/SHA256SUMS")
[ -n "$EXPECTED" ] || fail "checksum for $ASSET is missing from SHA256SUMS"

if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL=$(sha256sum "$TEMP_DIR/$ASSET" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
    ACTUAL=$(shasum -a 256 "$TEMP_DIR/$ASSET" | awk '{print $1}')
else
    fail "sha256sum or shasum is required to verify the download"
fi

[ "$ACTUAL" = "$EXPECTED" ] || fail "checksum verification failed for $ASSET"

tar -xzf "$TEMP_DIR/$ASSET" -C "$TEMP_DIR"
[ -f "$TEMP_DIR/scriptgo" ] || fail "release archive does not contain the scriptgo binary"

mkdir -p "$INSTALL_DIR"
[ -w "$INSTALL_DIR" ] || fail "$INSTALL_DIR is not writable; choose another directory with --install-dir"
TARGET_PATH="$INSTALL_DIR/scriptgo"
if [ -x "$TARGET_PATH" ]; then
    CURRENT_VERSION=$("$TARGET_PATH" version 2>/dev/null | awk '{print $3}' || true)
    if [ -n "$CURRENT_VERSION" ]; then
        printf 'Upgrading ScriptGo from %s to %s...\n' "$CURRENT_VERSION" "$TAG"
    else
        printf 'Replacing the existing ScriptGo installation with %s...\n' "$TAG"
    fi
fi

STAGED_PATH="$INSTALL_DIR/.scriptgo.install.$$"
install -m 0755 "$TEMP_DIR/scriptgo" "$STAGED_PATH"
mv -f "$STAGED_PATH" "$TARGET_PATH"
STAGED_PATH=

printf 'Installed ScriptGo %s to %s\n' "$TAG" "$TARGET_PATH"
case ":${PATH}:" in
    *:"$INSTALL_DIR":*) ;;
    *) printf 'Add %s to your PATH to run scriptgo.\n' "$INSTALL_DIR" ;;
esac
