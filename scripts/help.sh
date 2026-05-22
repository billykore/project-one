#!/bin/bash

set -euo pipefail
cd "$(dirname "$0")/.."

grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
