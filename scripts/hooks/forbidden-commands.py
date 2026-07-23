#!/usr/bin/env python3
"""Claude Code PreToolUse hook (matcher: Bash).

Reads the tool-call JSON on stdin and blocks genuinely dangerous commands.
Exit code 2 blocks the call and feeds the message back to the agent.
"""
import json
import re
import sys


def block(msg: str) -> None:
    print(f"anvil guard: {msg}", file=sys.stderr)
    sys.exit(2)


def main() -> None:
    try:
        data = json.load(sys.stdin)
    except Exception:
        sys.exit(0)  # can't parse → don't get in the way

    cmd = (data.get("tool_input") or {}).get("command", "") or ""

    # Force-push (but allow the safer --force-with-lease).
    if re.search(r"git\s+push\b", cmd) and re.search(r"(--force(?!-with-lease)|\s-f(\s|$))", cmd):
        block("no force-push — open a PR instead (or use --force-with-lease if you must)")
    if re.search(r"\brm\s+-rf\s+(/|~|\$HOME|\*)", cmd):
        block("refusing a dangerous recursive delete")
    if "git reset --hard" in cmd:
        block("avoid 'git reset --hard' — stash or branch instead")
    if re.search(r"\bgit\s+clean\s+-\w*f\w*d", cmd):
        block("avoid 'git clean -fd…' — it deletes untracked files")

    sys.exit(0)


if __name__ == "__main__":
    main()
