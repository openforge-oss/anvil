# Roadmap

High-level phases and status, for monitoring. Granular tasks live in
[`../tasks/todo.md`](../tasks/todo.md).

| Phase | What | Status |
|-------|------|--------|
| 0. Repo and pipeline | Go skeleton, wrapper scripts, Claude hooks, CI, docs, working rules | done |
| 1. Detection | Marker-file detectors for Flutter, React Native, Android, iOS, Swift, Kotlin; `anvil detect` | done |
| 2. Guided build | Driver lifecycle (deps, analyze, test, build) for all six stacks; `--flavor`; interactive TUI plus plain fallback; `anvil build` | in progress |
| 3. Signing | Guided Android keystore and iOS provisioning/signing | in progress |
| 4. Upload | TestFlight, Play, npm upload; GoReleaser distribution (Homebrew, Scoop, curl) | planned |
| 5. Breadth | More ecosystems (web, Go) via new drivers; flavor auto-detection; `--explain` educational mode | future |

Swift (SPM) and Kotlin/JVM drivers and `--flavor` were pulled forward into
Milestone 2.

## Guiding decisions (locked)

- Language: Go. Single static binary, neutral (not tied to any framework), strong
  guided-CLI toolkit, trivial cross-platform distribution.
- Shape: local-first, zero-config, auto-detecting, interactive. Not a cloud CI,
  not a config-heavy monorepo tool.
- Scope: mobile-first in v1. The driver architecture keeps new stacks cheap.

## Non-goals (for now)

- Replacing CI platforms (GitHub Actions, Codemagic). anvil is the local shortcut;
  it can emit CI config later.
- Being a monorepo task graph (that is Nx, Bazel, moon).
