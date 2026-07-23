// Package pipeline is the guided orchestrator: detect the stack, confirm with
// the user, then run each lifecycle phase in order (deps → analyze → surface
// errors → test → build → sign → upload), stopping cleanly on failure and
// reporting in plain language rather than dumping raw tool output.
//
// The core knows the phases but nothing about any specific framework — all
// ecosystem knowledge lives in drivers. Implemented in the tool-build phase;
// see docs/ARCHITECTURE.md.
package pipeline
