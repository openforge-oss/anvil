// Package driver defines the per-ecosystem lifecycle contract that every
// supported stack implements:
//
//	Detect(dir) -> confidence
//	InstallDeps(ctx) / Analyze(ctx) / Test(ctx) / Build(ctx) -> Result
//	Sign(ctx) / Upload(ctx)        -> Result   // later phases
//	Capabilities()                 -> what this driver supports
//
// Each driver shells out to that ecosystem's real toolchain (flutter, gradle,
// xcodebuild, npm, …) and returns normalized results so the pipeline can
// surface errors uniformly. Adding a framework = adding one driver.
// Implemented in the tool-build phase — see docs/ARCHITECTURE.md.
package driver
