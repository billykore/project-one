#!/bin/bash
# Usage: ./scripts/run.sh [APP_NAME] [BUILD_DIR]
APP_NAME="${1:-greeting}"
BUILD_DIR="${2:-./bin}"
"${BUILD_DIR}/${APP_NAME}"
