#!/bin/bash
# Usage: ./scripts/clean.sh [BUILD_DIR]
BUILD_DIR="${1:-./bin}"
rm -rf "${BUILD_DIR}"
