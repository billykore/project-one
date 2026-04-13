#!/bin/bash
# Usage: ./scripts/build.sh [APP_NAME] [BUILD_DIR]
APP_NAME="${1:-greeting}"
BUILD_DIR="${2:-./bin}"
go build -mod=mod -o "${BUILD_DIR}/${APP_NAME}" "./cmd/${APP_NAME}"
