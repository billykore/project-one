#!/bin/sh
set -eu

# ponytail: initializes fresh volumes; use versioned migrations for upgrades.
for migration in /migrations/*.up.sql; do
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$migration"
done
