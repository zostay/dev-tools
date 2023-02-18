package plugin_github

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v49/github"
	"golang.org/x/oauth2"
)

type Github struct {
	gh *github.Client
}

func (g *Github) SetupGithubClient(ctx context.Context) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is missing")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	g.gh = github.NewClient(tc)

	return nil
}
