// Package cmd wires up the anvil command-line interface.
//
// This is the minimal skeleton: it dispatches a couple of built-in commands so
// the binary builds and runs. The guided build/release pipeline (detect → deps
// → analyze → test → build → sign → upload) arrives in the tool-build phase;
// see docs/ROADMAP.md.
package cmd

import (
	"flag"
	"fmt"
	"os"
)

const usage = `anvil — a guided, zero-config build & release pipeline for mobile and app projects.

Usage:
  anvil <command> [flags]

Commands:
  version    Print the anvil version
  help       Show this help

anvil is under active development. See docs/ROADMAP.md for what's coming.
`

// Execute is the entry point for the anvil CLI.
func Execute() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	switch args[0] {
	case "version":
		fmt.Println(versionString())
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "anvil: unknown command %q\n\n", args[0])
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
