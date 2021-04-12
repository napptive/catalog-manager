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
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/provider/metadata"
	"github.com/napptive/catalog-manager/internal/pkg/server/catalog-manager"
	"github.com/napptive/catalog-manager/internal/pkg/storage"
	"github.com/napptive/njwt/pkg/interceptors"
	"golang.org/x/net/context"
	"net/http"
	"syscall"

	"github.com/napptive/grpc-catalog-go"

	njwtConfig "github.com/napptive/njwt/pkg/config"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
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

	providers, err := s.getProviders()
	if err != nil {
		log.Fatal().Err(err).Msg("error creating providers")
	}

	// launch services
	go s.LaunchHTTPService()
	s.LaunchGRPCService(providers)

}
// LaunchGRPCService launches a server for gRPC requests.
func (s *Service) LaunchGRPCService(providers *Providers) {
	manager := catalog_manager.NewManager(providers.repoStorage, providers.elasticProvider, s.cfg.CatalogUrl)
	handler := catalog_manager.NewHandler(manager)

	// create gRPC server
	var gRPCServer *grpc.Server
	if s.cfg.AuthEnabled {
		// interceptor
		config := njwtConfig.JWTConfig{
			Secret: s.cfg.JWTConfig.Secret,
			Header: s.cfg.JWTConfig.Header,
		}

		gRPCServer = grpc.NewServer(grpc.UnaryInterceptor(interceptors.JwtInterceptor(config)))
	}else {
		gRPCServer = grpc.NewServer()
	}

	grpc_catalog_go.RegisterCatalogServer(gRPCServer, handler)

	if s.cfg.Debug {
		// Register reflection service on gRPC server.
		reflection.Register(gRPCServer)
	}

	listener := s.getNetListener(s.cfg.Port)
	// start the service
	if err := gRPCServer.Serve(listener); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
}

// withCORSSupport creates a handler that supports CORS related preflights.
func (s *Service) withCORSSupport(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// LaunchHTTPService launches a server for HTTP requests.
func (s *Service) LaunchHTTPService() {
	mux := runtime.NewServeMux()
	grpcAddress := fmt.Sprintf(":%d", s.cfg.Port)
	grpcOptions := []grpc.DialOption{grpc.WithInsecure()}

	if err := grpc_catalog_go.RegisterCatalogHandlerFromEndpoint(context.Background(), mux, grpcAddress, grpcOptions); err != nil {
		log.Fatal().Err(err).Msg("failed to start catalog handler")
	}

	server := &http.Server{
		Addr:    grpcAddress,
		Handler: s.withCORSSupport(mux),
	}
	log.Info().Str("address", grpcAddress).Msg("HTTP Listening")
	// start the service
	if err := server.ListenAndServe(); err != nil {
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
