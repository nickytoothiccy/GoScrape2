#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_ROOT="${ROOT_DIR}/.dev-env/go/current"

if [ ! -x "${GO_ROOT}/bin/go" ]; then
	echo "Local Go not found at ${GO_ROOT}" >&2
	echo "Run scripts/setup-local-go.sh first." >&2
	return 1 2>/dev/null || exit 1
fi

export GOROOT="${GO_ROOT}"
export GOPATH="${ROOT_DIR}/.dev-env/gopath"
export GOMODCACHE="${ROOT_DIR}/.dev-env/gomodcache"
export GOCACHE="${ROOT_DIR}/.dev-env/gocache"
export PATH="${GOROOT}/bin:${GOPATH}/bin:${PATH}"

echo "Using local Go at ${GOROOT}"
go version
