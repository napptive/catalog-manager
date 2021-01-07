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

package catalog_manager

import (
	"fmt"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/git"

	"github.com/rs/zerolog/log"

)

// Service structure in charge of launching the application.
type Service struct {
	cfg config.Config
}

// NewService creates a new service with a given configuration
func NewService(cfg config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}
/*
func (s *Service) test2() {
	//var hostKey ssh.PublicKey
	url := "git@github.com:napptive/catalog.git"
	//path := "/ssh/id_rsa"
	path := "/Users/cdelope/.ssh/id_rsa"
	dir := "./tmp/"


	sshKey, _ := ioutil.ReadFile(path)
	publicKey, keyError := trans.NewPublicKeys("git", []byte(sshKey), "")
	if keyError != nil {
		fmt.Println(keyError)
	}

	repo,err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Auth:     publicKey,
	})
	if err != nil {
		// already exits, pull
		log.Fatal().Str("err", err.Error()).Msg("unable to connect")
	}

	wt, _ := repo.Worktree()
	
	s.wt = wt

	s.LookforComponent()
}

func (s *Service) LookforComponent() {
	files, err := s.wt.Filesystem.ReadDir(".")
	if err != nil {
		log.Fatal().Str("err", err.Error()).Msg("unable to connect")
	}
	for _, file := range files {
		s.getComponent(".", file)
	}
}


func (s *Service) getComponent (path string, file os.FileInfo) {

	if file.IsDir(){
		newPath := filepath.Join(path, file.Name())
		files, err := s.wt.Filesystem.ReadDir(filepath.Join(newPath))
		if err != nil {
			log.Fatal().Str("err", err.Error()).Msg("unable to connect")
		}
		for _, file := range files {
			s.getComponent(newPath, file)
		}
	}else{
		if strings.HasSuffix(file.Name(), ".yaml") {
			log.Info().Str("path", path).Str("file", file.Name()).Msg("FILE")
		}
	}

}
*/
// Run method starting the internal components and launching the service
func (s *Service) Run() {
	if err := s.cfg.IsValid(); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration options")
	}
	s.cfg.Print()
	s.registerShutdownListener()

	url := "git@github.com:napptive/catalog.git"
	//path := "/ssh/id_rsa"
	path := "/Users/cdelope/.ssh/id_rsa"
	dir := "./tmp/"

	gitClient := git.NewGitClient(url, path, dir)
	err := gitClient.Init()
	if err != nil {
		log.Fatal().Str("err", err.Error()).Msg("unable to init the client")
	}
	// load components
	gitClient.LoadComponents()

	// Substitute ticker loop with proper code
	for now := range time.Tick(time.Minute) {
		fmt.Println(now, "alive")
		gitClient.LoadComponents()
	}
}

func (s *Service) registerShutdownListener() {
	osChannel := make(chan os.Signal)
	signal.Notify(osChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-osChannel
		s.Shutdown()
		os.Exit(1)
	}()
}

// Shutdown code
func (s *Service) Shutdown() {
	log.Warn().Msg("shutting down service")
}
