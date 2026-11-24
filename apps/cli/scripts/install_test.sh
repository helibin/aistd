#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
release_version=$(cat VERSION)
export release_version
fixture_dir=$PWD/release/v$release_version
[ -f "$fixture_dir/SHA256SUMS" ] || { echo 'Build release artifacts first' >&2; exit 1; }
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT HUP INT TERM
mkdir "$work/bin" "$work/install" "$work/local"
export fixture_dir
export AISTD_INSTALL_DIR="$work/install"
export AISTD_RELEASE_BASE=https://example.invalid/release
export PATH="$work/bin:$PATH"
cat > "$work/bin/curl" <<'MOCK'
#!/bin/sh
set -eu
url=
output=
https_only=false
while [ "$#" -gt 0 ]; do
  case "$1" in
    https://*|http://*) url=$1 ;;
    -o) shift; output=$1 ;;
    --proto|--proto-redir) shift; [ "$1" = '=https' ]; https_only=true ;;
  esac
  shift
done
case "$url" in
  https://example.invalid/release/v$release_version/*) [ "$https_only" = true ] ;;
  http://localhost:18080/release/v$release_version/*) ;;
  *) echo "Unexpected release URL: $url" >&2; exit 1 ;;
esac
case "${scenario:-}:$url" in
  download:*) exit 1 ;;
  checksum:*/SHA256SUMS) awk '{print "0000000000000000000000000000000000000000000000000000000000000000  " $2}' "$fixture_dir/SHA256SUMS" > "$output" ;;
  runtime:*/SHA256SUMS) shasum -a 256 "$runtime_binary" | awk -v name="$release_name" '{print $1 "  " name}' > "$output" ;;
  runtime:*) cp "$runtime_binary" "$output" ;;
  *) cp "$fixture_dir/${url##*/}" "$output" ;;
esac
MOCK
chmod 755 "$work/bin/curl"

sh release/install.sh
"$AISTD_INSTALL_DIR/aistd" version
cp "$AISTD_INSTALL_DIR/aistd" "$work/expected"
if printf 'n\n' | "$AISTD_INSTALL_DIR/aistd" uninstall >/dev/null 2>&1; then
  echo 'Uninstall did not honor cancellation' >&2; exit 1
fi
cmp "$work/expected" "$AISTD_INSTALL_DIR/aistd"
printf 'old CLI\n' > "$AISTD_INSTALL_DIR/aistd"
sh release/install.sh
cmp "$work/expected" "$AISTD_INSTALL_DIR/aistd"

case $(uname -s) in Darwin) os=darwin;; Linux) os=linux;; esac
case $(uname -m) in x86_64|amd64) arch=amd64;; arm64|aarch64) arch=arm64;; esac
export release_name=aistd_${release_version}_${os}_${arch}
export runtime_binary="$work/broken"
printf '#!/bin/sh\nexit 1\n' > "$runtime_binary"
for scenario in download checksum runtime; do
  export scenario
  if sh release/install.sh > "$work/error" 2>&1; then
    echo "Installer ignored $scenario failure" >&2; exit 1
  fi
  cmp "$work/expected" "$AISTD_INSTALL_DIR/aistd"
done
unset scenario

mv "$AISTD_INSTALL_DIR/aistd" "$work/original"
for target in symlink dangling directory; do
  case "$target" in
    symlink) ln -s "$work/original" "$AISTD_INSTALL_DIR/aistd" ;;
    dangling) ln -s "$work/missing" "$AISTD_INSTALL_DIR/aistd" ;;
    directory) mkdir "$AISTD_INSTALL_DIR/aistd" ;;
  esac
  if sh release/install.sh > "$work/error" 2>&1; then
    echo "Installer replaced $target target" >&2; exit 1
  fi
  case "$target" in
    symlink|dangling) [ -L "$AISTD_INSTALL_DIR/aistd" ]; rm "$AISTD_INSTALL_DIR/aistd" ;;
    directory) [ -d "$AISTD_INSTALL_DIR/aistd" ]; rmdir "$AISTD_INSTALL_DIR/aistd" ;;
  esac
  cmp "$work/expected" "$work/original"
done
mv "$work/original" "$AISTD_INSTALL_DIR/aistd"
for staged in "$AISTD_INSTALL_DIR"/.aistd-install.*; do
  [ ! -e "$staged" ] || { echo 'Installer left staging files' >&2; exit 1; }
done

export AISTD_RELEASE_BASE=http://localhost:18080/release
export AISTD_INSTALL_DIR="$work/local"
sh release/install.sh
sh release/install.sh
"$AISTD_INSTALL_DIR/aistd" version
"$AISTD_INSTALL_DIR/aistd" uninstall --yes
[ ! -e "$AISTD_INSTALL_DIR/aistd" ] || { echo 'Uninstall left CLI installed' >&2; exit 1; }
printf 'PASS: install, automatic replacement, failure preservation, unsafe targets, localhost and uninstall\n'
