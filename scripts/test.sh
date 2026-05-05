#!/bin/bash
# Usage: ./scripts/test.sh [APP_NAME]
APP_NAME=$1
go test -v -race -count=1 ./... "./internal/app/${APP_NAME}/..."
