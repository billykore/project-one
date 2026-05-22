#!/bin/bash
# Usage: ./scripts/vet.sh

set -euo pipefail
cd "$(dirname "$0")/.."

go vet ./...
