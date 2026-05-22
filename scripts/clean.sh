#!/bin/bash
# Usage: ./scripts/clean.sh [BUILD_DIR]

set -euo pipefail
cd "$(dirname "$0")/.."

BUILD_DIR="${1:-./bin}"
rm -rf "${BUILD_DIR}"
