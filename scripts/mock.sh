#!/bin/bash

# Check if mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo "Error: mockgen is not installed. Please install it using 'go install go.uber.org/mock/mockgen@latest'"
    exit 1
fi

echo "Generating mocks..."

PORTS_DIR="internal/core/ports"
MOCKS_DEST="internal/core/service/mocks"

if [ -d "$PORTS_DIR" ]; then
    mkdir -p "$MOCKS_DEST"
    for file in $PORTS_DIR/*.go; do
        filename=$(basename "$file")
        mockname="mock_${filename}"
        
        echo "Processing $file -> $MOCKS_DEST/$mockname"
        
        mockgen -source="$file" -destination="$MOCKS_DEST/$mockname" -package=mocks
    done
fi

echo "Mocks generation completed."
