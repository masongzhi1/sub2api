#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="${SCRIPT_DIR}/deploy"
ENV_FILE="${DEPLOY_DIR}/.env"

COMPOSE=(
  docker compose
  -f "${DEPLOY_DIR}/docker-compose.local.yml"
  -f "${DEPLOY_DIR}/docker-compose.ha.yml"
  -f "${DEPLOY_DIR}/docker-compose.build-local.yml"
  --env-file "${ENV_FILE}"
)

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "[error] missing required command: $1" >&2
    exit 1
  fi
}

http_probe() {
  local url="$1"

  if command -v curl >/dev/null 2>&1; then
    curl -fsS "$url" >/dev/null
    return
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -q -T 5 -O /dev/null "$url"
    return
  fi

  echo "[error] curl or wget is required for health checks" >&2
  exit 1
}

wait_health() {
  local url="$1"
  local name="$2"
  local attempts="${3:-60}"

  for _ in $(seq 1 "$attempts"); do
    if http_probe "$url"; then
      echo "[ok] ${name} healthy"
      return 0
    fi
    sleep 2
  done

  echo "[error] health check timeout: ${name} (${url})" >&2
  return 1
}

require_command docker

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "[error] env file not found: ${ENV_FILE}" >&2
  echo "[hint] run: cp \"${DEPLOY_DIR}/.env.example\" \"${ENV_FILE}\"" >&2
  exit 1
fi

echo "[step] building local image (sequential build to avoid OOM)"
COMPOSE_PARALLEL_LIMIT=1 "${COMPOSE[@]}" build sub2api

echo "[step] recreating services with local image"
"${COMPOSE[@]}" up -d

echo "[step] waiting for local health checks"
wait_health "http://127.0.0.1:8080/health" "sub2api"
wait_health "http://127.0.0.1:1456/health" "sub2api-b"

echo "[step] current compose status"
"${COMPOSE[@]}" ps

echo "[done] rebuild and local-image restart completed"
