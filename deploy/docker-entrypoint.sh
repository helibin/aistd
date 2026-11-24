#!/bin/sh
set -eu

site=/usr/share/nginx/html
install_url=${AISTD_INSTALL_URL:-}

if [ -n "$install_url" ]; then
  case "$install_url" in
    https://*/install.sh|http://localhost/install.sh|http://localhost:[0-9]*/install.sh|http://127.0.0.1/install.sh|http://127.0.0.1:[0-9]*/install.sh) ;;
    *) echo 'AISTD_INSTALL_URL must be an HTTPS URL, or a localhost HTTP URL, ending in /install.sh' >&2; exit 1 ;;
  esac
  case "$install_url" in
    *[!A-Za-z0-9._~:/%-]*) echo 'AISTD_INSTALL_URL contains unsupported characters' >&2; exit 1 ;;
  esac
  site_host=${install_url#https://}
  site_host=${site_host#http://}
  site_host=${site_host%/install.sh}
  case "$site_host" in
    ''|*/*) echo 'AISTD_INSTALL_URL must point to /install.sh at the site root' >&2; exit 1 ;;
  esac
  release_base=${AISTD_RELEASE_BASE:-${install_url%/install.sh}/release}
  case "$release_base" in
    https://*|http://localhost|http://localhost:[0-9]*|http://127.0.0.1|http://127.0.0.1:[0-9]*) ;;
    *) echo 'AISTD_RELEASE_BASE must use HTTPS, or localhost HTTP for local deployment' >&2; exit 1 ;;
  esac
  case "$release_base" in
    *[!A-Za-z0-9._~:/%-]*) echo 'AISTD_RELEASE_BASE contains unsupported characters' >&2; exit 1 ;;
  esac
  printf "window.AISTD_INSTALL_URL = '%s';\n" "$install_url" > "$site/site-config.js"
  sed "s|__AISTD_RELEASE_BASE__|$release_base|g" /etc/aistd/install.sh > "$site/install.sh"
else
  printf "window.AISTD_INSTALL_URL = '';\n" > "$site/site-config.js"
  rm -f "$site/install.sh"
fi

exec nginx -g 'daemon off;'
