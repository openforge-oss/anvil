// Package upload pushes a signed artifact to its store or registry. It is a
// standalone subsystem (Android upload is an in-process API call, not a shell
// step), used by the anvil upload command.
package upload

import "context"

type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
	PlatformNPM     Platform = "npm"
)

// Uploader pushes one artifact to one destination. Validate checks required
// inputs and credentials; Describe returns a human summary for dry-run; Run
// performs the upload.
type Uploader interface {
	Platform() Platform
	Validate() error
	Describe() []string
	Run(ctx context.Context) error
}
