#!/bin/sh
set -eu
version=__AISTD_VERSION__
base=${AISTD_RELEASE_BASE:-__AISTD_RELEASE_BASE__}
base=${base%/}
case "$base" in
  https://*|http://localhost|http://localhost:[0-9]*|http://127.0.0.1|http://127.0.0.1:[0-9]*) ;;
  *) echo 'Release base must use HTTPS, or localhost HTTP for local deployment' >&2; exit 1;;
esac
case $(uname -s) in Darwin) os=darwin;; Linux) os=linux;; *) echo 'Supported systems: macOS and Linux' >&2; exit 1;; esac
case $(uname -m) in x86_64|amd64) arch=amd64;; arm64|aarch64) arch=arm64;; *) echo 'Unsupported CPU architecture' >&2; exit 1;; esac
name=aistd_${version}_${os}_${arch}
tmp=$(mktemp -d)
stage=
trap 'rm -rf "$tmp"; [ -z "$stage" ] || rm -rf "$stage"' EXIT
trap 'exit 1' HUP INT TERM
download() {
  case "$base" in
    http://localhost*|http://127.0.0.1*) curl --fail --silent --show-error --location "$1" -o "$2" ;;
    *) curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' "$1" -o "$2" ;;
  esac
}
download "$base/v$version/$name" "$tmp/$name"
download "$base/v$version/SHA256SUMS" "$tmp/SHA256SUMS"
expected=$(awk -v f="$name" '$2 == f {print $1}' "$tmp/SHA256SUMS")
[ ${#expected} -eq 64 ] || { echo 'Missing or invalid checksum' >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then actual=$(sha256sum "$tmp/$name" | awk '{print $1}'); else actual=$(shasum -a 256 "$tmp/$name" | awk '{print $1}'); fi
[ "$actual" = "$expected" ] || { echo 'Checksum mismatch' >&2; exit 1; }
install_dir=${AISTD_INSTALL_DIR:-"$HOME/.local/bin"}
mkdir -p "$install_dir"
[ ! -L "$install_dir/aistd" ] && { [ ! -e "$install_dir/aistd" ] || [ -f "$install_dir/aistd" ]; } || { echo 'Refusing to replace a non-regular aistd file' >&2; exit 1; }
chmod 755 "$tmp/$name"
"$tmp/$name" version
# Stage on the destination filesystem so replacement preserves the old CLI until rename succeeds.
stage=$(mktemp -d "$install_dir/.aistd-install.XXXXXX")
mv "$tmp/$name" "$stage/aistd"
[ ! -L "$install_dir/aistd" ] && { [ ! -e "$install_dir/aistd" ] || [ -f "$install_dir/aistd" ]; } || { echo 'Refusing to replace a non-regular aistd file' >&2; exit 1; }
mv -f "$stage/aistd" "$install_dir/aistd"
printf 'Installed: %s/aistd\n' "$install_dir"
printf 'If needed, add this directory to PATH: %s\n' "$install_dir"
