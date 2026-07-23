package upload

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

type Android struct {
	ServiceAccount string
	Package        string
	Track          string
	Artifact       string
}

func (Android) Platform() Platform { return PlatformAndroid }

func (u Android) Validate() error {
	switch {
	case u.ServiceAccount == "":
		return fmt.Errorf("Android upload needs a service account JSON (--service-account, GOOGLE_APPLICATION_CREDENTIALS, or PLAY_SERVICE_ACCOUNT_BASE64)")
	case u.Package == "":
		return fmt.Errorf("Android upload needs --package (the applicationId)")
	case u.Artifact == "":
		return fmt.Errorf("Android upload needs --artifact <aab>")
	}
	return nil
}

func (u Android) Describe() []string {
	return []string{
		fmt.Sprintf("Play edit: insert for %s", u.Package),
		fmt.Sprintf("upload bundle: %s", u.Artifact),
		fmt.Sprintf("assign track %q (status draft)", u.track()),
		"commit edit",
	}
}

func (u Android) Run(ctx context.Context) error {
	data, err := os.ReadFile(u.ServiceAccount)
	if err != nil {
		return err
	}
	conf, err := google.JWTConfigFromJSON(data, androidpublisher.AndroidpublisherScope)
	if err != nil {
		return fmt.Errorf("service account: %w", err)
	}
	svc, err := androidpublisher.NewService(ctx, option.WithTokenSource(conf.TokenSource(ctx)))
	if err != nil {
		return err
	}
	edit, err := svc.Edits.Insert(u.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("edits.insert: %w", err)
	}
	f, err := os.Open(u.Artifact)
	if err != nil {
		return err
	}
	defer f.Close()
	bundle, err := svc.Edits.Bundles.Upload(u.Package, edit.Id).Media(f).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("bundles.upload: %w", err)
	}
	_, err = svc.Edits.Tracks.Update(u.Package, edit.Id, u.track(), &androidpublisher.Track{
		Track: u.track(),
		Releases: []*androidpublisher.TrackRelease{{
			Status:       "draft",
			VersionCodes: []int64{bundle.VersionCode},
		}},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("tracks.update: %w", err)
	}
	if _, err := svc.Edits.Commit(u.Package, edit.Id).Context(ctx).Do(); err != nil {
		return fmt.Errorf("edits.commit: %w", err)
	}
	return nil
}

func (u Android) track() string {
	if u.Track == "" {
		return "internal"
	}
	return u.Track
}
