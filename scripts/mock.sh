#!/bin/bash

# Define apps to generate mocks for
APPS=("user" "post")

# Check if mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo "Error: mockgen is not installed. Please install it using 'go install go.uber.org/mock/mockgen@latest'"
    exit 1
fi

echo "Generating mocks..."

for app in "${APPS[@]}"; do
    PORTS_DIR="internal/app/$app/core/ports"
    MOCKS_DEST="internal/app/$app/core/service/mocks"

    if [ -d "$PORTS_DIR" ]; then
        mkdir -p "$MOCKS_DEST"
        for file in $PORTS_DIR/*.go; do
            filename=$(basename "$file")
            mockname="mock_${filename}"
            
            echo "Processing $file -> $MOCKS_DEST/$mockname"
            
            mockgen -source="$file" -destination="$MOCKS_DEST/$mockname" -package=mocks
        done
    fi
done

echo "Mocks generation completed."
