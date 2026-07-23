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
	buildSign    bool
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
	buildCmd.Flags().BoolVar(&buildSign, "sign", false, "set up signing and produce a signed artifact")
	buildCmd.Flags().BoolVar(&buildDryRun, "dry-run", false, "print the steps without running them")
	buildCmd.Flags().BoolVar(&buildPlain, "plain", false, "plain line output instead of the interactive view")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, _ []string) error {
	chosen, err := resolveProject(cmd, buildPath)
	if err != nil {
		return err
	}
	d, ok := driver.For(chosen.Stack, chosen.Path)
	if !ok {
		return fmt.Errorf("anvil build does not support the %q stack yet", chosen.Stack)
	}

	opts := driver.BuildOptions{Target: buildTarget, Flavor: buildFlavor, Release: buildRelease}
	phases := driver.Phases

	if buildSign {
		signing, extra, err := setupSigning(cmd, chosen, "", buildDryRun)
		if err != nil {
			return err
		}
		opts.Signing = signing
		phases = append(append([]driver.Phase{}, driver.Phases...), extra...)
	}

	if buildDryRun {
		printPlan(cmd, chosen, d, opts, phases)
		return nil
	}
	return runPipeline(cmd, chosen.Path, d, opts, phases)
}

func resolveProject(cmd *cobra.Command, path string) (detect.Project, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return detect.Project{}, err
	}
	projects, err := detect.Scan(root, detect.DefaultOptions())
	if err != nil {
		return detect.Project{}, err
	}
	return chooseProject(cmd, projects, root)
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

func runPipeline(cmd *cobra.Command, root string, d driver.Driver, opts driver.BuildOptions, phases []driver.Phase) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	events := make(chan pipeline.Event)
	go pipeline.RunPhases(ctx, root, d, opts, phases, events)

	var (
		success bool
		err     error
	)
	if usePlain() {
		success = tui.Plain(events, cmd.OutOrStdout())
	} else {
		success, err = tui.Run(events)
		if err != nil {
			return err
		}
	}
	if !success {
		return errors.New("one or more steps failed")
	}
	return nil
}

func printPlan(cmd *cobra.Command, p detect.Project, d driver.Driver, opts driver.BuildOptions, phases []driver.Phase) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s (%s) at %s\n\n", d.Name(), p.Subtype, p.Path)
	for _, it := range pipeline.PlanPhases(d, opts, phases) {
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
	if buildPlain || signPlain || os.Getenv("CI") != "" {
		return true
	}
	return !term.IsTerminal(int(os.Stdout.Fd()))
}
