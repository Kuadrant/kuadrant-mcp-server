#!/bin/bash

# Script to pull latest Kuadrant documentation based on mkdocs.yml configuration
# This script reads the multirepo plugin config to determine what files to extract

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCS_CONFIG_DIR="${SCRIPT_DIR}/.docs-config-temp"
OPERATOR_DIR="${SCRIPT_DIR}/.kuadrant-operator-temp"
AUTHORINO_DIR="${SCRIPT_DIR}/.authorino-temp"
DOCS_CONFIG_REPO="https://github.com/kuadrant/docs.kuadrant.io.git"
OUTPUT_DIR="${SCRIPT_DIR}/extracted-docs"

echo "📚 Updating Kuadrant documentation..."

# First, get the docs config to read mkdocs.yml
if [ -d "$DOCS_CONFIG_DIR" ]; then
    echo "Updating docs.kuadrant.io config..."
    cd "$DOCS_CONFIG_DIR"
    git fetch origin
    git reset --hard origin/main
else
    echo "Cloning docs.kuadrant.io config..."
    git clone "$DOCS_CONFIG_REPO" "$DOCS_CONFIG_DIR"
fi

# Parse mkdocs.yml to extract file paths
echo "📖 Reading mkdocs.yml configuration..."

# Extract kuadrant-operator imports using yq (if available) or fallback to grep/sed
if command -v yq &> /dev/null; then
    echo "Using yq to parse mkdocs.yml..."
    OPERATOR_FILES=$(yq eval '.plugins[] | select(has("multirepo")) | .multirepo.nav_repos[] | select(.name == "kuadrant-operator") | .imports[]' "$DOCS_CONFIG_DIR/mkdocs.yml" | sed 's/^//')
    AUTHORINO_FILES=$(yq eval '.plugins[] | select(has("multirepo")) | .multirepo.nav_repos[] | select(.name == "authorino") | .imports[]' "$DOCS_CONFIG_DIR/mkdocs.yml" | sed 's/^//')
else
    echo "yq not found, using grep/sed to parse mkdocs.yml..."
    # Extract lines between kuadrant-operator section and next section
    OPERATOR_FILES=$(awk '/name: kuadrant-operator/,/name: authorino/' "$DOCS_CONFIG_DIR/mkdocs.yml" | grep "^ *- /" | sed 's/^ *- //')
    AUTHORINO_FILES=$(awk '/name: authorino/,/^[[:space:]]*-[[:space:]]*name:/' "$DOCS_CONFIG_DIR/mkdocs.yml" | grep "^ *- /" | sed 's/^ *- //')
fi

# Clone/update source repositories
echo "📦 Updating source repositories..."

# Update kuadrant-operator
if [ -d "$OPERATOR_DIR" ]; then
    echo "Updating kuadrant-operator..."
    cd "$OPERATOR_DIR"
    git fetch origin
    git reset --hard origin/main
else
    echo "Cloning kuadrant-operator..."
    git clone https://github.com/kuadrant/kuadrant-operator.git "$OPERATOR_DIR"
fi

# Update authorino
if [ -d "$AUTHORINO_DIR" ]; then
    echo "Updating authorino..."
    cd "$AUTHORINO_DIR"
    git fetch origin
    git reset --hard origin/main
else
    echo "Cloning authorino..."
    git clone https://github.com/kuadrant/authorino.git "$AUTHORINO_DIR"
fi

# Create output directory structure
echo "📝 Extracting documentation files..."
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR/kuadrant-operator"
mkdir -p "$OUTPUT_DIR/authorino"

# Copy kuadrant-operator files
echo "Extracting kuadrant-operator docs..."
while IFS= read -r file; do
    # Skip empty lines and image files
    [[ -z "$file" ]] && continue
    [[ "$file" == *.png ]] && continue
    [[ "$file" == *.jpg ]] && continue

    # Remove leading slash if present
    file="${file#/}"

    src_file="$OPERATOR_DIR/$file"
    if [ -f "$src_file" ]; then
        # Create directory structure in output
        dest_dir="$OUTPUT_DIR/kuadrant-operator/$(dirname "$file")"
        mkdir -p "$dest_dir"
        cp "$src_file" "$OUTPUT_DIR/kuadrant-operator/$file"
        echo "  ✓ $file"
    else
        echo "  ⚠ $file not found"
    fi
done <<< "$OPERATOR_FILES"

# Copy authorino files
echo "Extracting authorino docs..."
while IFS= read -r file; do
    # Skip empty lines and image files
    [[ -z "$file" ]] && continue
    [[ "$file" == *.png ]] && continue
    [[ "$file" == *.jpg ]] && continue

    # Remove leading slash if present
    file="${file#/}"

    src_file="$AUTHORINO_DIR/$file"
    if [ -f "$src_file" ]; then
        # Create directory structure in output
        dest_dir="$OUTPUT_DIR/authorino/$(dirname "$file")"
        mkdir -p "$dest_dir"
        cp "$src_file" "$OUTPUT_DIR/authorino/$file"
        echo "  ✓ $file"
    else
        echo "  ⚠ $file not found"
    fi
done <<< "$AUTHORINO_FILES"

# Create a summary of what was extracted
echo ""
echo "📊 Creating extraction summary..."
cat > "$OUTPUT_DIR/extraction-summary.txt" <<EOF
Documentation Extraction Summary
================================
Generated: $(date)

Kuadrant Operator Files:
------------------------
$(find "$OUTPUT_DIR/kuadrant-operator" -name "*.md" -type f | sed "s|$OUTPUT_DIR/kuadrant-operator/||" | sort)

Authorino Files:
----------------
$(find "$OUTPUT_DIR/authorino" -name "*.md" -type f | sed "s|$OUTPUT_DIR/authorino/||" | sort)

Total files extracted: $(find "$OUTPUT_DIR" -name "*.md" -type f | wc -l | tr -d ' ')
EOF

echo ""
echo "✅ Documentation extraction complete!"
echo "📁 Extracted files are in: $OUTPUT_DIR"
echo ""
echo "Structure:"
echo "  $OUTPUT_DIR/"
echo "  ├── kuadrant-operator/   # Files from kuadrant-operator repo"
echo "  ├── authorino/          # Files from authorino repo"
echo "  └── extraction-summary.txt"
echo ""
echo "Next steps:"
echo "1. Review extraction-summary.txt to see what was extracted"
echo "2. Run process-docs.go to convert markdown to Go resources (if needed)"
echo "3. Or manually update resources.go with the new content"

# Optional: Install yq suggestion
if ! command -v yq &> /dev/null; then
    echo ""
    echo "💡 Tip: Install yq for better YAML parsing:"
    echo "   brew install yq  # on macOS"
    echo "   This will make the extraction more reliable"
fi