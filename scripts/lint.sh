#!/bin/bash
# Usage: ./scripts/lint.sh [APP_NAME]
APP_NAME=$1
golangci-lint run -c .golangci.yml ./... "./internal/app/${APP_NAME}/..."
