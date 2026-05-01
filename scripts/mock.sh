#!/bin/bash

# Define the ports directory and mocks destination
PORTS_DIR="internal/app/user/core/ports"
MOCKS_DEST="internal/app/user/core/service/mocks"

# Check if mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo "Error: mockgen is not installed. Please install it using 'go install go.uber.org/mock/mockgen@latest'"
    exit 1
fi

echo "Generating mocks..."

# Generate mocks for each file in ports directory
for file in $PORTS_DIR/*.go; do
    filename=$(basename "$file")
    mockname="mock_${filename}"
    
    echo "Processing $file -> $MOCKS_DEST/$mockname"
    
    mockgen -source="$file" -destination="$MOCKS_DEST/$mockname" -package=mocks
done

echo "Mocks generation completed."
