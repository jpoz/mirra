#!/bin/sh
# Setup script to install git hooks

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
GIT_DIR=$(git rev-parse --git-dir)

echo "Installing git hooks..."

# Install pre-commit hook
if [ -f "$SCRIPT_DIR/pre-commit" ]; then
    cp "$SCRIPT_DIR/pre-commit" "$GIT_DIR/hooks/pre-commit"
    chmod +x "$GIT_DIR/hooks/pre-commit"
    echo "✓ Installed pre-commit hook"
else
    echo "❌ pre-commit hook not found"
    exit 1
fi

# Install pre-push hook
if [ -f "$SCRIPT_DIR/pre-push" ]; then
    cp "$SCRIPT_DIR/pre-push" "$GIT_DIR/hooks/pre-push"
    chmod +x "$GIT_DIR/hooks/pre-push"
    echo "✓ Installed pre-push hook"
else
    echo "❌ pre-push hook not found"
    exit 1
fi

echo ""
echo "Git hooks installed successfully!"
echo ""
echo "To disable hooks temporarily, use:"
echo "  git commit --no-verify"
echo "  git push --no-verify"
