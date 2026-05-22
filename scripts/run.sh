#!/bin/bash
# Usage: ./scripts/run.sh [BUILD_DIR]

set -euo pipefail
cd "$(dirname "$0")/.."

BUILD_DIR="${1:-./bin}"
"${BUILD_DIR}/main"
