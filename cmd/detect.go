package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/openforge-oss/anvil/internal/detect"
	"github.com/spf13/cobra"
)

var (
	detectPath  string
	detectJSON  bool
	detectDepth int
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect the stack(s) in a project directory",
	Args:  cobra.NoArgs,
	RunE:  runDetect,
}

func init() {
	detectCmd.Flags().StringVar(&detectPath, "path", ".", "directory to scan")
	detectCmd.Flags().BoolVar(&detectJSON, "json", false, "output JSON")
	detectCmd.Flags().IntVar(&detectDepth, "depth", 4, "maximum directory depth to scan")
	rootCmd.AddCommand(detectCmd)
}

func runDetect(cmd *cobra.Command, _ []string) error {
	opts := detect.DefaultOptions()
	opts.MaxDepth = detectDepth

	projects, err := detect.Scan(detectPath, opts)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()

	if detectJSON {
		if projects == nil {
			projects = []detect.Project{}
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(projects)
	}

	if len(projects) == 0 {
		fmt.Fprintln(out, "No supported stack detected.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tSTACK\tSUBTYPE\tCONFIDENCE")
	for _, p := range projects {
		subtype := p.Subtype
		if len(p.Flags) > 0 {
			subtype += " [" + strings.Join(p.Flags, ",") + "]"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\n", p.Path, p.Stack, subtype, p.Confidence)
	}
	return w.Flush()
}
