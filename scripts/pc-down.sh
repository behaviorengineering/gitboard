#!/usr/bin/env bash
# Stop the Gitboard process-compose project only.
set -euo pipefail

GITBOARD_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SOCK="${PROCESS_COMPOSE_GITBOARD_SOCK:-$GITBOARD_DIR/.gitboard/process-compose.sock}"

if ! command -v process-compose >/dev/null 2>&1; then
  echo "process-compose not installed" >&2
  exit 1
fi

if [[ ! -S "$SOCK" ]]; then
  echo "Gitboard process-compose is not running ($SOCK)"
  exit 0
fi

exec process-compose down -U -u "$SOCK"
