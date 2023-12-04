/**
 * Copyright 2023 Napptive
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

package config

import (
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
)

// CatalogManager with the catalog-manager configuration
type CatalogManager struct {
	// GRPCPort with the port on which the service will be listening
	GRPCPort int
	// HTTPPort with the port on which the service will be listening for HTTP connections.
	HTTPPort int
	// AdminAPI determines in the administration API is to be launched.
	AdminAPI bool
	// AdminGRPCPort with the port on which the administration interface will be listening.
	AdminGRPCPort int
	// ElasticAddress with the address to connect to Elastic
	ElasticAddress string
	// Index with the name of the elastic index
	Index string
	// RepositoryPath with the path of the repository
	RepositoryPath string
	//CatalogUrl with the url of the repository (napptive repository must be nil)
	CatalogUrl string
	// UseZoneAwareInterceptors determines if the service will be using an interceptor that uses
	// a backend to retrieve zone signing secrets.
	UseZoneAwareInterceptors bool
	// SecretsProviderAddress with the address of the service providing JWT signing secrets.
	SecretsProviderAddress string
}

// IsValid checks if the configuration options are valid.
func (c *CatalogManager) IsValid() error {
	if c.GRPCPort <= 0 {
		return nerrors.NewFailedPreconditionError("invalid gRPC port number")
	}
	if c.HTTPPort <= 0 {
		return nerrors.NewFailedPreconditionError("invalid HTTP port number")
	}
	if c.ElasticAddress == "" {
		return nerrors.NewFailedPreconditionError("ElasticAddress must be filled")
	}
	if c.Index == "" {
		return nerrors.NewFailedPreconditionError("Index must be filled")
	}
	if c.RepositoryPath == "" {
		return nerrors.NewFailedPreconditionError("RepositoryPath must be filled")
	}
	if c.AdminAPI {
		if c.AdminGRPCPort <= 0 {
			return nerrors.NewFailedPreconditionError("invalid admin gRPC port number")
		}
	}
	if c.UseZoneAwareInterceptors {
		if c.SecretsProviderAddress == "" {
			return nerrors.NewFailedPreconditionError("secretsProviderAddress must be set")
		}
	}
	return nil
}

// Print the configuration using the application logger.
func (c *CatalogManager) Print() {
	log.Info().Int("gRPC", c.GRPCPort).Int("HTTP", c.HTTPPort).Msg("ports")
	adminLog := log.Info().Bool("enabled", c.AdminAPI)
	if c.AdminAPI {
		adminLog.Int("gRPC", c.AdminGRPCPort)
	}
	adminLog.Msg("admin API")
	log.Info().Str("ElasticAddress", c.ElasticAddress).Str("Index", c.Index).Msg("Elastic Search Address")
	log.Info().Str("CatalogUrl", c.CatalogUrl).Msg("Catalog URL")
	log.Info().Str("RepositoryPath", c.RepositoryPath).Msg("Repository base path")
	log.Info().Bool("useZoneAwareInterceptors", c.UseZoneAwareInterceptors).Str("secretsProviderAddress", c.SecretsProviderAddress).Msg("JWT interceptors")
}
