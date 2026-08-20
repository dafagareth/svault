#!/bin/sh
# Copyright 2026 Dafa. MIT License.
#
# Install script for svault, the local encrypted secret vault.
#
#   curl -fsSL https://raw.githubusercontent.com/dafagareth/svault/main/install.sh | sh
#
# Environment overrides:
#   SVAULT_VERSION   install a specific version (default: latest release)
#   SVAULT_BIN_DIR   install directory (default: /usr/local/bin, or ~/.local/bin
#                   when /usr/local/bin is not writable)

set -eu

REPO="dafagareth/svault"

info() { printf '%s\n' "$*"; }
err() { printf 'error: %s\n' "$*" >&2; exit 1; }

# Detect operating system.
os=$(uname -s)
case "$os" in
	Linux) os="linux" ;;
	Darwin) os="darwin" ;;
	*) err "unsupported OS: $os (use install.ps1 on Windows)" ;;
esac

# Detect architecture.
arch=$(uname -m)
case "$arch" in
	x86_64 | amd64) arch="amd64" ;;
	aarch64 | arm64) arch="arm64" ;;
	*) err "unsupported architecture: $arch" ;;
esac

asset="svault-${os}-${arch}"

# Resolve the version to install.
version="${SVAULT_VERSION:-}"
if [ -z "$version" ]; then
	info "Resolving latest release..."
	version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
		| grep '"tag_name":' \
		| head -1 \
		| sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
	[ -n "$version" ] || err "could not resolve latest version"
fi
info "Installing svault ${version} (${os}/${arch})"

base="https://github.com/${REPO}/releases/download/${version}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# Download the binary and checksums.
info "Downloading ${asset}..."
curl -fsSL "${base}/${asset}" -o "${tmp}/svault" || err "download failed for ${asset}"
curl -fsSL "${base}/checksums.txt" -o "${tmp}/checksums.txt" || err "download failed for checksums.txt"

# Verify the checksum.
expected=$(grep " ${asset}\$" "${tmp}/checksums.txt" | awk '{print $1}')
if [ -n "$expected" ]; then
	if command -v sha256sum >/dev/null 2>&1; then
		actual=$(sha256sum "${tmp}/svault" | awk '{print $1}')
	elif command -v shasum >/dev/null 2>&1; then
		actual=$(shasum -a 256 "${tmp}/svault" | awk '{print $1}')
	else
		actual=""
	fi
	if [ -n "$actual" ] && [ "$actual" != "$expected" ]; then
		err "checksum mismatch: expected ${expected}, got ${actual}"
	fi
	info "Checksum verified."
else
	info "Warning: no checksum found for ${asset}, skipping verification."
fi

chmod +x "${tmp}/svault"

# Pick an install directory.
bin_dir="${SVAULT_BIN_DIR:-/usr/local/bin}"
if [ ! -d "$bin_dir" ] || [ ! -w "$bin_dir" ]; then
	if [ "$bin_dir" = "/usr/local/bin" ]; then
		bin_dir="${HOME}/.local/bin"
		mkdir -p "$bin_dir"
		info "No write access to /usr/local/bin, installing to ${bin_dir}"
	else
		err "install directory not writable: ${bin_dir}"
	fi
fi

mv "${tmp}/svault" "${bin_dir}/svault"
info "Installed to ${bin_dir}/svault"

# Remind the user if the directory is not on PATH.
case ":${PATH}:" in
	*":${bin_dir}:"*) ;;
	*) info "Note: ${bin_dir} is not on your PATH. Add it to your shell profile." ;;
esac

info "Run 'svault init' to get started."
