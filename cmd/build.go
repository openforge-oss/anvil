package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/openforge-oss/anvil/internal/detect"
	"github.com/openforge-oss/anvil/internal/driver"
	"github.com/openforge-oss/anvil/internal/pipeline"
	"github.com/openforge-oss/anvil/internal/tui"
)

var (
	buildPath    string
	buildTarget  string
	buildFlavor  string
	buildRelease bool
	buildDryRun  bool
	buildPlain   bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Run a detected project through deps, analyze, test, and build",
	Args:  cobra.NoArgs,
	RunE:  runBuild,
}

func init() {
	buildCmd.Flags().StringVar(&buildPath, "path", ".", "project directory")
	buildCmd.Flags().StringVar(&buildTarget, "target", "", "build target (stack-specific, e.g. apk, appbundle, ios)")
	buildCmd.Flags().StringVar(&buildFlavor, "flavor", "", "build flavor, product flavor, or scheme")
	buildCmd.Flags().BoolVar(&buildRelease, "release", false, "release build where applicable")
	buildCmd.Flags().BoolVar(&buildDryRun, "dry-run", false, "print the steps without running them")
	buildCmd.Flags().BoolVar(&buildPlain, "plain", false, "plain line output instead of the interactive view")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, _ []string) error {
	root, err := filepath.Abs(buildPath)
	if err != nil {
		return err
	}

	projects, err := detect.Scan(root, detect.DefaultOptions())
	if err != nil {
		return err
	}
	chosen, err := chooseProject(cmd, projects, root)
	if err != nil {
		return err
	}

	d, ok := driver.For(chosen.Stack, chosen.Path)
	if !ok {
		return fmt.Errorf("anvil build does not support the %q stack yet", chosen.Stack)
	}

	opts := driver.BuildOptions{Target: buildTarget, Flavor: buildFlavor, Release: buildRelease}

	if buildDryRun {
		printPlan(cmd, chosen, d, opts)
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	events := make(chan pipeline.Event)
	go pipeline.Run(ctx, chosen.Path, d, opts, events)

	var success bool
	if usePlain() {
		success = tui.Plain(events, cmd.OutOrStdout())
	} else {
		success, err = tui.Run(events)
		if err != nil {
			return err
		}
	}
	if !success {
		return errors.New("build failed")
	}
	return nil
}

func chooseProject(cmd *cobra.Command, projects []detect.Project, root string) (detect.Project, error) {
	switch len(projects) {
	case 0:
		return detect.Project{}, errors.New("no supported project detected; pass --path to a project directory")
	case 1:
		return projects[0], nil
	}
	for _, p := range projects {
		if p.Path == root {
			return p, nil
		}
	}
	out := cmd.ErrOrStderr()
	fmt.Fprintln(out, "Multiple projects detected; pass --path to one of:")
	for _, p := range projects {
		fmt.Fprintf(out, "  %s (%s)\n", p.Path, p.Stack)
	}
	return detect.Project{}, errors.New("multiple projects detected")
}

func printPlan(cmd *cobra.Command, p detect.Project, d driver.Driver, opts driver.BuildOptions) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s (%s) at %s\n\n", d.Name(), p.Subtype, p.Path)
	for _, it := range pipeline.Plan(d, opts) {
		if it.Step == nil {
			fmt.Fprintf(out, "%s: skipped\n", it.Phase)
			continue
		}
		dir := it.Step.Dir
		if dir == "" {
			dir = "."
		}
		fmt.Fprintf(out, "%s: %s\n    dir: %s\n    cmd: %s\n", it.Phase, it.Name, dir, strings.Join(it.Step.Argv, " "))
	}
}

func usePlain() bool {
	if buildPlain || os.Getenv("CI") != "" {
		return true
	}
	return !term.IsTerminal(int(os.Stdout.Fd()))
}
