<div align="center">

# anvil

**A guided, zero-config build and release pipeline for mobile and app projects.**

One command detects the stack, fetches dependencies, analyzes, surfaces errors,
tests, builds, signs, and uploads, without memorizing each framework's CLI.

[![CI](https://github.com/openforge-oss/anvil/actions/workflows/ci.yml/badge.svg)](https://github.com/openforge-oss/anvil/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/openforge-oss/anvil?sort=semver)](https://github.com/openforge-oss/anvil/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Part of [OpenForge](https://github.com/openforge-oss).

</div>

## What it does

Point anvil at a project and it works out the stack, then runs the right
lifecycle: dependencies, static analysis, tests, build, and (where set up)
signing and store upload. It prints every command it runs, so it is also a way to
learn the underlying tools instead of hiding them.

```console
$ anvil detect
PATH  STACK    SUBTYPE  CONFIDENCE
.     flutter  app      0.98

$ anvil build
flutter pub get       ok
flutter analyze       0 issues
flutter test          42 passed
flutter build appbundle
```

### Supported stacks

Flutter, React Native, native Android, native iOS, Swift, Kotlin/JVM, Go, and
web/Node (Next, Nuxt, SvelteKit, Angular, Vite, CRA, Vue, Svelte, Astro, Remix,
Gatsby). Adding another is one driver. See
[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Install

anvil is a single static binary.

### Homebrew (macOS and Linux)

```bash
brew install --cask openforge-oss/tap/anvil
```

### Scoop (Windows)

```powershell
scoop bucket add openforge-oss https://github.com/openforge-oss/scoop-bucket
scoop install anvil
```

### Prebuilt binary

Download the archive for your OS and architecture from the
[latest release](https://github.com/openforge-oss/anvil/releases/latest), extract
it, and put `anvil` on your `PATH`.

### Go

```bash
go install github.com/openforge-oss/anvil@latest
```

Installs to `$(go env GOPATH)/bin`, which must be on your `PATH`.

## Usage

```bash
anvil detect                       # identify the project and its stack
anvil build                        # deps, analyze, test, build (guided)
anvil build --release --flavor prod
anvil build --until analyze        # deps, then analyze
anvil build --only test            # test phase only
anvil sign                         # set up Android or iOS signing
anvil build --sign                 # build and sign
anvil upload                       # dry-run by default, pass --yes to perform
```

Useful flags: `--path` (project directory), `--target` (android or ios),
`--flavor`, `--release`, `--until` (inclusive lifecycle prefix), `--only` (one
lifecycle phase), `--dry-run` (print the plan without running it), and `--plain`
(no TUI, for CI). `--until` and `--only` are mutually exclusive. Every command
has `--help`.

anvil never puts secrets on the command line or in the repo. Keystore and store
credentials come from a prompt or an environment variable, and a credential
located inside the working tree is refused.

## The problem

Shipping a build is a fiddly, error-prone grind: Gradle and AGP version matrices,
the CocoaPods to Swift Package Manager migration, iOS provisioning and code
signing, and per-stack commands nobody remembers. Existing tools either need
config and a cloud account (fastlane, Codemagic), only output server containers
(Nixpacks, buildpacks), or are heavy monorepo build systems (Nx, Bazel). anvil is
a local, zero-config, auto-detecting, guided CLI that produces real artifacts.

## Contributing

This repo runs a disciplined, low-token workflow. Read [`CLAUDE.md`](CLAUDE.md)
first. In short: branch then PR (never push `main`), and run `./check` (lint plus
test) before every push.

```bash
./check    # gofmt, go vet, staticcheck/govulncheck, go test -race
./build    # produces bin/anvil
```

Progress lives in [`tasks/todo.md`](tasks/todo.md) and
[`docs/ROADMAP.md`](docs/ROADMAP.md).

## License

[MIT](LICENSE), OpenForge, 2026.
