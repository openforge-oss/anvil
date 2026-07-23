package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/openforge-oss/anvil/internal/detect"
	"github.com/openforge-oss/anvil/internal/driver"
	"github.com/openforge-oss/anvil/internal/sign"
)

var (
	signPath     string
	signKeystore string
	signKeyAlias string
	signTeamID   string
	signMethod   string
	signPlatform string
	signDryRun   bool
	signPlain    bool
)

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Set up release signing for a detected project",
	Args:  cobra.NoArgs,
	RunE:  runSign,
}

func init() {
	signCmd.Flags().StringVar(&signPath, "path", ".", "project directory")
	signCmd.Flags().StringVar(&signKeystore, "keystore", "", "keystore path (Android; generated if missing)")
	signCmd.Flags().StringVar(&signKeyAlias, "key-alias", "upload", "keystore key alias (Android)")
	signCmd.Flags().StringVar(&signTeamID, "team-id", "", "Apple Developer Team ID (iOS)")
	signCmd.Flags().StringVar(&signMethod, "export-method", "development", "iOS export method: development or ad-hoc")
	signCmd.Flags().StringVar(&signPlatform, "platform", "", "android or ios (default from the detected stack)")
	signCmd.Flags().BoolVar(&signDryRun, "dry-run", false, "print what would happen without changing anything")
	signCmd.Flags().BoolVar(&signPlain, "plain", false, "plain line output instead of the interactive view")
	rootCmd.AddCommand(signCmd)
}

func runSign(cmd *cobra.Command, _ []string) error {
	chosen, err := resolveProject(cmd, signPath)
	if err != nil {
		return err
	}
	signing, extra, err := setupSigning(cmd, chosen, signPlatform, signDryRun)
	if err != nil {
		return err
	}
	d, ok := driver.For(chosen.Stack, chosen.Path)
	if !ok {
		return fmt.Errorf("signing not supported for the %q stack", chosen.Stack)
	}
	opts := driver.BuildOptions{Signing: signing}

	if len(extra) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Signing configured. Run 'anvil build --release' to produce a signed artifact.")
		return nil
	}
	if signDryRun {
		printPlan(cmd, chosen, d, opts, extra)
		return nil
	}
	return runPipeline(cmd, chosen.Path, d, opts, extra)
}

// setupSigning resolves signing config and performs the guided setup. When
// dryRun is true it makes no changes and generates no secrets.
func setupSigning(cmd *cobra.Command, p detect.Project, platformFlag string, dryRun bool) (driver.Signing, []driver.Phase, error) {
	switch resolvePlatform(platformFlag, p.Stack) {
	case "android":
		return setupAndroidSigning(cmd, p, dryRun)
	case "ios":
		return setupIOSSigning(cmd, p, dryRun)
	default:
		return driver.Signing{}, nil, fmt.Errorf("signing not supported for the %q stack", p.Stack)
	}
}

func resolvePlatform(flag string, stack detect.Stack) string {
	if flag != "" {
		return flag
	}
	switch stack {
	case detect.IOS:
		return "ios"
	case detect.Flutter, detect.ReactNative, detect.Android:
		return "android"
	default:
		return ""
	}
}

func setupAndroidSigning(cmd *cobra.Command, p detect.Project, dryRun bool) (driver.Signing, []driver.Phase, error) {
	gradleRoot, appBuildFile, kotlinDSL, ok := sign.AndroidLayout(p.Path)
	if !ok {
		return driver.Signing{}, nil, errors.New("could not find an Android app module (android/app or app)")
	}
	keystore := signKeystore
	if keystore == "" {
		keystore = filepath.Join(gradleRoot, "upload-keystore.jks")
	}
	out := cmd.OutOrStdout()

	if dryRun {
		fmt.Fprintf(out, "Would set up Android signing:\n")
		fmt.Fprintf(out, "  generate keystore: %s\n", keystore)
		fmt.Fprintf(out, "  write: %s\n", filepath.Join(gradleRoot, "key.properties"))
		if kotlinDSL {
			fmt.Fprintf(out, "  wire (manual, Kotlin DSL): %s\n", appBuildFile)
		} else {
			fmt.Fprintf(out, "  wire signingConfigs: %s\n", appBuildFile)
		}
		fmt.Fprintf(out, "  gitignore: %v\n", sign.SecretPatterns)
		return driver.Signing{}, nil, nil
	}

	storePass, err := resolveSecret("ANVIL_STORE_PASS", "Keystore password")
	if err != nil {
		return driver.Signing{}, nil, err
	}
	ks := sign.Keystore{Path: keystore, Alias: signKeyAlias, StorePass: storePass, KeyPass: storePass, DName: "CN=anvil, O=anvil, C=US"}
	if err := sign.EnsureGitignore(p.Path, sign.SecretPatterns); err != nil {
		return driver.Signing{}, nil, err
	}
	if err := sign.GenerateKeystore(ks); err != nil {
		return driver.Signing{}, nil, err
	}
	if _, err := sign.WriteKeyProperties(gradleRoot, ks); err != nil {
		return driver.Signing{}, nil, err
	}
	wired, err := sign.WireAndroidGradle(appBuildFile, kotlinDSL)
	if err != nil {
		return driver.Signing{}, nil, err
	}
	if !wired {
		fmt.Fprintf(out, "Kotlin DSL not auto-wired. Add a signingConfig reading key.properties to %s\n", appBuildFile)
	}
	return driver.Signing{}, nil, nil
}

func setupIOSSigning(cmd *cobra.Command, p detect.Project, dryRun bool) (driver.Signing, []driver.Phase, error) {
	teamID := signTeamID
	if teamID == "" {
		teamID = os.Getenv("ANVIL_TEAM_ID")
	}
	plist := filepath.Join(p.Path, "ExportOptions.plist")

	if dryRun {
		if teamID == "" {
			teamID = "<TEAM_ID>"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Would write %s (method=%s, teamID=%s) and gitignore it\n", plist, signMethod, teamID)
		return driver.Signing{TeamID: teamID, ExportMethod: signMethod, ExportPlist: plist}, []driver.Phase{driver.Sign}, nil
	}

	if teamID == "" {
		if !interactive() {
			return driver.Signing{}, nil, errors.New("set --team-id or ANVIL_TEAM_ID")
		}
		if err := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Apple Developer Team ID").Value(&teamID),
		)).Run(); err != nil {
			return driver.Signing{}, nil, err
		}
	}
	if _, err := sign.WriteExportOptions(p.Path, teamID, signMethod); err != nil {
		return driver.Signing{}, nil, err
	}
	if err := sign.EnsureGitignore(p.Path, sign.SecretPatterns); err != nil {
		return driver.Signing{}, nil, err
	}
	return driver.Signing{TeamID: teamID, ExportMethod: signMethod, ExportPlist: plist}, []driver.Phase{driver.Sign}, nil
}

func resolveSecret(env, title string) (string, error) {
	if v := os.Getenv(env); v != "" {
		return v, nil
	}
	if !interactive() {
		return "", fmt.Errorf("set %s or run interactively to provide the %s", env, title)
	}
	var v string
	if err := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title(title).EchoMode(huh.EchoModePassword).Value(&v),
	)).Run(); err != nil {
		return "", err
	}
	if v == "" {
		return "", fmt.Errorf("%s is required", title)
	}
	return v, nil
}

func interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}
