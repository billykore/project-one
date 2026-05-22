#!/bin/bash
# Usage: ./scripts/lint.sh

set -euo pipefail
cd "$(dirname "$0")/.."

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "Error: 'golangci-lint' command not found." >&2
    echo "Please install golangci-lint (see https://golangci-lint.run/welcome/install/) and ensure it is in your PATH." >&2
    exit 1
fi

golangci-lint run -c .golangci.yml ./...
