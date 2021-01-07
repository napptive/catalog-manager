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

package dummy

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/napptive/go-template/internal/pkg/config"

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

// Run method starting the internal components and launching the service
func (s *Service) Run() {
	if err := s.cfg.IsValid(); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration options")
	}
	s.cfg.Print()
	s.registerShutdownListener()
	// Substitute ticker loop with proper code
	for now := range time.Tick(time.Minute) {
		fmt.Println(now, "alive")
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
