#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE=(docker compose -f "$SCRIPT_DIR/docker-compose.local.yml" -f "$SCRIPT_DIR/docker-compose.ha.yml" --env-file "$SCRIPT_DIR/.env")

if [[ "${EUID}" -ne 0 ]]; then
  echo "Please run with sudo: sudo $0"
  exit 1
fi

wait_health() {
  local port="$1"
  local name="$2"
  local attempts=60

  for _ in $(seq 1 "$attempts"); do
    if curl -fsS "http://127.0.0.1:${port}/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done

  echo "Health check timeout: ${name} (127.0.0.1:${port})" >&2
  return 1
}

restart_one() {
  local service="$1"
  local port="$2"

  echo "[restart] ${service} ..."
  "${COMPOSE[@]}" restart "${service}"
  wait_health "${port}" "${service}"
  echo "[ok] ${service} healthy"
}

echo "[check] validating both instances are healthy before restart"
wait_health 1455 sub2api
wait_health 1456 sub2api-b

echo "[start] rolling restart (zero-downtime)"
restart_one sub2api-b 1456
restart_one sub2api 1455

echo "[done] rolling restart completed"
