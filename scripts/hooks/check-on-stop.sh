#!/usr/bin/env bash
# Claude Code Stop hook. When the agent finishes, run ./check — but ONLY if the
# working tree changed since the last check (fingerprint-gated), so we never
# waste a run. On failure, surface it once (exit 2) so the agent fixes it; the
# fingerprint is recorded first so the same failing state won't loop.
set -uo pipefail
repo_root="${CLAUDE_PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
cd "$repo_root" || exit 0

git rev-parse --is-inside-work-tree >/dev/null 2>&1 || exit 0

fp_file=".cache/last_check_fingerprint"
mkdir -p .cache

status="$(git status --porcelain 2>/dev/null)"
[ -z "$status" ] && exit 0  # clean tree, nothing to verify

if command -v md5sum >/dev/null 2>&1; then
  fingerprint="$(printf '%s' "$status" | md5sum | awk '{print $1}')"
else
  fingerprint="$(printf '%s' "$status" | md5)"
fi

[ "$fingerprint" = "$(cat "$fp_file" 2>/dev/null || true)" ] && exit 0
printf '%s' "$fingerprint" > "$fp_file"  # record before running to avoid loops

if ./check >/tmp/anvil_check.log 2>&1; then
  exit 0
fi

echo "anvil: ./check failed after your changes — fix before finishing:" >&2
tail -n 30 /tmp/anvil_check.log >&2
exit 2
