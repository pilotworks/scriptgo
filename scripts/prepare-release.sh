#!/usr/bin/env bash

set -euo pipefail

# ==============================================================================
# ScriptGo Local Release Preparation Script
# Usage: ./scripts/prepare-release.sh <version>
# Example: ./scripts/prepare-release.sh 0.1.0 (or v0.1.0)
# ==============================================================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}${BOLD}==>${NC} ${BOLD}$*${NC}"
}

log_success() {
    echo -e "${GREEN}${BOLD}✔${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}${BOLD}⚠${NC} $*"
}

log_error() {
    echo -e "${RED}${BOLD}✖${NC} $*" >&2
}

# 1. Check version argument
if [ $# -lt 1 ]; then
    log_error "Missing version argument."
    echo "Usage: $0 <version>"
    echo "Example: $0 0.1.0"
    exit 1
fi

RAW_VERSION="$1"
# Strip leading 'v' if provided
VERSION="${RAW_VERSION#v}"
TAG_NAME="v${VERSION}"

# Validate SemVer format: X.Y.Z or X.Y.Z-beta etc.
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
    log_error "Invalid version format '$VERSION'. Expected SemVer (e.g. 0.1.0 or 0.1.0-alpha.1)"
    exit 1
fi

log_info "Preparing release for ${BOLD}${TAG_NAME}${NC}..."

# 2. Check for existing tag
if git rev-parse "$TAG_NAME" >/dev/null 2>&1; then
    log_error "Tag '$TAG_NAME' already exists in local git repository!"
    exit 1
fi

# 3. Check git working directory status
if ! git diff-index --quiet HEAD --; then
    log_warn "You have uncommitted changes in your working tree."
    read -p "Do you want to proceed anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "Release preparation aborted."
        exit 1
    fi
fi

# 4. Generate / Update CHANGELOG.md if not already present for this version
CHANGELOG_FILE="CHANGELOG.md"
DATE=$(date +"%Y-%m-%d")

if [ ! -f "$CHANGELOG_FILE" ]; then
    cat <<EOF > "$CHANGELOG_FILE"
# Changelog

All notable changes to ScriptGo will be documented in this file.

## [$VERSION] - $DATE

### Highlights
- Initial release with 100% parity across 136 corpus regression test cases against Node.js v22+.
- High-performance AOT compilation using LLVM and Clang / Zig CC.
- Full support for Classes, Async/Await, Generators, BigInt, Symbols, and Node.js core modules (fs, path, os, crypto).
EOF
    log_success "Created $CHANGELOG_FILE for $TAG_NAME."
else
    if ! grep -q "## \[$VERSION\]" "$CHANGELOG_FILE"; then
        log_info "Updating $CHANGELOG_FILE with entry for $TAG_NAME..."
        # Prepend new version entry after top header
        TEMP_CHANGELOG=$(mktemp)
        awk -v ver="$VERSION" -v dt="$DATE" '
        BEGIN { inserted = 0 }
        /^## / && !inserted {
            print "## [" ver "] - " dt "\n"
            print "### Changes"
            print "- Release " ver "\n"
            inserted = 1
        }
        { print }
        END {
            if (!inserted) {
                print "\n## [" ver "] - " dt "\n"
                print "### Changes"
                print "- Release " ver "\n"
            }
        }' "$CHANGELOG_FILE" > "$TEMP_CHANGELOG"
        mv "$TEMP_CHANGELOG" "$CHANGELOG_FILE"
        log_success "Updated $CHANGELOG_FILE."
    fi
fi

# 5. Git commit changelog if modified
if ! git diff --quiet "$CHANGELOG_FILE" 2>/dev/null; then
    git add "$CHANGELOG_FILE"
    git commit -m "chore: release ${TAG_NAME}"
    log_success "Committed changelog updates: 'chore: release ${TAG_NAME}'"
fi

# 6. Create annotated Git Tag
log_info "Creating annotated git tag '${TAG_NAME}'..."
git tag -a "$TAG_NAME" -m "Release ${TAG_NAME}"
log_success "Created git tag ${BOLD}${TAG_NAME}${NC}."

echo
echo -e "${GREEN}${BOLD}========================================================================${NC}"
echo -e "${GREEN}${BOLD}  Release ${TAG_NAME} is ready!${NC}"
echo -e "${GREEN}${BOLD}========================================================================${NC}"
echo -e "To publish this release and trigger automated binary builds on GitHub Actions, run:"
echo
echo -e "  ${BOLD}git push origin main${NC}"
echo -e "  ${BOLD}git push origin ${TAG_NAME}${NC}"
echo
