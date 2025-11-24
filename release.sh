#!/bin/bash

# release.sh - Automate codemap release process

set -e

# Configuration
TAP_DIR="../homebrew-tap"
FORMULA_FILE="codemap.rb"
REPO_URL="https://github.com/JordanCoin/codemap"

# 1. Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "❌ Error: You have uncommitted changes."
    echo "Please commit or stash them before releasing."
    exit 1
fi

# 2. Get current version from git tags
# If no tags exist, default to v1.0 (since user just released 1.0 manually)
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0")
# Remove 'v' prefix
CURRENT_VERSION=${CURRENT_VERSION#v}

echo "Current version: $CURRENT_VERSION"

# 3. Calculate next version
IFS='.' read -r -a parts <<< "$CURRENT_VERSION"
MAJOR=${parts[0]}
MINOR=${parts[1]}

# User requested logic: 1.9 -> 2.0
if [ "$MINOR" -ge 9 ]; then
    NEXT_MAJOR=$((MAJOR + 1))
    NEXT_MINOR=0
else
    NEXT_MAJOR=$MAJOR
    NEXT_MINOR=$((MINOR + 1))
fi

NEXT_VERSION="$NEXT_MAJOR.$NEXT_MINOR"
TAG_NAME="v$NEXT_VERSION"

echo "Preparing to release: $TAG_NAME"
read -p "Press enter to continue or Ctrl+C to cancel..."

# 4. Create git tag and push
echo "Creating git tag $TAG_NAME..."
git tag "$TAG_NAME"
git push origin "$TAG_NAME"

# 5. Create GitHub Release (if gh is installed)
if command -v gh &> /dev/null; then
    echo "Creating GitHub Release..."
    gh release create "$TAG_NAME" --generate-notes
else
    echo "⚠️ 'gh' CLI not found. Skipping GitHub Release creation."
    echo "You can create it manually at: $REPO_URL/releases/new?tag=$TAG_NAME"
fi

echo "Waiting 5 seconds for GitHub to generate tarball..."
sleep 5

# 6. Calculate SHA256
TARBALL_URL="$REPO_URL/archive/refs/tags/$TAG_NAME.tar.gz"
echo "Downloading tarball from $TARBALL_URL..."
SHA256=$(curl -L -s "$TARBALL_URL" | shasum -a 256 | awk '{print $1}')

if [ -z "$SHA256" ] || [ ${#SHA256} -ne 64 ]; then
    echo "Error: Failed to calculate valid SHA256. Got: $SHA256"
    exit 1
fi

echo "Calculated SHA256: $SHA256"

# 7. Update codemap.rb locally
echo "Updating $FORMULA_FILE..."
# Use sed to replace url and sha256 (macOS compatible sed -i '')
sed -i '' "s|url \".*\"|url \"$TARBALL_URL\"|" "$FORMULA_FILE"
sed -i '' "s|sha256 \".*\"|sha256 \"$SHA256\"|" "$FORMULA_FILE"

# 8. Push local changes
echo "Committing updated formula to main repo..."
git add "$FORMULA_FILE"
git commit -m "Bump version to $TAG_NAME"
git push origin main

# 9. Update Homebrew Tap
if [ -d "$TAP_DIR" ]; then
    echo "Updating Homebrew Tap at $TAP_DIR..."
    cp "$FORMULA_FILE" "$TAP_DIR/$FORMULA_FILE"
    
    # Capture current directory
    CURRENT_DIR=$(pwd)
    
    cd "$TAP_DIR"
    git add "$FORMULA_FILE"
    git commit -m "Update codemap to $TAG_NAME"
    git push origin main
    
    # Return to original directory
    cd "$CURRENT_DIR"
    
    echo "Homebrew Tap updated successfully!"
else
    echo "Warning: Directory $TAP_DIR not found."
    echo "Skipping automatic tap update."
    echo "Please manually copy $FORMULA_FILE to your homebrew-tap repo and push."
fi

echo "✅ Release $TAG_NAME complete!"
