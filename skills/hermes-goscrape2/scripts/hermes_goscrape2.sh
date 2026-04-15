#!/usr/bin/env bash
set -euo pipefail

BIN_PATH="${HERMES_GOSCRAPE_BIN:-/opt/hermes/bin/hermes-goscrape2}"

if [ ! -x "$BIN_PATH" ]; then
  echo "{\"error\":\"missing hermes-goscrape2 binary at $BIN_PATH\"}" >&2
  exit 1
fi

exec "$BIN_PATH" "$@"
