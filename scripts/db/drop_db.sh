#!/usr/bin/env bash
set -euo pipefail

# === Config (override via env) ===
PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-postgres}"
export PGPASSWORD

DB_NAME="${DB_NAME:-streaming_platform_db}"
APP_USER="${APP_USER:-streaming_platform_app}"

CONFIRM="${1:-}"

if [[ "$CONFIRM" != "-y" && "$CONFIRM" != "--yes" ]]; then
  echo "⚠️  This will DROP database '$DB_NAME' and role '$APP_USER'."
  echo "    Re-run with -y to proceed: $0 -y"
  exit 1
fi

psql_cmd=(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -v ON_ERROR_STOP=1 -w -d postgres)

echo "Terminating active connections to $DB_NAME..."
"${psql_cmd[@]}" -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${DB_NAME}' AND pid <> pg_backend_pid();"

echo "Dropping database (if exists)..."
"${psql_cmd[@]}" -c "DROP DATABASE IF EXISTS ${DB_NAME};"

echo "Dropping role (if exists)..."
"${psql_cmd[@]}" -c "DROP ROLE IF EXISTS ${APP_USER};"

echo "🗑️  Drop complete."
