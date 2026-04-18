#!/bin/bash

MIGRATIONS_DIR="db/migrations"
COMMAND=$1
DSN=$2
NAME=$3

case $COMMAND in
    create)
        migrate create -ext sql -dir "$MIGRATIONS_DIR" -seq "$NAME"
        ;;
    up)
        migrate -path "$MIGRATIONS_DIR" -database "$DSN" up $NAME
        ;;
    down)
        migrate -path "$MIGRATIONS_DIR" -database "$DSN" down $NAME
        ;;
    *)
        echo "Usage: $0 [create|up|down] [dsn] [name/steps]"
        exit 1
        ;;
esac