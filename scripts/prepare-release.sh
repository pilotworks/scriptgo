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

# 4. Auto-generate / Update CHANGELOG.md from git commits
CHANGELOG_FILE="CHANGELOG.md"
DATE=$(date +"%Y-%m-%d")

log_info "Generating changelog entries from git commits for ${TAG_NAME}..."

python3 - << EOF
import subprocess, re, os, datetime

version = "${VERSION}"
tag_name = "${TAG_NAME}"
changelog_path = "${CHANGELOG_FILE}"
date_str = "${DATE}"

# Find previous tag
prev_tag = ""
try:
    tags = subprocess.check_output(['git', 'tag', '--sort=-creatordate']).decode('utf-8').strip().splitlines()
    tags = [t for t in tags if t != tag_name]
    if tags:
        prev_tag = tags[0]
except Exception:
    pass

log_cmd = ['git', 'log', '--no-merges', '--pretty=format:%h %s']
if prev_tag:
    log_cmd.append(f'{prev_tag}..HEAD')

commits = subprocess.check_output(log_cmd).decode('utf-8').strip().splitlines()

categories = {
    'Features': [],
    'Bug Fixes': [],
    'Performance': [],
    'Refactoring': [],
    'Documentation': [],
    'Maintenance & Chores': [],
    'Other': []
}

for line in commits:
    if not line.strip(): continue
    parts = line.split(' ', 1)
    if len(parts) != 2: continue
    sha, msg = parts
    
    # Skip internal release chore commits
    if re.match(r'^(chore(\([^\)]+\))?:\s*release|release\s*v)', msg, re.IGNORECASE):
        continue
    
    entry = f'- {msg} ({sha})'
    if re.match(r'^feat(\([^\)]+\))?:', msg, re.I):
        categories['Features'].append(entry)
    elif re.match(r'^fix(\([^\)]+\))?:', msg, re.I):
        categories['Bug Fixes'].append(entry)
    elif re.match(r'^perf(\([^\)]+\))?:', msg, re.I):
        categories['Performance'].append(entry)
    elif re.match(r'^refactor(\([^\)]+\))?:', msg, re.I):
        categories['Refactoring'].append(entry)
    elif re.match(r'^docs(\([^\)]+\))?:', msg, re.I):
        categories['Documentation'].append(entry)
    elif re.match(r'^(chore|ci|build|test)(\([^\)]+\))?:', msg, re.I):
        categories['Maintenance & Chores'].append(entry)
    else:
        categories['Other'].append(entry)

out_lines = [f'## [{version}] - {date_str}\n']
has_entries = False
for cat, entries in categories.items():
    if entries:
        has_entries = True
        out_lines.append(f'### {cat}')
        out_lines.extend(entries)
        out_lines.append('')

if not has_entries:
    out_lines.append('### Changes')
    out_lines.append(f'- Release {version}\n')

new_section = '\n'.join(out_lines).strip()

existing_content = ""
if os.path.exists(changelog_path):
    with open(changelog_path, 'r', encoding='utf-8') as f:
        existing_content = f.read()

if not existing_content:
    header = "# Changelog\n\nAll notable changes to ScriptGo will be documented in this file.\n\n"
    final_content = header + new_section + "\n"
else:
    # If version block exists, replace it
    pattern = rf'## \[{re.escape(version)}\][\s\S]*?(?=\n## \[|\Z)'
    if re.search(pattern, existing_content):
        final_content = re.sub(pattern, new_section + "\n", existing_content, count=1)
    else:
        # Prepend after header
        header_match = re.search(r'^(#\s*Changelog[^\n]*\n+([^\n]+\n+)*?)(?=## |\Z)', existing_content, re.MULTILINE)
        if header_match:
            header_end = header_match.end()
            final_content = existing_content[:header_end] + new_section + "\n\n" + existing_content[header_end:]
        else:
            final_content = "# Changelog\n\n" + new_section + "\n\n" + existing_content

with open(changelog_path, 'w', encoding='utf-8') as f:
    f.write(final_content.strip() + "\n")

EOF

log_success "Updated ${CHANGELOG_FILE} with commit log entries."

# 5. Git commit changelog if modified or newly created
git add "$CHANGELOG_FILE"
if ! git diff --cached --quiet; then
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
