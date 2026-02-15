#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COMPOSE_ENV=(env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD)
COMPOSE_CMD=(docker compose)
if [[ -f env.config ]]; then
  COMPOSE_CMD+=(--env-file env.config)
fi

if [[ -f env.config ]]; then
  set -a
  # shellcheck disable=SC1091
  source env.config
  set +a
fi

DB_NAME="${DB_NAME:-jpx_data}"
DB_USER="${DB_USER:-jpx_user}"
DB_PASSWORD="${DB_PASSWORD:-jpx_password}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-rootpassword}"

SQL=$(cat <<EOSQL
CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\`;
CREATE USER IF NOT EXISTS '${DB_USER}'@'%' IDENTIFIED BY '${DB_PASSWORD}';
ALTER USER '${DB_USER}'@'%' IDENTIFIED BY '${DB_PASSWORD}';
GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'%';
FLUSH PRIVILEGES;
EOSQL
)

echo "[INFO] Reconfiguring MySQL user '${DB_USER}' for database '${DB_NAME}'..."
"${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" up -d mysql >/dev/null
"${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" exec -T mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "$SQL"
echo "[OK] Authentication settings have been repaired."
