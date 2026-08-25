# anvil todo and progress

Live progress board. Check items off as they land, add a Review note per
milestone. Read this first to see where we are.

## Milestone 0: repo and working pipeline (done)

- [x] Go skeleton: module, cmd (version/help), internal placeholders, builds and tests clean
- [x] Wrapper scripts: `./check`, `./lint`, `./test`, `./build` (plus `Taskfile.yml`)
- [x] Claude hooks: fingerprint-gated Stop hook plus command guard, wired in `.claude/settings.json`
- [x] `CLAUDE.md`, `tasks/todo.md`, `tasks/lessons.md`
- [x] CI: matrix build/vet/test plus staticcheck and govulncheck, concurrency-cancel
- [x] Docs: `README.md`, `docs/ROADMAP.md`, `docs/ARCHITECTURE.md`, `CHANGELOG.md`
- [x] Repo created on GitHub, pushed, branch protection on main and develop
- [x] CI green on first push

### Review, Milestone 0
Shipped the repo and the full working pipeline. CI is green on all jobs. Next:
build the detection engine.

## Milestone 1: detection engine (in progress)

- [x] Style cleanup (no emojis, no dash connectors, minimal comments); Style section in CLAUDE.md
- [x] Adopt Cobra; cmd restructured (root, version, detect)
- [x] `internal/detect`: types, scanner (prune-on-detect, skip lists, containment sweep)
- [x] Detectors: Flutter, React Native, Android, iOS
- [x] `anvil detect` command (table plus `--json`, `--path`, `--depth`)
- [x] Tests over fixture trees; `node_modules` exclusion and android/ios absorption covered
- [ ] PR into develop, CI green

### Review, Milestone 1
Detection works end to end. `anvil detect` returns one row per real project and
attributes android/ios folders to their Flutter/RN parent (verified on ca-mobile
and motobites). 16 fixture-tree tests pass; detect package coverage 89.6%.
Known limitation: a project nested inside a detected root is not separately
surfaced (add-to-app, plugin example), a consequence of prune-on-detect; revisit
if needed. Next: Milestone 2 (guided build lifecycle).

## Milestone 2: guided build lifecycle (in progress)

- [x] `driver` lifecycle contract (Phase, Status, Step, BuildOptions, Driver) + registry
- [x] Drivers: Flutter, React Native, Android, iOS, plus Swift and Kotlin/JVM (pulled forward)
- [x] `--flavor` threaded through build and test steps
- [x] Detection refinement: Gradle is not always Android, `Package.swift` is Swift, plus Swift and Kotlin stacks
- [x] `internal/pipeline` runner: exec streaming, exit codes, classify, artifacts, fail-fast
- [x] `internal/tui`: Bubble Tea view + plain non-TTY renderer
- [x] `anvil build` command (`--path`, `--target`, `--flavor`, `--release`, `--dry-run`, `--plain`)
- [x] Tests: driver Steps, runner via subprocess helper, plain renderer; `./check` green
- [ ] PR into develop, CI green

### Review, Milestone 2
Build lifecycle works end to end. `anvil build` detects the project, picks the
driver, and runs deps, analyze, test, build with a live TUI or a plain CI
renderer. Verified: dry-run and flavor wiring on ca-mobile, the runner via a real
subprocess (streaming, exit codes, fail-fast), and the no-project error path.
Signing and store upload remain for Milestones 3 and 4. iOS scheme is derived by
convention (or `--flavor`); auto-detecting flavors and schemes is a later step.

## Milestone 3: signing (in progress)

- [x] Sign phase in the driver contract; iOS sign steps on Flutter, RN, and native iOS
- [x] `internal/sign`: keytool keystore generation, key.properties, Gradle wiring, ExportOptions.plist, gitignore
- [x] `anvil sign` and `anvil build --sign`; `--dry-run` makes no changes
- [x] Secrets via prompt or env (huh), never committed
- [x] Tests: live keystore gen (keytool), gradle wiring idempotence, gitignore, ExportOptions, Step argv
- [ ] PR into develop, CI green

### Review, Milestone 3
Android signs at build time via wired Gradle signingConfigs; iOS archives and
exports a signed ipa. Guided setup is side-effect-free under --dry-run (a bug
caught in review after it briefly wrote into a real project, now fixed and the
project restored). Deferred: App Store export, App Store Connect API and
fastlane match, OS keychain, Play enrollment, and Android apksigner for a loose
prebuilt APK.

## Milestone 5: breadth, Go and web/Node (in progress)

- [x] detect: Go (go.mod, app/library) and Web (framework/app/library; workspace roots descend)
- [x] drivers: Go (deps, vet + gofmt, test, build; Analyze classify for gofmt) and Web (install, lint, test, build via package-manager scripts)
- [x] jsInstall Yarn Berry (`--immutable`); `pmRun` helper; framework output dirs added to skip list
- [x] Tests (Go app/library/cmd/go.work; web frameworks; workspace-root and bare not claimed); ./check + staticcheck green
- [x] Verified: anvil self-detects as go/app; build --dry-run shows the Go pipeline
- [ ] PR into develop, CI green

### Review, Milestone 5
anvil now spans eight stacks. Go and web reuse the detector+driver contract with
no new dependencies. Web detection descends into monorepo roots so members still
surface, and treats a single-package Turbo repo as a leaf. Remaining polish
issues: #4, #5, #6, #9, #10.

## Milestone 4: upload

- [x] `internal/upload`: Uploader interface + iOS (altool), Android (Play API), npm
- [x] Credential resolution (flag/env/base64-to-temp) that refuses in-repo secrets
- [x] `anvil upload` with dry-run default (`--yes` to perform)
- [x] GoReleaser self-distribution: `.goreleaser.yaml` + tag-triggered `release.yml`
- [x] Tests (cred resolution, in-repo refusal, Validate, Describe); ./check + staticcheck green
- [ ] PR into develop, CI green

### Review, Milestone 4
Upload works end to end in dry-run for all three targets (verified). Live pushes
need real store accounts, so they are deferred; unit tests cover credential
resolution, the in-repo refusal, and each uploader's validation and plan.
GoReleaser was pulled into this milestone; it needs `openforge-oss/homebrew-tap`
and `scoop-bucket` repos plus a `HOMEBREW_TAP_TOKEN` secret before the first
release tag. Deferred: App Store submission metadata, Play staged rollout,
fastlane back-ends, `anvil build --upload`, npm OIDC. This completes the core
detect -> build -> sign -> upload pipeline.

## Issue 9: lifecycle phase selection

- [x] Add mutually exclusive `anvil build --until <phase>` and `--only <phase>` flags.
- [x] Validate phase names and filter the selected driver's plan before dry-run or execution.
- [x] Cover inclusive prefix selection, single-phase selection, invalid values, and conflicting flags.
- [ ] Run `./check`, review the focused diff, and open a pull request into `develop`.
