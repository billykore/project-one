#!/bin/bash
# Usage: ./scripts/docs.sh
swag fmt
swag init -g cmd/main.go -o api/swagger --parseDependency --parseInternal
