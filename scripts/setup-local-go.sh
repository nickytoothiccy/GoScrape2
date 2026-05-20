#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_VERSION="${1:-1.24.1}"
TOOLS_DIR="${ROOT_DIR}/.dev-env"
INSTALL_DIR="${TOOLS_DIR}/go/${GO_VERSION}"
CURRENT_LINK="${TOOLS_DIR}/go/current"

case "$(uname -s)" in
Linux) GOOS="linux" ;;
Darwin) GOOS="darwin" ;;
*)
	echo "Unsupported OS: $(uname -s)" >&2
	exit 1
	;;
esac

case "$(uname -m)" in
x86_64|amd64) GOARCH="amd64" ;;
aarch64|arm64) GOARCH="arm64" ;;
*)
	echo "Unsupported architecture: $(uname -m)" >&2
	exit 1
	;;
esac

TARBALL="go${GO_VERSION}.${GOOS}-${GOARCH}.tar.gz"
TMP_DIR="$(mktemp -d)"
ARCHIVE_PATH="${TMP_DIR}/${TARBALL}"

cleanup() {
	rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

mkdir -p "${TOOLS_DIR}/go"

if [ -x "${INSTALL_DIR}/bin/go" ]; then
	ln -sfn "${INSTALL_DIR}" "${CURRENT_LINK}"
	echo "Go ${GO_VERSION} already installed at ${INSTALL_DIR}"
	exit 0
fi

echo "Downloading ${TARBALL}..."
if command -v curl >/dev/null 2>&1; then
	curl -fsSL "https://go.dev/dl/${TARBALL}" -o "${ARCHIVE_PATH}"
elif command -v wget >/dev/null 2>&1; then
	wget -qO "${ARCHIVE_PATH}" "https://go.dev/dl/${TARBALL}"
else
	echo "curl or wget is required to download Go" >&2
	exit 1
fi

mkdir -p "${INSTALL_DIR}"
tar -xzf "${ARCHIVE_PATH}" -C "${TMP_DIR}"
mv "${TMP_DIR}/go/"* "${INSTALL_DIR}/"
rmdir "${TMP_DIR}/go"
ln -sfn "${INSTALL_DIR}" "${CURRENT_LINK}"

echo "Installed Go ${GO_VERSION} to ${INSTALL_DIR}"
echo "Run: source scripts/use-local-go.sh"
