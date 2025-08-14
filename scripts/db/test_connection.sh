#!/usr/bin/env bash
set -euo pipefail

PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-postgres}"
export PGPASSWORD

DB_NAME="${DB_NAME:-streaming_platform_db}"
APP_USER="${APP_USER:-streaming_platform_app}"

psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -w -c "SELECT current_database() AS db, current_user AS user, version() AS version;"
echo
echo "Checking extensions:"
psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -w -c "SELECT extname FROM pg_extension ORDER BY 1;"
