#!/usr/bin/env bash
set -euo pipefail

# === Config (override via env) ===
PGHOST="${PGHOST:-localhost}"
PGPORT="${PGPORT:-5432}"
PGUSER="${PGUSER:-edgar}"
PGPASSWORD="${PGPASSWORD:-postgres}"   # superuser password (current: postgres)
export PGPASSWORD

DB_NAME="${DB_NAME:-streaming_platform_db}"
APP_USER="${APP_USER:-streaming_platform_app}"
APP_PASSWORD="${APP_PASSWORD:-m4m4n1}"

ENABLE_UUID_EXTENSIONS="${ENABLE_UUID_EXTENSIONS:-1}"

psql_cmd=(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -v ON_ERROR_STOP=1 -w -d postgres)

echo "Using server $PGUSER@$PGHOST:$PGPORT"
echo "Ensuring role $APP_USER exists..."

role_exists="$("${psql_cmd[@]}" -tAc "SELECT 1 FROM pg_roles WHERE rolname='${APP_USER}'" || true)"
if [[ "$role_exists" != "1" ]]; then
  "${psql_cmd[@]}" -c "CREATE ROLE ${APP_USER} WITH LOGIN PASSWORD '${APP_PASSWORD}';"
  echo "Role ${APP_USER} created."
else
  "${psql_cmd[@]}" -c "ALTER ROLE ${APP_USER} WITH LOGIN PASSWORD '${APP_PASSWORD}';"
  echo "Role ${APP_USER} already exists; password updated."
fi

echo "Ensuring database $DB_NAME exists (owner: $APP_USER)..."
db_exists="$("${psql_cmd[@]}" -tAc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" || true)"
if [[ "$db_exists" != "1" ]]; then
  "${psql_cmd[@]}" -c "CREATE DATABASE ${DB_NAME} OWNER ${APP_USER} TEMPLATE template0 ENCODING 'UTF8';"
  echo "Database ${DB_NAME} created."
else
  echo "Database ${DB_NAME} already exists."
fi

db_psql_cmd=(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -v ON_ERROR_STOP=1 -w -d "$DB_NAME")

# Grant & schema ownership
"${db_psql_cmd[@]}" -c "GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${APP_USER};"
"${db_psql_cmd[@]}" -c "REVOKE ALL ON SCHEMA public FROM PUBLIC;"
"${db_psql_cmd[@]}" -c "GRANT USAGE, CREATE ON SCHEMA public TO ${APP_USER};"
"${db_psql_cmd[@]}" -c "ALTER SCHEMA public OWNER TO ${APP_USER};"

if [[ "$ENABLE_UUID_EXTENSIONS" == "1" ]]; then
  echo "Creating uuid extensions (if not present)..."
  "${db_psql_cmd[@]}" -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
  "${db_psql_cmd[@]}" -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
fi

echo "✅ Setup complete: DB=$DB_NAME, ROLE=$APP_USER"
