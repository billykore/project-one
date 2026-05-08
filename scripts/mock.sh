#!/bin/bash

# Check if mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo "Error: mockgen is not installed. Please install it using 'go install go.uber.org/mock/mockgen@latest'"
    exit 1
fi

echo "Generating mocks..."

PORTS_DIR="internal/core/ports"
SERVICE_MOCKS_DEST="internal/core/service/mocks"
USECASE_MOCKS_DEST="internal/core/usecase/mocks"

if [ -d "$PORTS_DIR" ]; then
    mkdir -p "$SERVICE_MOCKS_DEST"
    mkdir -p "$USECASE_MOCKS_DEST"
    for file in $PORTS_DIR/*.go; do
        filename=$(basename "$file")
        mockname="mock_${filename}"
        
        echo "Processing $file -> $SERVICE_MOCKS_DEST/$mockname"
        mockgen -source="$file" -destination="$SERVICE_MOCKS_DEST/$mockname" -package=mocks

        echo "Processing $file -> $USECASE_MOCKS_DEST/$mockname"
        mockgen -source="$file" -destination="$USECASE_MOCKS_DEST/$mockname" -package=mocks
    done
fi

echo "Mocks generation completed."
