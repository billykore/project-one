#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
# -e: exit on error
# -u: exit on unset variables
# -o pipefail: exit if any command in a pipe fails
set -euo pipefail

# Configuration
# PORTS_DIR: Source directory for interfaces
# MOCKS_DEST: Destination directory for generated mocks
PORTS_DIR="internal/core/ports"
MOCKS_DEST="internal/core/usecase/mocks"

# Command to run mockgen.
# Using 'go run' ensures we use the version pinned in go.mod.
# For better performance, install it locally: go install go.uber.org/mock/mockgen@latest
MOCKGEN="go run go.uber.org/mock/mockgen"

echo "Mock Generation"

# 1. Validation
if [ ! -d "$PORTS_DIR" ]; then
    echo "Error: ports directory '$PORTS_DIR' not found."
    exit 1
fi

# 2. Cleanup
# Remove old mocks to ensure no stale mocks remain.
# Warning: This will delete everything matching mock_*.go in MOCKS_DEST.
if [ -d "$MOCKS_DEST" ]; then
    echo "Cleaning up stale mocks in $MOCKS_DEST..."
    find "$MOCKS_DEST" -name "mock_*.go" -type f -delete
else
    mkdir -p "$MOCKS_DEST"
fi

# 3. Generation
# Find all .go files in the ports directory
# Using a glob is fine here since we checked the directory exists
files=("$PORTS_DIR"/*.go)
total_files=${#files[@]}
current=0

# Check if the glob found anything (handles the case where the directory is empty)
if [ "$total_files" -eq 0 ] || [ ! -e "${files[0]}" ]; then
    echo "No Go files found in $PORTS_DIR."
    exit 0
fi

echo "Generating $total_files mocks..."

for file in "${files[@]}"; do
    ((current++))
    filename=$(basename "$file")
    mockname="mock_${filename}"

    # Generate the mock
    # -source: the file containing the interface(s)
    # -destination: where to save the generated mock
    # -package: the package name for the generated mock
    printf " [%d/%d] %s -> %s\n" "$current" "$total_files" "$filename" "$mockname"
    
    if ! $MOCKGEN -source="$file" -destination="$MOCKS_DEST/$mockname" -package=mocks; then
        echo "Error: Failed to generate mock for $file"
        exit 1
    fi
done

echo "Mocks generation completed successfully."
