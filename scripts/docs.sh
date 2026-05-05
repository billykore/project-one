#!/bin/bash
# Usage: ./scripts/docs.sh
swag init -g cmd/main.go -o api/swagger --parseDependency --parseInternal
