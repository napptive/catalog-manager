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
}

func (c *CatalogManager) IsValid() error {
	if c.Port <= 0 {
		return fmt.Errorf("invalid port number")
	}
	return nil
}

func (c *CatalogManager) Print() {
	log.Info().Uint("Port", c.Port).Msg("grpc Port")
}
