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
package config

import (
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
)

// CatalogManager with the catalog-manager configuration
type CatalogManager struct {
	// Port with the port on which the service will be listening
	Port uint
	// ElasticAddress with the address to connect to Elastic
	ElasticAddress string
	// Index with the name of the elastic index
	Index string
	// RepositoryPath with the path of the repository
	RepositoryPath string
}

func (c *CatalogManager) IsValid() error {
	if c.Port <= 0 {
		return nerrors.NewFailedPreconditionError("invalid port number")
	}
	if c.ElasticAddress == "" {
		return nerrors.NewFailedPreconditionError("ElasticAddress must be filled")
	}

	return nil
}

func (c *CatalogManager) Print() {
	log.Info().Uint("Port", c.Port).Msg("grpc Port")
	log.Info().Str("ElasticAddress", c.ElasticAddress).Str("Index", c.Index).Msg("Elastic Search Address")
	log.Info().Str("RepositoryPath", c.RepositoryPath).Msg("Repository base path")
}
