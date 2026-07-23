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

## Milestone 2: guided build lifecycle

- [ ] `driver` lifecycle contract; Flutter driver first (deps, analyze, test, build)
- [ ] Interactive TUI (Charm/Bubble Tea); unified error surfacing
- [ ] Android, iOS, React Native drivers

## Milestone 3: signing and upload

- [ ] Guided Android keystore and iOS provisioning/signing
- [ ] Store/registry upload (TestFlight, Play, npm)
- [ ] GoReleaser to Homebrew/Scoop/curl distribution
