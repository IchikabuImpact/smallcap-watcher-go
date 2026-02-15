echo "[INFO] Reconfiguring MySQL user '${DB_USER}' for database '${DB_NAME}'..."
"${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" up -d mysql >/dev/null

echo "[INFO] Waiting for MySQL to become ready..."
ready=0
for i in $(seq 1 30); do
  if "${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" exec -T mysql \
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

"${COMPOSE_ENV[@]}" "${COMPOSE_CMD[@]}" exec -T mysql \
  mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "$SQL"

echo "[OK] Authentication settings have been repaired."
