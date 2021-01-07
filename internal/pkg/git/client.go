package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/napptive/nerrors/pkg"
)

// GitClient with the client responsible for connecting to Git
type GitClient struct {
	// url with the url of the repository
	url string
	// sshPath with the path of the publicKey
	sshPath string
	// dir with the directory to clone the repo
	dir string
	// wt with the worktree of the cloned repo
	wt *git.Worktree
}

func NewGitClient(url, sshPath, dir string) GitClient {
	return GitClient{
		url:     url,
		sshPath: sshPath,
		dir:     dir,
		wt:      nil,
	}
}

func (g *GitClient) Init() error {
	sshKey, _ := ioutil.ReadFile(g.sshPath)
	publicKey, keyError := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if keyError != nil {
		fmt.Println(keyError)
	}

	repo, err := git.PlainClone(g.dir, false, &git.CloneOptions{
		URL:      g.url,
		Progress: os.Stdout,
		Auth:     publicKey,
	})
	if err != nil {
		log.Error().Str("err", err.Error()).Msg("unable to clone the repo")
	}

	wt, err := repo.Worktree()
	if err != nil {
		log.Error().Str("err", err.Error()).Msg("unable to get the worktree")
	}
	g.wt = wt

	return nil
}

func (g *GitClient) LoadComponents() error {
	if g.wt == nil {
		return nerrors.NewFailedPreconditionError("Unable to load components, init the client first")
	}
	// git pull
	g.wt.Pull(&git.PullOptions{
		RemoteName: "origin",
	})

	files, err := g.wt.Filesystem.ReadDir(".")
	if err != nil {
		return nerrors.FromError(err)
	}
	for _, file := range files {
		g.getComponents(".", file)
	}
	return nil
}

func (g *GitClient) getComponents (path string, file os.FileInfo) {

	if file.IsDir(){
		newPath := filepath.Join(path, file.Name())
		files, err := g.wt.Filesystem.ReadDir(filepath.Join(newPath))
		if err != nil {
			log.Warn().Str("path", path).Str("file", file.Name()).
				Str("err", err.Error()).Msg("unable to read the worktree")
		}
		for _, file := range files {
			g.getComponents(newPath, file)
		}
	}else{
		if strings.HasSuffix(file.Name(), ".yaml") {
			log.Info().Str("path", path).Str("file", file.Name()).Msg("Component")
		}
	}

}