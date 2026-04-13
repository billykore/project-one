#!/bin/bash
# Usage: ./scripts/vet.sh [APP_NAME]
APP_NAME="${1:-greeting}"
go vet ./... "./internal/app/${APP_NAME}/..."
