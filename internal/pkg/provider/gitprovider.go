/**
 * Copyright 2020 Napptive
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package provider

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	sshgit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/napptive/nerrors/pkg/nerrors"
)

// GitProvider with the client responsible for connecting to Git
type GitProvider struct {
	// Name with the name of the provider
	name string
	// url with the git url of the repository
	url string
	// sshPath with the path of the publicKey
	sshPath string
	// dir with the directory to clone the repo
	dir string
	// wt with the worktree of the cloned repo
	wt *git.Worktree
	// PublicKey with the key to clone and pull the repo
	PublicKey *sshgit.PublicKeys
}

// NewGitProvider returns a new git provider object
func NewGitProvider(name, url, sshPath, dir string) (CatalogProvider, error) {
	gp := GitProvider{
		name:    name,
		url:     url,
		sshPath: sshPath,
		dir:     dir,
		wt:      nil,
	}
	if err := gp.init(); err != nil {
		return nil, err
	}
	return &gp, nil
}
func (g *GitProvider) init() error {
	sshKey, _ := ioutil.ReadFile(g.sshPath)
	publicKey, keyError := sshgit.NewPublicKeys("git", []byte(sshKey), "")
	if keyError != nil {
		log.Error().Str("error", keyError.Error()).Msg("error getting public key")
		return nerrors.NewInternalError(keyError.Error())
	}
	// Ignore known_hosts (accept any host key)
	publicKey.HostKeyCallbackHelper = sshgit.HostKeyCallbackHelper{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	g.PublicKey = publicKey
	path := fmt.Sprintf("%s/%s/", g.dir, g.name)
	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      g.url,
		Progress: os.Stdout,
		Auth:     publicKey,
	})
	if err != nil {
		log.Error().Str("err", err.Error()).Msg("unable to clone the repo")
		return nerrors.NewInternalError(keyError.Error())
	}

	wt, err := repo.Worktree()
	if err != nil {
		log.Error().Str("err", err.Error()).Msg("unable to get the worktree")
		return nerrors.NewInternalError(keyError.Error())
	}
	g.wt = wt

	return nil
}

func (g *GitProvider) GetName() string {
	return g.name
}

// GetComponents get the components from the repo and returns them
func (g *GitProvider) GetComponents() ([]CatalogEntry, error) {
	if g.wt == nil {
		return nil, nerrors.NewFailedPreconditionError("Unable to load components, init the client first")
	}
	// git pull

	if err:= g.wt.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth: g.PublicKey,
	}); err != nil {
		log.Warn().Str("repo", g.url).
			Str("error", err.Error()).
			Msg("Error pulling the repo")
	}
	entries :=  g.loadComponents(".", nil)
	return entries, nil
}

func (g *GitProvider) loadComponents(path string, file os.FileInfo) []CatalogEntry {

	fullPath := path
	if file != nil {
		fullPath = filepath.Join(path, file.Name())
	}

	files, err := g.wt.Filesystem.ReadDir(filepath.Join(fullPath))
	if err != nil {
		log.Warn().Str("path", path).Str("file", file.Name()).
			Str("err", err.Error()).Msg("unable to read the worktree")
	}
	var result []CatalogEntry
	for _, file := range files {
		if !file.IsDir() {
			if  utils.IsYamlFile(file.Name()) {
				path := fmt.Sprintf("%s/%s/%s/%s", g.dir, g.name, fullPath, file.Name())
				isComponent, component, err := utils.IsComponent(path)
				if err != nil {
					log.Warn().Str("error", err.Error()).Str("path", path).Msg("error getting component")
				}else{
					if isComponent {
						result = append(result, CatalogEntry{
							EntryId:        fmt.Sprintf("%s:%s/%s",g.name, fullPath, file.Name()),
							Component: component,
						})
					}
				}
			}
		} else {
			result = append(result, g.loadComponents(fullPath, file)...)
		}
	}
	return result

}

// RemoveRepo removes the cloned repository
func (g *GitProvider) EmptyCache() error {
	err := os.RemoveAll(g.dir)
	if err != nil {
		log.Error().Str("err", err.Error()).Msg("unable to remove the repo")
		return err
	}

	return nil
}
