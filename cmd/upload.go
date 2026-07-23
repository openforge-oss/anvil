package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/openforge-oss/anvil/internal/detect"
	"github.com/openforge-oss/anvil/internal/upload"
)

var (
	uploadPath     string
	uploadArtifact string
	uploadPlatform string
	uploadTrack    string
	uploadPackage  string
	uploadKeyID    string
	uploadIssuer   string
	uploadKeyPath  string
	uploadSA       string
	uploadYes      bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a signed artifact to its store or registry",
	Args:  cobra.NoArgs,
	RunE:  runUpload,
}

func init() {
	f := uploadCmd.Flags()
	f.StringVar(&uploadPath, "path", ".", "project directory")
	f.StringVar(&uploadArtifact, "artifact", "", "path to the signed artifact (ipa or aab)")
	f.StringVar(&uploadPlatform, "platform", "", "ios, android, or npm (default from the detected stack)")
	f.StringVar(&uploadTrack, "track", "internal", "Play track (android)")
	f.StringVar(&uploadPackage, "package", "", "applicationId / package name (android)")
	f.StringVar(&uploadKeyID, "api-key-id", "", "App Store Connect key id (ios)")
	f.StringVar(&uploadIssuer, "api-issuer-id", "", "App Store Connect issuer id (ios)")
	f.StringVar(&uploadKeyPath, "api-key-path", "", "App Store Connect .p8 path (ios)")
	f.StringVar(&uploadSA, "service-account", "", "Google Play service account JSON path (android)")
	f.BoolVar(&uploadYes, "yes", false, "perform the upload (default is a dry run)")
	rootCmd.AddCommand(uploadCmd)
}

func runUpload(cmd *cobra.Command, _ []string) error {
	root, err := filepath.Abs(uploadPath)
	if err != nil {
		return err
	}
	platform := upload.Platform(uploadPlatform)
	if platform == "" {
		chosen, err := resolveProject(cmd, uploadPath)
		if err != nil {
			return err
		}
		platform = platformForStack(chosen.Stack)
	}

	u, cleanup, err := buildUploader(root, platform, uploadYes)
	if err != nil {
		return err
	}
	defer cleanup()

	out := cmd.OutOrStdout()
	if !uploadYes {
		fmt.Fprintf(out, "Dry run (%s). Would:\n", platform)
		for _, line := range u.Describe() {
			fmt.Fprintf(out, "  %s\n", line)
		}
		fmt.Fprintln(out, "Re-run with --yes to upload.")
		return nil
	}

	if err := u.Validate(); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return u.Run(ctx)
}

func platformForStack(stack detect.Stack) upload.Platform {
	switch stack {
	case detect.IOS:
		return upload.PlatformIOS
	case detect.Flutter, detect.ReactNative, detect.Android:
		return upload.PlatformAndroid
	default:
		return upload.Platform(stack)
	}
}

func buildUploader(root string, platform upload.Platform, stage bool) (upload.Uploader, func(), error) {
	cleanup := func() {}
	switch platform {
	case upload.PlatformIOS:
		keyPath, cu := orEnv(uploadKeyPath, "ASC_KEY_PATH"), cleanup
		if stage {
			p, c, err := upload.FileCred(root, uploadKeyPath, "ASC_KEY_PATH", "ASC_KEY_P8_BASE64", "AuthKey.p8")
			if err != nil {
				return nil, cleanup, err
			}
			keyPath, cu = p, c
		}
		return upload.IOS{KeyID: orEnv(uploadKeyID, "ASC_KEY_ID"), IssuerID: orEnv(uploadIssuer, "ASC_ISSUER_ID"), KeyPath: keyPath, Artifact: uploadArtifact}, cu, nil
	case upload.PlatformAndroid:
		saPath, cu := orEnv(uploadSA, "GOOGLE_APPLICATION_CREDENTIALS"), cleanup
		if stage {
			p, c, err := upload.FileCred(root, uploadSA, "GOOGLE_APPLICATION_CREDENTIALS", "PLAY_SERVICE_ACCOUNT_BASE64", "play-sa.json")
			if err != nil {
				return nil, cleanup, err
			}
			saPath, cu = p, c
		}
		return upload.Android{ServiceAccount: saPath, Package: uploadPackage, Track: uploadTrack, Artifact: uploadArtifact}, cu, nil
	case upload.PlatformNPM:
		return upload.NPM{Dir: root}, cleanup, nil
	default:
		return nil, cleanup, fmt.Errorf("unsupported upload platform %q", platform)
	}
}

func orEnv(flag, env string) string {
	if flag != "" {
		return flag
	}
	return os.Getenv(env)
}
