#!/bin/sh
# POSIX shell installer for svault.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/dafagareth/svault/main/install.sh | sh
#
# Environment variables:
#   SVAULT_VERSION   Target version to install (default: latest release)
#   SVAULT_BIN_DIR   Target installation directory (default: /usr/local/bin or ~/.local/bin)

set -eu

REPO="dafagareth/svault"

info() { printf '[info] %s\n' "$*"; }
err() { printf '[error] %s\n' "$*" >&2; exit 1; }

download() {
	url="$1"
	out="$2"
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$out"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$out" "$url"
	else
		err "neither curl nor wget found in PATH"
	fi
}

os=$(uname -s)
case "$os" in
	Linux) os="linux" ;;
	Darwin) os="darwin" ;;
	*) err "unsupported OS: $os (use install.ps1 on Windows)" ;;
esac

arch=$(uname -m)
case "$arch" in
	x86_64 | amd64) arch="amd64" ;;
	aarch64 | arm64) arch="arm64" ;;
	*) err "unsupported architecture: $arch" ;;
esac

asset="svault-${os}-${arch}"

version="${SVAULT_VERSION:-}"
if [ -z "$version" ]; then
	info "Resolving latest release tag..."
	tmp_json=$(mktemp)
	download "https://api.github.com/repos/${REPO}/releases/latest" "$tmp_json" || err "failed to query GitHub API"
	version=$(grep '"tag_name":' "$tmp_json" | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
	rm -f "$tmp_json"
	[ -n "$version" ] || err "could not resolve latest release version"
fi

info "Installing svault ${version} (${os}/${arch})"

base="https://github.com/${REPO}/releases/download/${version}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

info "Downloading ${asset}..."
download "${base}/${asset}" "${tmp}/svault" || err "download failed for ${asset}"
download "${base}/checksums.txt" "${tmp}/checksums.txt" || err "download failed for checksums.txt"

expected=$(grep " ${asset}\$" "${tmp}/checksums.txt" | awk '{print $1}')
if [ -n "$expected" ]; then
	actual=""
	if command -v sha256sum >/dev/null 2>&1; then
		actual=$(sha256sum "${tmp}/svault" | awk '{print $1}')
	elif command -v shasum >/dev/null 2>&1; then
		actual=$(shasum -a 256 "${tmp}/svault" | awk '{print $1}')
	elif command -v openssl >/dev/null 2>&1; then
		actual=$(openssl dgst -sha256 "${tmp}/svault" | awk '{print $2}')
	fi

	if [ -n "$actual" ] && [ "$actual" != "$expected" ]; then
		err "checksum verification mismatch: expected ${expected}, got ${actual}"
	fi
	info "Checksum verification successful."
else
	info "Warning: checksum for ${asset} not found, skipping verification."
fi

chmod +x "${tmp}/svault"

bin_dir="${SVAULT_BIN_DIR:-/usr/local/bin}"
if [ ! -d "$bin_dir" ] || [ ! -w "$bin_dir" ]; then
	if [ "$bin_dir" = "/usr/local/bin" ]; then
		bin_dir="${HOME}/.local/bin"
		mkdir -p "$bin_dir"
		info "No write access to /usr/local/bin, installing to ${bin_dir}"
	else
		err "installation directory not writable: ${bin_dir}"
	fi
fi

mv "${tmp}/svault" "${bin_dir}/svault"
info "Successfully installed binary to ${bin_dir}/svault"

case ":${PATH}:" in
	*":${bin_dir}:"*) ;;
	*) info "Notice: ${bin_dir} is not currently in your PATH environment variable." ;;
esac

info "Run 'svault init' to initialize your vault."
