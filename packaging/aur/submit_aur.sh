#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== AUR Submission Helper ===${NC}"

# Check if SSH key for AUR is configured
if ! ssh -T aur@aur.archlinux.org help >/dev/null 2>&1; then
    echo -e "${RED}Error: Cannot authenticate with AUR.${NC}"
    echo "Please ensure you have added your SSH key to your AUR account."
    echo "And that you have configured ~/.ssh/config if necessary."
    echo "Try running: ssh aur@aur.archlinux.org help"
    exit 1
fi

# Variables
PACKAGE_NAME="go-send-git"
AUR_URL="ssh://aur@aur.archlinux.org/${PACKAGE_NAME}.git"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
AUR_DIR="$PROJECT_ROOT/packaging/aur"
TEMP_DIR=$(mktemp -d)

echo "Project Root: $PROJECT_ROOT"
echo "AUR Directory: $AUR_DIR"
echo "Temp Directory: $TEMP_DIR"

# Ensure .SRCINFO exists
if [ ! -f "$AUR_DIR/.SRCINFO" ]; then
    echo "Generating .SRCINFO..."
    cd "$AUR_DIR"
    makepkg --printsrcinfo > .SRCINFO
fi

# Clone AUR repo
echo "Cloning AUR repository..."
cd "$TEMP_DIR"
if ! git clone "$AUR_URL"; then
    echo -e "${RED}Failed to clone AUR repo. Does it exist?${NC}"
    echo "If not, create it at https://aur.archlinux.org/packages/submit/"
    exit 1
fi

cd "$PACKAGE_NAME"

# Ensure we are on master branch (AUR requirement)
if ! git rev-parse --verify master >/dev/null 2>&1; then
    git checkout -b master
else
    git checkout master
fi

# Copy files
echo "Copying files..."
cp "$AUR_DIR/PKGBUILD" .
cp "$AUR_DIR/.SRCINFO" .

# Check for changes
if [ -z "$(git status --porcelain)" ]; then
    echo "No changes to submit."
    rm -rf "$TEMP_DIR"
    exit 0
fi

# Commit and Push
echo "Committing and pushing..."
git add PKGBUILD .SRCINFO
git commit -m "Update package: $(date +%Y-%m-%d)"
git push origin master

echo -e "${GREEN}Successfully submitted to AUR!${NC}"

# Cleanup
rm -rf "$TEMP_DIR"
