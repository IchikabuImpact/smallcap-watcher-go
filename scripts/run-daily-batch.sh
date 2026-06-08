#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="${1:-$(cd "$(dirname "$0")/.." && pwd)}"
cd "${REPO_DIR}"

if [[ ! -f env.config ]]; then
  echo "ERROR: env.config not found in ${REPO_DIR}" >&2
  exit 1
fi

# shellcheck disable=SC1091
set -a
source env.config
set +a

if [[ ! -d node_modules ]]; then
  echo "ERROR: node_modules not found. Run 'npm ci --omit=dev' in ${REPO_DIR} first." >&2
  exit 1
fi

node ./bin/smallcap-watcher.js --batch --gen

# 生成物の簡易ヘルスチェック（index/detail の鮮度整合）
./scripts/check-index-freshness.sh "${OUTPUT_DIR:-public}"
