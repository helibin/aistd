#!/bin/sh
# Run from any directory. Generated projects, build caches and logs stay in TMPDIR.
set -eu
mode=${1:-offline}
case "$mode" in
  offline) test "$#" -le 1 || { echo 'offline takes no platform arguments' >&2; exit 2; } ;;
  deep) ;;
  *) echo 'Usage: sh scripts/verify.sh offline | deep [api|nestjs|web|vue|expo|flutter-android|flutter-ios|taro|uniapp|vben|soybean ...]' >&2; exit 2 ;;
esac
if [ "$mode" = deep ]; then
  shift
  if [ "$#" -eq 0 ]; then set -- api nestjs web vue expo taro uniapp; fi
  for platform do
    case "$platform" in api|nestjs|web|vue|expo|flutter-android|flutter-ios|taro|uniapp|vben|soybean) ;; *) echo "Unknown platform: $platform" >&2; exit 2 ;; esac
  done
fi
: "${TMPDIR:?Set TMPDIR to a writable temporary directory}"
cli=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
work=$(mktemp -d "$TMPDIR/aistd-verify.XXXXXX")
cleanup() {
  status=$?
  if [ "$status" -eq 0 ]; then
    rm -rf "$work"
  else
    printf 'Verification FAILED (exit %s); evidence retained at %s\n' "$status" "$work" >&2
  fi
}
trap cleanup EXIT
cli_module_cache=$(go env GOMODCACHE)
export GOCACHE="$work/cache/go-build" GOMODCACHE="$work/cache/go-mod"
export npm_config_cache="$work/cache/npm" npm_config_update_notifier=false
export CI=1 EXPO_NO_TELEMETRY=1
# Preserve registry/auth configuration while isolating tool-written home files.
export npm_config_userconfig="${npm_config_userconfig:-${HOME:?}/.npmrc}"
export HOME="$work/home" XDG_CACHE_HOME="$work/cache/xdg"
if [ "$mode" = offline ]; then export GOPROXY=off GOSUMDB=off; fi
mkdir -p "$work/projects" "$HOME"
if [ "$mode" = offline ]; then
  # Read prepared CLI modules without downloads; generated projects keep their isolated empty cache.
  (cd "$cli" && GOMODCACHE="$cli_module_cache" go build -o "$work/aistd" ./cmd/aistd) || {
    echo 'Prepare CLI dependencies with make setup before offline verification.' >&2
    exit 1
  }
else
  (cd "$cli" && go build -o "$work/aistd" ./cmd/aistd)
