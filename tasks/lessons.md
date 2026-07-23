# Lessons — mistake → prevention rule

Append a line after **every** correction from the user. Format:
`YYYY-MM-DD: <mistake> → <rule to prevent it>`. Review at session start.

Seeded from the OpenForge collaboration so far:

- 2026-07-23: Credited Claude as commit co-author → **Never** add `Co-Authored-By: Claude` or any agent attribution; commits are authored solely by the user's git identity.
- 2026-07-23: Made naming/scope decisions unilaterally → **Confirm every consequential decision** (names, scope, architecture, dependencies, outward-facing actions) before acting.
- 2026-07-23: Assumed a Flutter-only framing → OpenForge is **framework-agnostic**; design tools to span many stacks unless told otherwise.
- 2026-07-23: Confused an org's display **Name** with the **Rename organization** (login/URL) action → they are different settings; verify the actual `login` via API, not the display name.
- 2026-07-23: Reached for a tool version that lags a brand-new toolchain (prebuilt binary vs bleeding-edge Go) → prefer `go run …@latest` / source-built tools that track the installed toolchain to avoid red CI.
