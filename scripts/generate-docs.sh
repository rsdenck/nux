#!/bin/bash
# scripts/generate-docs.sh
# Validates markdown, generates indices, and checks links for NUX documentation.

set -e

DOCS_DIR="docs/wiki"
MAN_DIR="docs/man"

echo "Starting documentation generation and validation..."

# 1. Validate Markdown Files
echo "Validating Markdown files..."
if command -v markdownlint >/dev/null 2>&1; then
    markdownlint "$DOCS_DIR"/*.md
    echo "Markdown validation passed."
else
    echo "Warning: markdownlint not found. Skipping validation."
fi

# 2. Generate Automatic Index (Home.md update)
echo "Updating Wiki Index..."
INDEX_FILE="$DOCS_DIR/Home.md"

# Ensure header exists
if ! grep -q "# Home" "$INDEX_FILE"; then
    echo "# Home" > "$INDEX_FILE"
    echo "" >> "$INDEX_FILE"
    echo "Welcome to the **NUX** Wiki!" >> "$INDEX_FILE"
    echo "" >> "$INDEX_FILE"
fi

# 3. Check for broken links (Simple check)
echo "Checking for broken local links..."
grep -r "\[.*\](.*)" "$DOCS_DIR" | while read -r line; do
    link=$(echo "$line" | sed -n 's/.*(\(.*\)).*/\1/p')
    # Skip http/https links
    if [[ "$link" == http* ]]; then
        continue
    fi
    # Remove anchors
    file_path="$DOCS_DIR/${link%%#*}"
    if [ ! -f "$file_path" ] && [ ! -d "$file_path" ]; then
        echo "Warning: Possible broken link in $DOCS_DIR: $link"
    fi
done

echo "Documentation generation complete."
