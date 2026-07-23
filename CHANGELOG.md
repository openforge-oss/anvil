# Changelog

All notable changes are documented here, following
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Upload: `anvil upload` pushes a signed artifact to its store. iOS via
  `xcrun altool` to App Store Connect/TestFlight, Android via the Google Play
  Publisher API (insert edit, upload bundle, assign track, commit) using a
  service account, and npm via `npm publish`. Credentials come from flags, env,
  or a base64 env decoded to a temp file; a credential inside the repo is
  refused; uploads are a dry run unless `--yes`.
- Self-distribution: a GoReleaser config and a tag-triggered release workflow
  that build cross-platform binaries and publish a GitHub release plus Homebrew
  (cask) and Scoop manifests.
- Guided release signing: `anvil sign` and `anvil build --sign`, a Sign phase
  that runs after Build. Android setup generates a PKCS12 keystore with keytool,
  writes key.properties, wires Gradle signingConfigs, and gitignores the secrets,
  so a release build comes out signed. iOS writes an ExportOptions.plist and then
  archives and exports a development or ad-hoc signed ipa (Flutter via
  `flutter build ipa`, React Native and native iOS via xcodebuild). Passwords
  come from prompts or environment, never the repo; `--dry-run` changes nothing.
- Build lifecycle and `anvil build`: runs deps, analyze, test, and an
  unsigned/debug/simulator build for a detected project, with a live Bubble Tea
  view and a plain non-TTY renderer. Flags `--path`, `--target`, `--flavor`,
  `--release`, `--dry-run`, `--plain`.
- Driver contract (`internal/driver`) and drivers for Flutter, React Native,
  native Android, native iOS, Swift (SPM), and Kotlin/JVM, with `--flavor`
  threaded into build and test steps.
- Runner (`internal/pipeline`): executes steps, streams combined output, applies
  the driver's classification, collects artifacts, and stops on first failure.
- Detection refinement: Gradle is not always Android and `Package.swift` is
  Swift, adding Swift and Kotlin stacks.
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
