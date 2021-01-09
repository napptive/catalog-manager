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
	"fmt"
	"github.com/rs/zerolog/log"
)

// CatalogManager with the catalog-manager configuration
type CatalogManager struct {
	// Port with the port on which the service will be listening
	Port uint
	// ClonePath with the path where the repositories are going to be cloned
	ClonePath string
	// ConfigPath with the path with the repositories configurations
	ConfigPath string
	// Minutes between repositories pulls checking new components
	PullInterval int
}

func (c *CatalogManager) IsValid() error {
	if c.Port <= 0 {
		return fmt.Errorf("invalid port number")
	}
	if c.ClonePath == "" {
		return fmt.Errorf("clone path must be filled")
	}
	if c.ConfigPath == "" {
		return fmt.Errorf("config path must be filled")
	}
	if c.PullInterval == 0 {
		return fmt.Errorf("pull_interval must be more than 0 minutes")
	}
	return nil
}

func (c *CatalogManager) Print() {
	log.Info().Str("ClonePath", c.ClonePath).Msg("Path where to clone the repos")
	log.Info().Str("ConfigPath", c.ConfigPath).Msg("Repositories configuration path")
	log.Info().Int("PullInterval", c.PullInterval).Msg("Minutes between repositories pulls checking new components")
	log.Info().Uint("Port", c.Port).Msg("grpc Port")
}
