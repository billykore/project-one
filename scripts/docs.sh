#!/bin/bash
# Usage: ./scripts/docs.sh

set -euo pipefail
cd "$(dirname "$0")/.."

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "Error: 'swag' command not found." >&2
    echo "Please install swag CLI (e.g., 'go install github.com/swaggo/swag/cmd/swag@latest') and ensure it is in your PATH." >&2
    exit 1
fi

swag fmt
swag init -g cmd/main.go -o api/swagger
