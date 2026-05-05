#!/bin/bash
# Usage: ./scripts/build.sh [BUILD_DIR]
BUILD_DIR="${1:-./bin}"
go build -mod=mod -o "${BUILD_DIR}/main" "./cmd/main.go"
echo "Build completed. Binary is located at ${BUILD_DIR}/main"
