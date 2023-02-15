package git

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

var ignoreStatus = map[string]struct{}{
	".session.vim": {},
}

type Git struct {
	repo   *git.Repository
	remote *git.Remote
	wc     *git.Worktree
}

func ref(t, n string) plumbing.ReferenceName {
	return plumbing.ReferenceName(path.Join("refs", t, n))
}

func refSpec(r plumbing.ReferenceName) gitConfig.RefSpec {
	sr := string(r)
	return gitConfig.RefSpec(strings.Join([]string{sr, sr}, ":"))
}

func (g *Git) SetupGitRepo() error {
	l, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("unable to open git repository at .: %w", err)
	}

	g.repo = l

	r, err := g.repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("unable to connect to remote origin: %w", err)
	}

	g.remote = r

	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("unable to examine the working copy: %w", err)
	}

	g.wc = w

	return nil
}
