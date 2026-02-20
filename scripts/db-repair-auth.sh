#!/usr/bin/env bash
set -euo pipefail

COMPOSE_CMD=(docker compose)
COMPOSE_ENV=(env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD)
COMPOSE_FILE_ARGS=()

if [[ -f "env.config" ]]; then
  # shellcheck disable=SC1091
  source env.config
  COMPOSE_FILE_ARGS=(--env-file env.config)
fi

# Keep defaults aligned with docker-compose.yml / env.config.sample
DB_USER="${DB_USER:-jpx_user}"
DB_PASSWORD="${DB_PASSWORD:-jpx_password}"
DB_NAME="${DB_NAME:-jpx_data}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-rootpassword}"
LEGACY_DB_USER="${LEGACY_DB_USER:-jpx}"

EXTRA_SQL=""
if [[ "${LEGACY_DB_USER}" != "${DB_USER}" ]]; then
  EXTRA_SQL=$(cat <<SQL_EOF
CREATE USER IF NOT EXISTS '${LEGACY_DB_USER}'@'%' IDENTIFIED BY '${DB_PASSWORD}';
ALTER USER '${LEGACY_DB_USER}'@'%' IDENTIFIED BY '${DB_PASSWORD}';
GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${LEGACY_DB_USER}'@'%';
SQL_EOF
)
fi

SQL=$(cat <<SQL_EOF
CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER IF NOT EXISTS '${DB_USER}'@'%' IDENTIFIED BY '${DB_PASSWORD}';
ALTER USER '${DB_USER}'@'%' IDENTIFIED BY '${DB_PASSWORD}';
GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'%';
${EXTRA_SQL}
FLUSH PRIVILEGES;
SQL_EOF
)

echo "[INFO] Reconfiguring MySQL user '${DB_USER}' for database '${DB_NAME}'..."
if [[ "${LEGACY_DB_USER}" != "${DB_USER}" ]]; then
  echo "[INFO] Also updating compatibility user '${LEGACY_DB_USER}' (same password)."
fi
"${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" "${COMPOSE_FILE_ARGS[@]}" up -d mysql >/dev/null

echo "[INFO] Waiting for MySQL to become ready..."
ready=0
for i in $(seq 1 30); do
  if "${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" "${COMPOSE_FILE_ARGS[@]}" exec -T mysql \
       mysqladmin ping -uroot -p"${MYSQL_ROOT_PASSWORD}" --silent >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done

if [[ "$ready" -ne 1 ]]; then
  echo "[ERROR] MySQL did not become ready in time with MYSQL_ROOT_PASSWORD from env.config/defaults." >&2
  echo "[ERROR] Confirm MYSQL_ROOT_PASSWORD in env.config matches the existing mysql-data volume initialization value." >&2
  exit 1
fi

"${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" "${COMPOSE_FILE_ARGS[@]}" exec -T mysql \
  mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "$SQL"

echo "[OK] Authentication settings have been repaired."
