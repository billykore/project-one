#!/bin/bash
# Usage: ./scripts/test.sh

set -euo pipefail
cd "$(dirname "$0")/.."

go test -v -race -count=1 ./... "./internal/..."
