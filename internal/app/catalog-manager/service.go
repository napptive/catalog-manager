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
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/rs/zerolog/log"

	bqinterceptor "github.com/napptive/analytics/pkg/interceptors"
	analytics "github.com/napptive/analytics/pkg/provider"
	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/provider/metadata"
	"github.com/napptive/catalog-manager/internal/pkg/server/admin"
	"github.com/napptive/catalog-manager/internal/pkg/server/apps"
	catalog_manager "github.com/napptive/catalog-manager/internal/pkg/server/catalog-manager"
	"github.com/napptive/catalog-manager/internal/pkg/storage"
	grpc_catalog_go "github.com/napptive/grpc-catalog-go"
	njwtConfig "github.com/napptive/njwt/pkg/config"
	"github.com/napptive/njwt/pkg/interceptors"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	// analyticsProvider to store operation metrics
	analyticsProvider analytics.Provider
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

	if s.cfg.BQConfig.Enabled {
		provider, err := analytics.NewBigQueryProvider(s.cfg.BQConfig.Config)
		if err != nil {
			return nil, err
		}
		return &Providers{
			elasticProvider:   pr,
			repoStorage:       storage.NewStorageManager(s.cfg.RepositoryPath),
			analyticsProvider: provider}, nil
	}
	// ! s.cfg.BQConfig.Enabled
	return &Providers{elasticProvider: pr,
		repoStorage: storage.NewStorageManager(s.cfg.RepositoryPath)}, nil

}

// Run method starting the internal components and launching the service
func (s *Service) Run() {
	if err := s.cfg.IsValid(); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration options")
	}
	s.cfg.Print()
	providers, err := s.getProviders()
	if err != nil {
		log.Fatal().Err(err).Msg("error creating providers")
	}

	s.registerShutdownListener(providers)

	// launch services
	go s.LaunchHTTPService()
	if s.cfg.AdminAPI {
		go s.LaunchGRPCAdminService(providers)
	}
	s.LaunchGRPCService(providers)
}

// Launch the gRPC admin service.
func (s *Service) LaunchGRPCAdminService(providers *Providers) {
	manager := admin.NewManager(providers.repoStorage, providers.elasticProvider)
	handler := admin.NewHandler(manager)

	// No analytics exported for the administration service.
	gRPCServer := grpc.NewServer()

	grpc_catalog_go.RegisterNamespaceAdministrationServer(gRPCServer, handler)

	if s.cfg.Debug {
		// Register reflection service on gRPC server.
		reflection.Register(gRPCServer)
	}

	listener := s.getNetListener(s.cfg.AdminGRPCPort)
	// start the service
	if err := gRPCServer.Serve(listener); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
}

// LaunchGRPCService launches a server for gRPC requests.
func (s *Service) LaunchGRPCService(providers *Providers) {
	manager := catalog_manager.NewManager(providers.repoStorage, providers.elasticProvider, s.cfg.CatalogUrl)
	handler := catalog_manager.NewHandler(manager, s.cfg.AuthEnabled, s.cfg.TeamConfig)

	appManager := apps.NewManager(&s.cfg, manager)
	appHandler := apps.NewHandler(&s.cfg.JWTConfig, appManager)

	var unaryInterceptorsChain []grpc.UnaryServerInterceptor
	//var unaryStreamChain grpc.ServerOption
	var unaryStreamChain []grpc.StreamServerInterceptor

	// create gRPC server
	var gRPCServer *grpc.Server
	if s.cfg.AuthEnabled {
		// interceptor
		config := njwtConfig.JWTConfig{
			Secret: s.cfg.JWTConfig.Secret,
			Header: s.cfg.JWTConfig.Header,
		}
		unaryInterceptorsChain = append(unaryInterceptorsChain, interceptors.JwtInterceptor(config))
		//unaryStreamChain =  interceptors.WithServerJWTStreamInterceptor(config)
		unaryStreamChain = append(unaryStreamChain, interceptors.JwtStreamInterceptor(config))

	}
	if s.cfg.BQConfig.Enabled {
		log.Info().Msg("analytics enabled, create the interceptor")
		unaryInterceptorsChain = append(unaryInterceptorsChain, bqinterceptor.OpInterceptor(providers.analyticsProvider))
		unaryStreamChain = append(unaryStreamChain, bqinterceptor.OpStreamInterceptor(providers.analyticsProvider))
	}
	if s.cfg.TLSConfig.LaunchSecureService {
		serverCert, err := credentials.NewServerTLSFromFile(s.cfg.CertificatePath, s.cfg.PrivateKeyPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create gRPC cert")
		}
		gRPCServer = grpc.NewServer(grpc.Creds(serverCert), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptorsChain...)),
			grpc_middleware.WithStreamServerChain(unaryStreamChain...))
	} else {
		gRPCServer = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptorsChain...)),
			grpc_middleware.WithStreamServerChain(unaryStreamChain...))
	}
	grpc_catalog_go.RegisterCatalogServer(gRPCServer, handler)
	grpc_catalog_go.RegisterApplicationsServer(gRPCServer, appHandler)

	if s.cfg.Debug {
		// Register reflection service on gRPC server.
		reflection.Register(gRPCServer)
	}

	listener := s.getNetListener(s.cfg.GRPCPort)
	// start the service
	if err := gRPCServer.Serve(listener); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
}

// HealthzHandler to return 200 if called.
func (s *Service) HealthzHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	w.WriteHeader(http.StatusOK)
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
	grpcAddress := fmt.Sprintf(":%d", s.cfg.GRPCPort)
	var grpcOptions []grpc.DialOption
	if s.cfg.TLSConfig.LaunchSecureService {
		tlsConfig := &tls.Config{
			// Since the proxy may not be seeing the same DNS address, skip the verification for now.
			InsecureSkipVerify: true,
		}
		tlsCredentials := credentials.NewTLS(tlsConfig)
		grpcOptions = []grpc.DialOption{grpc.WithTransportCredentials(tlsCredentials)}
	} else {
		grpcOptions = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}

	if err := grpc_catalog_go.RegisterCatalogHandlerFromEndpoint(context.Background(), mux, grpcAddress, grpcOptions); err != nil {
		log.Fatal().Err(err).Msg("failed to start catalog handler")
	}

	if err := grpc_catalog_go.RegisterApplicationsHandlerFromEndpoint(context.Background(), mux, grpcAddress, grpcOptions); err != nil {
		log.Fatal().Err(err).Msg("failed to start applications handler")
	}

	if err := mux.HandlePath("GET", "/healthz", s.HealthzHandler); err != nil {
		log.Fatal().Err(err).Msg("unable to register healthz handler")
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.HTTPPort),
		Handler: s.withCORSSupport(mux),
	}

	// start the service
	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
}

func (s *Service) registerShutdownListener(providers *Providers) {
	osChannel := make(chan os.Signal, 2)
	signal.Notify(osChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-osChannel
		s.Shutdown(providers)
		os.Exit(1)
	}()
}

// Shutdown code
func (s *Service) Shutdown(providers *Providers) {
	log.Warn().Msg("shutting down service")
	if s.cfg.BQConfig.Enabled {
		providers.analyticsProvider.Flush()
	}
}

func (s *Service) getNetListener(port int) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}
	return lis
}
