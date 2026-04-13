#!/bin/bash
# Usage: ./scripts/lint.sh [APP_NAME]
APP_NAME="${1:-greeting}"
golangci-lint run ./... "./internal/app/${APP_NAME}/..."
