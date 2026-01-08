#!/bin/sh
set -e

mkdir -p /app/output /app/output/detail /app/output/static
chmod -R 0777 /app/output || true

exec "$@"
