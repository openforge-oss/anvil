# Architecture

The design goal: **the core knows the *phases*; drivers know the *frameworks*.**
Adding support for a new stack = implementing one detector + one driver and
registering it. (Pattern borrowed from Nx plugins / Buck2's language-agnostic
core / Cloud Native Buildpacks' detect phase.)

```
        ┌──────────────────────────────────────────────┐
        │                  pipeline                     │
        │  detect → confirm → deps → analyze → test →   │
        │  build → sign → upload   (guided TUI, unified │
        │  error surfacing)                             │
        └───────────────┬───────────────┬──────────────┘
                        │               │
                 internal/detect   internal/driver (registry)
                        │               │
        markers → confidence     ┌───────┴────────┬───────────┐
                              Flutter   ReactNative   Android   iOS …
```

## Components

### `internal/detect`
Each detector inspects a directory and returns a **confidence score** from marker
files:

| Marker | Stack |
|--------|-------|
| `pubspec.yaml` | Flutter / Dart |
| `package.json` with a `react-native` dep | React Native |
| `build.gradle` / `settings.gradle` | native Android |
| `*.xcodeproj` / `*.xcworkspace` / `Podfile` | native iOS |

Detectors run independently; the pipeline resolves multiple hits (e.g. an RN app
with `android/` + `ios/` folders) by confidence and nesting.

### `internal/driver`
Every ecosystem implements one lifecycle contract (sketch):

```go
type Driver interface {
    Detect(dir string) (confidence float64)
    InstallDeps(ctx Context) Result
    Analyze(ctx Context) Result
    Test(ctx Context) Result
    Build(ctx Context) (Artifacts, error)
    Sign(ctx Context) Result     // phase 3
    Upload(ctx Context) Result   // phase 4
    Capabilities() Capabilities  // artifact types, supportsSigning, …
}
```

A driver just shells out to that ecosystem's real toolchain (`flutter`, `gradle`,
`xcodebuild`, `npm`) and returns **normalized, structured results**. That
normalization — turning raw tool output into one consistent, plain-language error
model — is the actual product.

### `internal/pipeline`
Language-agnostic orchestrator: pick driver(s) via the registry, query
`Capabilities()` to decide which phases/prompts apply (skip signing for a web
build), run phases in order, stop cleanly on failure, and drive the interactive
UI. Knows the phases, nothing about any framework.

## Implementation notes
- **Go**, single static binary. CLI via Cobra and the guided TUI via the Charm
  stack (Bubble Tea + Huh) land with the tool build (Milestone 2); the current
  skeleton is stdlib-only so it builds offline.
- Drivers are compiled-in first; can be externalized as plugins later without
  touching the core.
- Distribution (later): GoReleaser → Homebrew tap + Scoop + `curl | sh`.
