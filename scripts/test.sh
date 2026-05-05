#!/bin/bash
# Usage: ./scripts/test.sh
go test -v -race -count=1 ./... "./internal/..."
