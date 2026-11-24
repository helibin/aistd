#!/bin/sh
# Build a complete static release for Workers, Pages, Vercel or EdgeOne without changing pinned standards.
set -eu
cd "$(dirname "$0")/.."

install_url=${AISTD_INSTALL_URL:?Set AISTD_INSTALL_URL to the production HTTPS /install.sh URL}
case "$install_url" in
  https://*/install.sh) ;;
  *) echo 'AISTD_INSTALL_URL must be an HTTPS URL ending in /install.sh' >&2; exit 1 ;;
esac
case "$install_url" in
  *[!A-Za-z0-9._~:/%-]*) echo 'AISTD_INSTALL_URL contains unsupported characters' >&2; exit 1 ;;
esac
site_host=${install_url#https://}
site_host=${site_host%/install.sh}
case "$site_host" in
  ''|*/*) echo 'AISTD_INSTALL_URL must point to /install.sh at the site root' >&2; exit 1 ;;
esac
release_base=${install_url%/install.sh}/release

make release
version=$(cat apps/cli/VERSION)
mkdir -p "dist/release/v$version"
cp apps/cli/release/v"$version"/aistd_* apps/cli/release/v"$version"/SHA256SUMS "dist/release/v$version/"
printf "window.AISTD_INSTALL_URL = '%s';\n" "$install_url" > dist/site-config.js
sed "s|__AISTD_RELEASE_BASE__|$release_base|g" apps/cli/release/install.sh > dist/install.sh
(cd "dist/release/v$version" && shasum -a 256 -c SHA256SUMS)
echo "Static release ready in dist/ for $install_url"
