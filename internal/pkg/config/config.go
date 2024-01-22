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
	"github.com/rs/zerolog/log"
)

// Config structure with all the options required by the service and service components.
type Config struct {
	// CatalogManager with all the config parameters related to catalog-manager component
	CatalogManager
	// JWTConfig with the JWT specific configuration.
	JWTConfig
	// TeamConfig with the team configuration (repos and users)
	TeamConfig
	// BQConfig contains the configuration to connect to BigQuery
	BQConfig
	// TLS configuration
	TLSConfig
	// Playground connection configuration.
	PlaygroundConnection
	// Version of the application.
	Version string
	// Commit related to this built.
	Commit string
	// Debug flag.
	Debug bool
}

// IsValid checks if the configuration options are valid.
func (c *Config) IsValid() error {
	if err := c.CatalogManager.IsValid(); err != nil {
		return err
	}
	if err := c.JWTConfig.IsValid(); err != nil {
		return err
	}
	if err := c.TeamConfig.IsValid(); err != nil {
		return err
	}
	if err := c.BQConfig.IsValid(); err != nil {
		return err
	}
	if err := c.TLSConfig.IsValid(); err != nil {
		return err
	}
	if err := c.PlaygroundConnection.IsValid(); err != nil {
		return err
	}
	return nil
}

// Print the configuration using the application logger.
func (c *Config) Print() {
	// Use logger to print the configuration
	log.Info().Str("version", c.Version).Str("commit", c.Commit).Msg("Application config")
	c.CatalogManager.Print()
	c.JWTConfig.Print()
	c.TeamConfig.Print()
	c.BQConfig.Print()
	c.TLSConfig.Print()
	c.PlaygroundConnection.Print()
}
