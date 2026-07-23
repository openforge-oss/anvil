# Changelog

All notable changes are documented here, following
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial repository scaffold: Go module + CLI skeleton (`version`/`help`) and
  `internal/{detect,driver,pipeline}` package stubs.
- Working pipeline: `./check`/`./lint`/`./test`/`./build` wrapper scripts
  (Go-only), `Taskfile.yml`, and `CLAUDE.md` collaboration rules.
- Claude Code hooks: fingerprint-gated Stop-hook (`./check` on finish) and a
  dangerous-command guard, wired in `.claude/settings.json`.
- CI: cross-platform matrix (build/vet/test) + staticcheck + govulncheck, with
  concurrency cancellation.
- Docs: `README.md`, `docs/ROADMAP.md`, `docs/ARCHITECTURE.md`, and the
  `tasks/todo.md` / `tasks/lessons.md` progress + lessons files.

[Unreleased]: https://github.com/openforge-oss/anvil/commits/main
