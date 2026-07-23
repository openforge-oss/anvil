# anvil — todo & progress

Live progress board. Check items off as they land; add a **Review** note per
milestone. This is the file to read first to see where we are.

## Milestone 0 — repo + working pipeline (in progress)

- [x] Go skeleton: module, `cmd` (version/help), `internal/` placeholders, builds & tests clean
- [x] Wrapper scripts: `./check`, `./lint`, `./test`, `./build` (+ `Taskfile.yml`)
- [x] Claude hooks: smart Stop-hook (fingerprint-gated `./check`) + command guard, wired in `.claude/settings.json`
- [x] `CLAUDE.md` (how we work), `tasks/todo.md`, `tasks/lessons.md`
- [x] CI: matrix build/vet/test + staticcheck + govulncheck; concurrency-cancel
- [x] Docs: `README.md`, `docs/ROADMAP.md`, `docs/ARCHITECTURE.md`, `CHANGELOG.md`
- [ ] Repo created on GitHub, pushed, branch protection on `main`+`develop`
- [ ] CI green on first push

### Review — Milestone 0
_(fill in once pushed: what shipped, what to watch, any follow-ups)_

## Milestone 1 — detection engine (next phase, separate plan)

- [ ] `detect` interface + marker-file detectors: Flutter, React Native, Android, iOS
- [ ] `anvil detect` command prints the resolved stack(s) + confidence

## Milestone 2 — guided build lifecycle

- [ ] `driver` lifecycle contract; Flutter driver first (deps → analyze → test → build)
- [ ] Interactive TUI (Charm/Bubble Tea); unified error surfacing
- [ ] Android + iOS + React Native drivers

## Milestone 3 — signing & upload

- [ ] Guided Android keystore + iOS provisioning/signing
- [ ] Store/registry upload (TestFlight / Play / npm)
- [ ] GoReleaser → Homebrew/Scoop/curl distribution
