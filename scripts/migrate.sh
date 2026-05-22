#!/bin/bash

set -euo pipefail
cd "$(dirname "$0")/.."

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo "Error: 'migrate' command not found." >&2
    echo "Please install golang-migrate CLI (see https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) and ensure it is in your PATH." >&2
    exit 1
fi

MIGRATIONS_DIR="db/migrations"
COMMAND="${1:-}"
DSN="${2:-}"
NAME="${3:-}"

case "$COMMAND" in
    create)
        if [ -z "$NAME" ]; then
            echo "Error: Name/sequence is required for 'create' command." >&2
            exit 1
        fi
        migrate create -ext sql -dir "$MIGRATIONS_DIR" -seq "$NAME"
        ;;
    up)
        if [ -z "$DSN" ]; then
            echo "Error: DSN is required for 'up' command." >&2
            exit 1
        fi
        migrate -path "$MIGRATIONS_DIR" -database "$DSN" up ${NAME:+"$NAME"}
        ;;
    down)
        if [ -z "$DSN" ]; then
            echo "Error: DSN is required for 'down' command." >&2
            exit 1
        fi
        migrate -path "$MIGRATIONS_DIR" -database "$DSN" down ${NAME:+"$NAME"}
        ;;
    *)
        echo "Usage: $0 [create|up|down] [dsn] [name/steps]"
        exit 1
        ;;
esac