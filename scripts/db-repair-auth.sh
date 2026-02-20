#!/usr/bin/env bash
set -euo pipefail

COMPOSE_CMD=(docker compose)
COMPOSE_ENV=(env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD)

if [[ -f "env.config" ]]; then
  # shellcheck disable=SC1091
  source env.config
fi

DB_USER="${DB_USER:-jpx}"
DB_PASSWORD="${DB_PASSWORD:-jpxpass}"
DB_NAME="${DB_NAME:-jpx_data}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-rootpass}"

SQL=$(cat <<SQL_EOF
CREATE DATABASE IF NOT EXISTS \
\`${DB_NAME}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER IF NOT EXISTS '${DB_USER}'@'%' IDENTIFIED BY '${DB_PASSWORD}';
ALTER USER '${DB_USER}'@'%' IDENTIFIED BY '${DB_PASSWORD}';
GRANT ALL PRIVILEGES ON \
\`${DB_NAME}\`.* TO '${DB_USER}'@'%';
FLUSH PRIVILEGES;
SQL_EOF
)

echo "[INFO] Reconfiguring MySQL user '${DB_USER}' for database '${DB_NAME}'..."
"${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" --env-file env.config up -d mysql >/dev/null

echo "[INFO] Waiting for MySQL to become ready..."
ready=0
for i in $(seq 1 30); do
  if "${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" --env-file env.config exec -T mysql \
       mysqladmin ping -uroot -p"${MYSQL_ROOT_PASSWORD}" --silent >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done

if [[ "$ready" -ne 1 ]]; then
  echo "[ERROR] MySQL did not become ready in time." >&2
  exit 1
fi

"${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" --env-file env.config exec -T mysql \
  mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "$SQL"

echo "[OK] Authentication settings have been repaired."