fi
expect_failure() {
  if "$@" >"$work/expected-failure.log" 2>&1; then
    echo "Unexpected success: $*" >&2
    exit 1
  fi
  test -s "$work/expected-failure.log"
}
if [ "$mode" = offline ]; then
  "$work/aistd" create --name dry --targets api,web,ios,android,wechat,alipay --dir "$work/projects" --yes --dry-run >"$work/create-dry.log"
  test -s "$work/create-dry.log"
  test ! -e "$work/projects/dry"
  "$work/aistd" create --name alternatives --targets api,web,ios,android,wechat,alipay --api-stack nestjs --web-stack vue --mobile-stack flutter --miniapp-stack uniapp --dir "$work/projects" --yes --dry-run >"$work/alternatives-dry.log"
  test ! -e "$work/projects/alternatives"
  for admin_stack in vben soybean; do
    "$work/aistd" create --name "admin-$admin_stack" --targets admin --admin-stack "$admin_stack" --dir "$work/projects" --yes --dry-run >"$work/admin-$admin_stack-dry.log"
    test ! -e "$work/projects/admin-$admin_stack"
    "$work/aistd" create --name "admin-$admin_stack" --targets admin --admin-stack "$admin_stack" --dir "$work/projects" --yes >"$work/admin-$admin_stack-create.log"
    "$work/aistd" doctor --dir "$work/projects/admin-$admin_stack"
    test -s "$work/projects/admin-$admin_stack/apps/admin/package.json"
    test -s "$work/projects/admin-$admin_stack/.aistd/admin-template.json"
    cp "$work/projects/admin-$admin_stack/.aistd/project.json" "$work/admin-$admin_stack.before"
    expect_failure "$work/aistd" create --name "admin-$admin_stack" --targets admin --admin-stack "$admin_stack" --dir "$work/projects" --yes
    cmp "$work/admin-$admin_stack.before" "$work/projects/admin-$admin_stack/.aistd/project.json"
  done
  "$work/aistd" create --name demo --targets api --dir "$work/projects" --yes
  "$work/aistd" doctor --dir "$work/projects/demo"
  cp "$work/projects/demo/README.md" "$work/readme.before"
  expect_failure "$work/aistd" create --name demo --targets api --dir "$work/projects" --yes
  cmp "$work/readme.before" "$work/projects/demo/README.md"
  mkdir "$work/projects/existing"
  printf 'existing instructions\n' >"$work/projects/existing/AGENTS.md"
  cp "$work/projects/existing/AGENTS.md" "$work/agents.before"
  "$work/aistd" init --dir "$work/projects/existing" --dry-run >"$work/init-dry.log"
  test -s "$work/init-dry.log"
  test "$(find "$work/projects/existing" -type f | wc -l | tr -d ' ')" = 1
  test ! -e "$work/projects/existing/.aistd"
  "$work/aistd" init --dir "$work/projects/existing"
  "$work/aistd" doctor --dir "$work/projects/existing"
  expect_failure "$work/aistd" init --dir "$work/projects/existing"
  cmp "$work/agents.before" "$work/projects/existing/AGENTS.md"
  "$work/aistd" doctor --dir "$work/projects/existing"
  printf 'tampered\n' >>"$work/projects/demo/.aistd/standards/visual/v1.0/docs/VISUAL-STANDARD.md"
  expect_failure "$work/aistd" doctor --dir "$work/projects/demo"
  # The API has no external modules. Disable downloads to enforce offline validation.
  (cd "$work/projects/demo" && GOPROXY=off GOSUMDB=off make setup check build)
  test -x "$work/projects/demo/bin/api"
  echo 'PASS: offline CLI lifecycle, Admin templates, overwrite/dry-run/tamper rejection, API check/build (no npm installation).'
else
  for platform do
    case "$platform" in
      api) set -- --targets api ;;
      nestjs) set -- --targets api --api-stack nestjs ;;
      web) set -- --targets web ;;
      vue) set -- --targets web --web-stack vue ;;
      expo) set -- --targets ios,android ;;
      flutter-android) set -- --targets android --mobile-stack flutter ;;
      flutter-ios) set -- --targets ios --mobile-stack flutter ;;
      taro) set -- --targets wechat,alipay ;;
      uniapp) set -- --targets wechat,alipay --miniapp-stack uniapp ;;
      vben|soybean) set -- --targets admin --admin-stack "$platform" ;;
    esac
    "$work/aistd" create --name "$platform" "$@" --dir "$work/projects" --yes
    "$work/aistd" doctor --dir "$work/projects/$platform"
    (cd "$work/projects/$platform" && make setup check build) >"$work/$platform.log" 2>&1 || {
      echo "FAIL: $platform install/check/build; see $work/$platform.log" >&2
      exit 1
    }
    case "$platform" in
      api) test -x "$work/projects/api/bin/api" ;;
      nestjs) test -s "$work/projects/nestjs/apps/api/dist/main.js" ;;
      web|vue) test -s "$work/projects/$platform/apps/web/dist/index.html" ;;
      vben) test -s "$work/projects/vben/apps/admin/apps/web-naive/dist/index.html" ;;
      soybean) test -s "$work/projects/soybean/apps/admin/dist/index.html" ;;
      flutter-android) test -s "$work/projects/$platform/apps/mobile/build/app/outputs/flutter-apk/app-debug.apk" ;;
      flutter-ios) test -d "$work/projects/$platform/apps/mobile/build/ios/iphoneos/Runner.app" ;;
      expo)
        test -s "$work/projects/expo/apps/mobile/dist/ios/metadata.json"
        test -s "$work/projects/expo/apps/mobile/dist/android/metadata.json" ;;
      taro)
        test -s "$work/projects/taro/apps/miniapp/dist/weapp/app.json"
        test -s "$work/projects/taro/apps/miniapp/dist/alipay/app.json" ;;
      uniapp)
        test -s "$work/projects/uniapp/apps/miniapp/dist/build/mp-weixin/app.json"
        test -s "$work/projects/uniapp/apps/miniapp/dist/build/mp-alipay/app.json" ;;
    esac
    echo "PASS: $platform install/check/build and output artifacts"
  done
fi
