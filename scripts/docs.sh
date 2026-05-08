#!/bin/bash
# Usage: ./scripts/docs.sh
swag fmt
swag init  -ot go,yaml -g cmd/main.go -o api/swagger --parseDependency --parseInternal
