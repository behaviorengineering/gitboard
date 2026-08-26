#!/usr/bin/env bash
# Start Gitboard via process-compose (standalone stack).
# Rebuilds and restarts on .go / static changes (process-compose watch).
# Disable: PC_NO_WATCH=1 make serve
set -euo pipefail

GITBOARD_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$GITBOARD_DIR"

if ! command -v process-compose >/dev/null 2>&1; then
  echo "process-compose not found; install: https://f1bonacc1.github.io/process-compose/installation/" >&2
  echo "  macOS: brew install process-compose" >&2
  exit 1
fi

export GITBOARD_DIR
export GITBOARD_CONFIG="${GITBOARD_CONFIG:-${HOME}/.config/gitboard/config.yaml}"
export GITBOARD_ADDR="${GITBOARD_ADDR:-127.0.0.1:1325}"

if [[ ! -x "$GITBOARD_DIR/bin/gitboard" ]]; then
  echo "building gitboard…"
  mkdir -p "$GITBOARD_DIR/bin"
  go build -o "$GITBOARD_DIR/bin/gitboard" ./cmd/gitboard
fi
export GITBOARD_BIN="$GITBOARD_DIR/bin/gitboard"

if [[ ! -f "$GITBOARD_CONFIG" ]]; then
  echo "missing config: $GITBOARD_CONFIG — creating via init" >&2
  "$GITBOARD_BIN" init -config "$GITBOARD_CONFIG"
fi

chmod +x "$GITBOARD_DIR/scripts/"*.sh 2>/dev/null || true

mkdir -p "$GITBOARD_DIR/.gitboard"
SOCK="${PROCESS_COMPOSE_GITBOARD_SOCK:-$GITBOARD_DIR/.gitboard/process-compose.sock}"
export PROCESS_COMPOSE_GITBOARD_SOCK="$SOCK"

echo "Gitboard: http://${GITBOARD_ADDR}/  config: ${GITBOARD_CONFIG}"
echo "Watch: rebuild/restart on .go and static UI changes (PC_NO_WATCH=1 to disable)"
echo "process-compose TUI (this project only); 0 quit."

exec process-compose up \
  --config "$GITBOARD_DIR/process-compose.yaml" \
  --shortcuts "$GITBOARD_DIR/process-compose-shortcuts.yaml" \
  -U -u "$SOCK" \
  "$@"
