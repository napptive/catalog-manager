/*
 Copyright 2022 Napptive

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package config

import "github.com/rs/zerolog/log"

type PlaygroundConnection struct {
	// UseTLS indicates that a TLS connection is expected with the Catalog Manager.
	UseTLS bool
	// SkipCertValidation flag that enables ignoring the validation step of the certificate presented by the server.
	SkipCertValidation bool
	// ClientCA with a client trusted CA
	ClientCA string
}

// IsValid checks if the configuration options are valid.
func (pc *PlaygroundConnection) IsValid() error {
	return nil
}

// Print the configuration using the application logger.
func (pc *PlaygroundConnection) Print() {
	log.Info().Bool("useTLS", pc.UseTLS).Bool("skipCertValidation", pc.SkipCertValidation).Bool("use_client_ca", len(pc.ClientCA) > 0).Msg("Connection options")
}
