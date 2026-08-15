#!/usr/bin/env bash
# .github/workflows/backend-ci.yml / frontend-ci.yml と同じ静的解析を、
# 変更されたファイルのスコープだけに絞ってローカルで先取り実行する（PostToolUse: Edit|Write|MultiEdit）。
# Go の変更でTS一式、TSの変更でGo一式を毎回走らせないことで高速さを保つ。
set -uo pipefail

input=$(cat)
file=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')
[ -z "$file" ] && exit 0
[ -f "$file" ] || exit 0

repo_root=$(git rev-parse --show-toplevel 2>/dev/null) || exit 0
rel=${file#"$repo_root"/}

status=0
out=""

run() {
  local label="$1"
  shift
  local result
  if ! result=$("$@" 2>&1); then
    status=1
    out="${out}
--- ${label} failed ---
${result}"
  fi
}

case "$rel" in
  backend/*.go)
    pkg_dir=$(dirname "$rel")
    pkg_rel=${pkg_dir#backend/}
    pushd "$repo_root/backend" >/dev/null || exit 0
    run "golangci-lint" golangci-lint run "./${pkg_rel}/..."
    run "go build" go build "./${pkg_rel}/..."
    run "go test" go test "./${pkg_rel}/..."
    popd >/dev/null
    ;;

  *.ts | *.tsx)
    pushd "$repo_root" >/dev/null || exit 0
    run "oxlint" pnpm exec oxlint "$rel"
    run "oxfmt --check" pnpm exec oxfmt --check "$rel"
    run "typecheck" pnpm typecheck

    case "$rel" in
      *.test.ts | *.spec.ts)
        run "vitest" pnpm exec vitest run "$rel"
        ;;
      *.test.tsx)
        run "jest VRT" pnpm exec jest "$rel"
        ;;
      *.tsx)
        run "jest VRT (related)" pnpm exec jest --findRelatedTests "$rel" --passWithNoTests
        ;;
      *.ts)
        run "vitest (related)" pnpm exec vitest related "$rel" --run
        ;;
    esac
    popd >/dev/null
    ;;

  *)
    exit 0
    ;;
esac

if [ "$status" -ne 0 ]; then
  printf 'CIチェックのローカル先取り実行で問題が見つかりました（%s）:\n%s\n' "$rel" "$out" >&2
  exit 2
fi

exit 0
