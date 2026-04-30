#!/bin/bash
# Usage: ./scripts/docs.sh [APP_NAME]
APP_NAME="${1:-user}"
swag init -g cmd/${APP_NAME}/main.go -o api/${APP_NAME} --parseDependency --parseInternal
