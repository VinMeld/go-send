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
PACKAGE_NAME="${1:-go-send-git}"
NEW_VERSION="$2"
AUR_URL="ssh://aur@aur.archlinux.org/${PACKAGE_NAME}.git"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
AUR_DIR="$PROJECT_ROOT/packaging/aur"
TEMP_DIR=$(mktemp -d)

echo "Package Name: $PACKAGE_NAME"
echo "Project Root: $PROJECT_ROOT"
echo "AUR Directory: $AUR_DIR"
echo "Temp Directory: $TEMP_DIR"

# Determine which PKGBUILD to use
if [ "$PACKAGE_NAME" == "go-send-bin" ]; then
    SOURCE_PKGBUILD="$AUR_DIR/PKGBUILD-bin"
else
    SOURCE_PKGBUILD="$AUR_DIR/PKGBUILD"
fi

if [ ! -f "$SOURCE_PKGBUILD" ]; then
    echo -e "${RED}Error: PKGBUILD not found at $SOURCE_PKGBUILD${NC}"
    exit 1
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

# Copy PKGBUILD
cp "$SOURCE_PKGBUILD" PKGBUILD

# Update Version if provided
if [ -n "$NEW_VERSION" ]; then
    echo "Updating pkgver to $NEW_VERSION..."
    sed -i "s/^pkgver=.*/pkgver=$NEW_VERSION/" PKGBUILD
    sed -i "s/^pkgrel=.*/pkgrel=1/" PKGBUILD
    
    # Update checksums
    echo "Updating checksums..."
    if command -v updpkgsums >/dev/null 2>&1; then
        updpkgsums
    else
        echo "updpkgsums not found, attempting to use makepkg -g..."
        # This is a bit hacky, but better than nothing if updpkgsums is missing
        # It appends new sums to the end. Ideally user should install pacman-contrib
        makepkg -g >> PKGBUILD
    fi
fi

# Generate .SRCINFO
echo "Generating .SRCINFO..."
makepkg --printsrcinfo > .SRCINFO

# Check for changes
if [ -z "$(git status --porcelain)" ]; then
    echo "No changes to submit."
    rm -rf "$TEMP_DIR"
    exit 0
fi

# Commit and Push
echo "Committing and pushing..."
git add PKGBUILD .SRCINFO
git commit -m "Update package: ${NEW_VERSION:-$(date +%Y-%m-%d)}"
git push origin master

echo -e "${GREEN}Successfully submitted $PACKAGE_NAME to AUR!${NC}"

# Cleanup
rm -rf "$TEMP_DIR"
