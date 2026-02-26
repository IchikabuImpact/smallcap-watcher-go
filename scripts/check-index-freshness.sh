#!/usr/bin/env bash
set -euo pipefail

OUTPUT_DIR="${1:-output}"
MAX_AGE_SECONDS="${INDEX_MAX_AGE_SECONDS:-172800}"

index_file="${OUTPUT_DIR}/index.html"
detail_dir="${OUTPUT_DIR}/detail"

if [[ ! -f "${index_file}" ]]; then
  echo "ERROR: index not found: ${index_file}" >&2
  exit 1
fi

if [[ ! -d "${detail_dir}" ]]; then
  echo "ERROR: detail dir not found: ${detail_dir}" >&2
  exit 1
fi

latest_detail="$(find "${detail_dir}" -maxdepth 1 -type f -name '*.html' -printf '%T@ %p\n' | sort -nr | head -n1 | cut -d' ' -f2-)"
if [[ -z "${latest_detail}" ]]; then
  echo "ERROR: no detail html files in ${detail_dir}" >&2
  exit 1
fi

index_epoch="$(stat -c '%Y' "${index_file}")"
detail_epoch="$(stat -c '%Y' "${latest_detail}")"
index_size="$(stat -c '%s' "${index_file}")"
now_epoch="$(date +%s)"
index_age="$((now_epoch - index_epoch))"

echo "index_path=${index_file} index_mtime=$(date -u -d @${index_epoch} +%FT%TZ) index_size=${index_size} index_age_seconds=${index_age}"
echo "detail_path=${latest_detail} detail_mtime=$(date -u -d @${detail_epoch} +%FT%TZ)"

if (( index_size <= 0 )); then
  echo "ERROR: index.html is empty" >&2
  exit 1
fi

if (( index_age > MAX_AGE_SECONDS )); then
  echo "ERROR: index is too old (age=${index_age}s, max=${MAX_AGE_SECONDS}s)" >&2
  exit 1
fi

if (( index_epoch + 60 < detail_epoch )); then
  echo "ERROR: index is older than newest detail by more than 60s" >&2
  exit 1
fi
