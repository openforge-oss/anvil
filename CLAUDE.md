# CLAUDE.md — how to work in `anvil`

`anvil` is a guided, zero-config **build & release pipeline CLI** for mobile and
app projects (Go, single binary). This file is the contract for any AI agent
working here. Read it, and `tasks/lessons.md`, at the start of every session.

## Golden rules (non-negotiable)

1. **Plan first.** For anything non-trivial (3+ steps or a design choice), write
   the plan to `tasks/todo.md` and get confirmation before executing. If it goes
   sideways, stop and re-plan.
2. **Confirm, don't assume.** Never make a consequential decision (naming, scope,
   architecture, dependencies, anything outward-facing) on your own — ask first.
3. **Branch → PR, never push `main`.** Cut `feat/…` / `fix/…` from `develop`;
   open a PR into `develop`. `main` is protected and released.
4. **Verify before "done".** Run `./check` and prove it passes. "Would a staff
   engineer approve this?" No feature is complete without evidence.
5. **Simplicity first, no laziness, minimal impact.** Find root causes; touch
   only what's necessary; no temporary hacks.
6. **Commits are the user's.** Author commits as the user's git identity. Never
   add a `Co-Authored-By: Claude` trailer or otherwise credit the agent.

## The loop

1. Session start: skim `tasks/todo.md` (where we are) and `tasks/lessons.md`
   (mistakes not to repeat).
2. Plan → confirm → implement in small steps → **`./check`** → open a PR.
3. After **any** correction from the user, append a `mistake → rule` line to
   `tasks/lessons.md`. Iterate until the mistake rate drops.

## Commands (use the wrappers, not raw tools)

- `./check` — lint + test. **Run before every push.**
- `./lint` — gofmt check, `go vet`, staticcheck/govulncheck (if installed).
- `./test` — `go test ./... -race -cover` (args pass through).
- `./build` — build `bin/anvil` (`GORELEASE=vX.Y.Z ./build` stamps the version).

(`task <name>` also works if [go-task](https://taskfile.dev) is installed.)

## Token & context economy (treat tokens as money)

- **Locate, then read narrowly** (grep/glob → read the needed range). Never read a
  whole large file to find one thing. Never re-read a file you just wrote.
- **Filter output at the source** — pipe through `grep`/`head`/`tail`; never dump
  whole logs or JSON.
- **Batch & parallelize** independent reads/searches; don't re-run expensive commands.
- **Offload heavy research to subagents** and take back only the conclusion — keep
  the main context clean.
- **Subagent test discipline:** subagents lint-only; the orchestrator runs the full
  `./test` **once** at the end — never N× per subagent.
- **Be concise & surgical.** Reference `path:line`; change the few lines that matter;
  no rewrites, no filler. No rabbit holes — if it isn't converging, stop and re-plan.

## Review loop

- CodeRabbit / `@claude` are **opt-in** (comment-triggered), not on every commit.
- Cap automated fix iterations at **3**; escalate security or ambiguous items to the
  user rather than looping.

## Automation in this repo

- **Stop hook** (`scripts/hooks/check-on-stop.sh`): auto-runs `./check` when you
  finish — but only if the working tree changed since the last check (fingerprint
  cached in `.cache/`). If it fails, fix before finishing.
- **Command guard** (`scripts/hooks/forbidden-commands.py`): blocks force-push,
  `rm -rf` on dangerous paths, `git reset --hard`, `git clean -fd…`.
- Both are wired in `.claude/settings.json`.

## Where things are

- `cmd/` — CLI entry (stdlib skeleton today; Cobra/Charm arrive with the tool build).
- `internal/detect`, `internal/driver`, `internal/pipeline` — the tool's core
  (detector + per-ecosystem driver + orchestrator). See `docs/ARCHITECTURE.md`.
- `docs/ROADMAP.md` — phases and status. `tasks/todo.md` — live progress.
