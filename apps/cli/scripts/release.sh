#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
make check
version=$(cat VERSION)
mkdir -p "release/v$version"
sed "s/__AISTD_VERSION__/$version/g" scripts/install.sh > release/install.sh
for os in darwin linux; do
 for arch in amd64 arm64; do
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags='-s -w' -o "release/v$version/aistd_${version}_${os}_${arch}" ./cmd/aistd
 done
done
(cd "release/v$version"; if command -v sha256sum >/dev/null 2>&1; then sha256sum aistd_* > SHA256SUMS; else shasum -a 256 aistd_* > SHA256SUMS; fi)
sh scripts/install_test.sh
echo "Release artifacts created in apps/cli/release/v$version/. Smoke-test target binaries before publication."
