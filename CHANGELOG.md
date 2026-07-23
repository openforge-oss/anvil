# Changelog

All notable changes are documented here, following
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Stack detection engine (`internal/detect`): marker-file detectors for Flutter
  (vs pure Dart; app/module/plugin subtypes), React Native (bare, Expo managed,
  Expo prebuild), native Android (app vs library, KMP flag), and native iOS
  (Xcode project/workspace, SPM, Podfile). Depth-limited prune-on-detect scanner
  with skip lists and a containment sweep, so `android/`/`ios/` folders are
  attributed to their Flutter or React Native parent and monorepos surface each
  project once.
- `anvil detect` command: table and `--json` output, `--path` and `--depth` flags.
- Adopted Cobra for the CLI.
- Style rules in `CLAUDE.md` (no emojis, no dash connectors, minimal comments).
- Initial repository scaffold: Go module plus CLI skeleton (`version`, `help`)
  and `internal/{detect,driver,pipeline}` package stubs.
- Working pipeline: `./check`, `./lint`, `./test`, `./build` wrapper scripts
  (Go only), `Taskfile.yml`, and `CLAUDE.md` collaboration rules.
- Claude Code hooks: fingerprint-gated Stop hook (`./check` on finish) and a
  dangerous-command guard, wired in `.claude/settings.json`.
- CI: cross-platform matrix (build, vet, test) plus staticcheck and govulncheck,
  with concurrency cancellation.
- Docs: `README.md`, `docs/ROADMAP.md`, `docs/ARCHITECTURE.md`, and the
  `tasks/todo.md` and `tasks/lessons.md` progress and lessons files.

[Unreleased]: https://github.com/openforge-oss/anvil/commits/main
