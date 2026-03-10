#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   ./scripts/check-prevclose-consistency.sh 5817
#   ./scripts/check-prevclose-consistency.sh 5817,5020

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -f env.config ]]; then
  set -a
  # shellcheck disable=SC1091
  . ./env.config
  set +a
fi

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3312}"
DB_USER="${DB_USER:-jpxuser}"
DB_PASSWORD="${DB_PASSWORD:-jpxpass}"
DB_NAME="${DB_NAME:-jpx_data}"

TICKERS="${1:-5817}"

mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" <<SQL
WITH ranked AS (
  SELECT
    ticker,
    yymmdd,
    currentPrice,
    previousClose,
    pricemovement,
    ROW_NUMBER() OVER (PARTITION BY ticker ORDER BY yymmdd DESC) AS rn
  FROM watch_detail
  WHERE ticker IN ('${TICKERS//,/\',\'}')
),
latest2 AS (
  SELECT
    r1.ticker,
    r1.yymmdd AS latest_date,
    r1.currentPrice AS latest_price,
    r1.previousClose AS latest_prevclose,
    r1.pricemovement AS latest_change,
    r2.yymmdd AS prev_date,
    r2.currentPrice AS prev_price,
    r2.previousClose AS prev_prevclose
  FROM ranked r1
  LEFT JOIN ranked r2 ON r1.ticker = r2.ticker AND r2.rn = 2
  WHERE r1.rn = 1
)
SELECT
  wl.ticker,
  wl.currentPrice AS list_price,
  wl.previousClose AS list_prevclose,
  wl.pricemovement AS list_change,
  l.latest_date,
  l.latest_price,
  l.latest_prevclose,
  l.latest_change,
  l.prev_date,
  l.prev_price,
  CASE
    WHEN l.prev_price IS NULL THEN NULL
    WHEN wl.currentPrice IS NULL THEN NULL
    WHEN l.prev_price = 0 THEN NULL
    ELSE ROUND(((wl.currentPrice - l.prev_price) / l.prev_price) * 100, 2)
  END AS recomputed_change_from_latest2,
  CASE
    WHEN l.prev_price IS NULL THEN 'N/A(no 2nd row)'
    WHEN CAST(REPLACE(wl.previousClose, ',', '') AS DECIMAL(12,2)) = l.prev_price THEN 'MATCH'
    ELSE 'MISMATCH'
  END AS prevclose_vs_prev_price,
  CASE
    WHEN wl.currentPrice = l.latest_price THEN 'MATCH'
    ELSE 'MISMATCH'
  END AS list_price_vs_latest_price
FROM watch_list wl
LEFT JOIN latest2 l ON wl.ticker = l.ticker
WHERE wl.ticker IN ('${TICKERS//,/\',\'}')
ORDER BY wl.ticker;
SQL
