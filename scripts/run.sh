#!/bin/bash
# Usage: ./scripts/run.sh [BUILD_DIR]

set -euo pipefail
cd "$(dirname "$0")/.."

BUILD_DIR="./bin"
if [ $# -gt 0 ] && [[ "$1" != -* ]]; then
  BUILD_DIR="$1"
  shift
fi

"${BUILD_DIR}/main" "$@"
