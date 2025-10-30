#!/bin/bash
# Duplicate Code Detection
# Finds potential duplicate code blocks across Go files
# Usage: ./scripts/check-duplicates.sh [--verbose]

set -e

VERBOSE="${1:-}"
THRESHOLD=10  # Lines to consider as potential duplicate

echo "🔍 Scanning for duplicate code blocks..."
echo "   Threshold: $THRESHOLD+ lines"
echo ""

# Create a temporary directory for analysis
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Function to extract code blocks and find duplicates
check_file_duplicates() {
    local file=$1
    
    # Skip test files and vendor
    if [[ $file == *"_test.go" ]] || [[ $file == */vendor/* ]]; then
        return
    fi
    
    # Create a normalized version for comparison
    awk '
    BEGIN { line_num = 0; block = ""; start_line = 0 }
    {
        line_num++
        # Strip comments and extra whitespace
        gsub(/\/\/.*$/, "")
        gsub(/^[ \t]+/, "")
        gsub(/[ \t]+$/, "")
        
        if (NF > 0) {
            if (block == "") start_line = line_num
            block = block "\n" $0
        } else {
            if (length(block) > 0 && NR - start_line > 10) {
                print FILENAME ":" start_line ":" block
            }
            block = ""
        }
    }
    END {
        if (length(block) > 0 && NR - start_line > 10) {
            print FILENAME ":" start_line ":" block
        }
    }
    ' "$file" >> "$TEMP_DIR/blocks.txt" 2>/dev/null || true
}

# Scan all Go files (except tests and vendor)
while IFS= read -r -d '' file; do
    check_file_duplicates "$file"
done < <(find ./internal ./apps ./cmd -name "*.go" -not -path "*/vendor/*" -not -name "*_test.go" -print0)

# Analyze for duplicates
if [ -f "$TEMP_DIR/blocks.txt" ]; then
    echo "📊 Analysis Results:"
    echo "==================="
    
    # Use dupl tool if available
    if command -v dupl &> /dev/null; then
        echo ""
        echo "Using dupl tool for duplicate detection:"
        dupl ./internal ./apps ./cmd 2>/dev/null | head -50 || echo "No dupl issues found"
    else
        echo "Install 'dupl' for advanced duplicate detection:"
        echo "  go install github.com/remyoudompheng/dupl@latest"
        echo ""
        
        # Fallback: manual pattern matching
        echo "Manual duplicate scan results:"
        
        # Check for similar function patterns
        grep -r "func.*GetRecentChecks\|func.*GetCheckHistory" ./internal/logs/services/ 2>/dev/null || true
        grep -r "func.*GetHealthHistory\|func.*GetRepairHistory" ./cmd/logs/handlers/ 2>/dev/null || true
    fi
else
    echo "✅ No obvious duplicate patterns detected (simple scan)"
fi

echo ""
echo "💡 Recommendation: Install dupl for comprehensive duplicate detection"
echo "   go install github.com/remyoudompheng/dupl@latest"
echo "   dupl ./internal ./apps ./cmd"
