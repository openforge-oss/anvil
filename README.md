<div align="center">

# anvil

**A guided, zero-config build and release pipeline for mobile and app projects.**

One command detects the stack, fetches dependencies, analyzes, surfaces errors,
tests, and builds (then signs and uploads), without memorizing each framework's CLI.

[![CI](https://github.com/openforge-oss/anvil/actions/workflows/ci.yml/badge.svg)](https://github.com/openforge-oss/anvil/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
![status](https://img.shields.io/badge/status-early%20development-orange)

Part of [OpenForge](https://github.com/openforge-oss).

</div>

## Status

Early development, Milestone 1 (detection engine). The CLI can detect a project's
stack; the build lifecycle comes next. Track progress in
[`tasks/todo.md`](tasks/todo.md) and [`docs/ROADMAP.md`](docs/ROADMAP.md).

## The problem

Shipping a mobile or app build is a fiddly, error-prone grind: Gradle and AGP
version matrices, the CocoaPods to Swift Package Manager migration, iOS
provisioning and code signing, and per-stack build commands nobody remembers.
Existing tools either need config and a cloud account (fastlane, Codemagic), only
output server containers (Nixpacks, buildpacks), or are heavy monorepo build
systems (Nx, Bazel). None is a local, zero-config, auto-detecting, guided CLI
that produces mobile artifacts.

## The idea

```console
$ anvil ship          # (planned)
Detected: Flutter app (android, ios)
flutter pub get       ok
flutter analyze       0 issues
flutter test          42 passed
Build target? Android App Bundle (.aab) / iOS Archive (.ipa) / Both
...
```

Auto-detect the stack, run the right lifecycle, explain failures in plain
language, and (later) walk you through signing and store upload. It graduates
beginners by printing the exact commands it runs.

- v1 stacks: Flutter, React Native, native Android, native iOS.
- Built in Go (single static binary; nothing to install but the binary).
- Pluggable: adding a framework means adding one driver. See
  [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Contributing

This repo runs a disciplined, low-token workflow. Read [`CLAUDE.md`](CLAUDE.md)
first. In short: branch then PR (never push `main`), and run `./check` (lint plus
test) before every push.

```bash
./check    # gofmt, go vet, staticcheck/govulncheck, go test -race
./build    # produces bin/anvil
```

## License

[MIT](LICENSE), OpenForge, 2026.
