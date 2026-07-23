# Architecture

Design goal: the core knows the phases; drivers know the frameworks. Adding
support for a new stack means implementing one detector plus one driver and
registering it. The pattern is borrowed from Nx plugins, Buck2's
language-agnostic core, and Cloud Native Buildpacks' detect phase.

```
        core: pipeline
        detect, confirm, deps, analyze, test, build, sign, upload
        (guided TUI, unified error surfacing)
                    |                 |
             internal/detect    internal/driver (registry)
                    |                 |
          markers -> confidence   Flutter  ReactNative  Android  iOS ...
```

## Components

### internal/detect

Each detector inspects a directory and returns a confidence score from marker
files.

| Marker | Stack |
|--------|-------|
| `pubspec.yaml` with a `flutter:` block or `sdk: flutter` | Flutter (else pure Dart) |
| `package.json` with a `react-native` or `expo` dependency | React Native (bare or Expo) |
| `settings.gradle(.kts)` | native Android build root |
| `*.xcodeproj`, `*.xcworkspace`, `Podfile`, `Package.swift` | native iOS |

Detectors run over a depth-limited walk. The scanner prunes on detect (it does
not descend into a detected root), so a Gradle multi-module build is one project
and the `android/` and `ios/` folders inside a Flutter or React Native app are
attributed to that parent, never reported as separate stacks. A containment
sweep is the safety net. Directories like `node_modules`, `Pods`, `build`, and
`.dart_tool` are skipped.

### internal/driver

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
    Capabilities() Capabilities
}
```

A driver shells out to that ecosystem's real toolchain (flutter, gradle,
xcodebuild, npm) and returns normalized, structured results so the pipeline can
surface errors uniformly. That normalization, turning raw tool output into one
consistent plain-language error model, is the actual product.

### internal/pipeline

Language-agnostic orchestrator: pick the driver(s) via the registry, query
`Capabilities()` to decide which phases and prompts apply (skip signing for a web
build), run phases in order, stop cleanly on failure, and drive the interactive
UI. It knows the phases, nothing about any framework.

## Implementation notes

- Go, single static binary. CLI via Cobra. The guided TUI (Charm stack: Bubble
  Tea plus Huh) lands with the build lifecycle in Milestone 2.
- Drivers are compiled in first; they can be externalized as plugins later
  without touching the core.
- Distribution (later): GoReleaser to a Homebrew tap, Scoop, and `curl | sh`.
