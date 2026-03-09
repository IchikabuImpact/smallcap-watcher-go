#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="${1:-$(cd "$(dirname "$0")/.." && pwd)}"
cd "${REPO_DIR}"

if [[ ! -f env.config ]]; then
  echo "ERROR: env.config not found in ${REPO_DIR}" >&2
  exit 1
fi

# cron環境で export 済みのDB_* が優先される事故を防ぐ
env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD \
  /usr/bin/docker compose --env-file env.config run --rm app --batch --gen

# 生成物の簡易ヘルスチェック（index/detail の鮮度整合）
./scripts/check-index-freshness.sh output
