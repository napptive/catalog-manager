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
	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/provider/metadata"
	"github.com/napptive/catalog-manager/internal/pkg/server/catalog-manager"
	"github.com/napptive/catalog-manager/internal/pkg/storage"

	"github.com/napptive/grpc-catalog-go"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"syscall"
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

// Providers with all the providers needed
type Providers struct {
	// elasticProvider with a elastic provider to store metadata
	elasticProvider metadata.MetadataProvider
	// repoStorage to store the applications
	repoStorage storage.StorageManager
}

// getProviders creates and initializes all the providers
func (s *Service) getProviders() (*Providers, error) {
	pr, err := metadata.NewElasticProvider(s.cfg.Index, s.cfg.ElasticAddress)
	if err != nil {
		return nil, err
	}
	err = pr.Init()
	if err != nil {
		return nil, err
	}
	return &Providers{elasticProvider: pr,
		repoStorage: storage.NewStorageManager(s.cfg.RepositoryPath)}, nil
}

// Run method starting the internal components and launching the service
func (s *Service) Run() {
	if err := s.cfg.IsValid(); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration options")
	}
	s.cfg.Print()
	s.registerShutdownListener()

	listener := s.getNetListener(s.cfg.Port)

	providers, err := s.getProviders()
	if err != nil {
		log.Fatal().Err(err).Msg("error creating providers")
	}

	manager := catalog_manager.NewManager(providers.repoStorage, providers.elasticProvider, s.cfg.CatalogUrl)
	handler := catalog_manager.NewHandler(manager)

	// create gRPC server
	gRPCServer := grpc.NewServer()

	grpc_catalog_go.RegisterCatalogServer(gRPCServer, handler)

	if s.cfg.Debug {
		// Register reflection service on gRPC server.
		reflection.Register(gRPCServer)
	}
	// start the service
	if err := gRPCServer.Serve(listener); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
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

func (s *Service) getNetListener(port uint) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}
	return lis
}
