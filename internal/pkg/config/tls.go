/**
 * Copyright 2021 Napptive
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

// TLSConfig configuration of the service.
type TLSConfig struct {
	// LaunchSecureService determines if the gRPC service with TLS enabled must be launched.
	LaunchSecureService bool
	// CertificatePath with the path of the server certificate.
	CertificatePath string
	// PrivateKeyPath with the path of the private key associated with the certificate.
	PrivateKeyPath string
}

// IsValid checks if the configuration options are valid.
func (t *TLSConfig) IsValid() error {
	if t.LaunchSecureService {
		if t.CertificatePath == "" {
			return nerrors.NewInvalidArgumentError("certificatePath cannot be empty")
		}
		if t.PrivateKeyPath == "" {
			return nerrors.NewInvalidArgumentError("privateKeyPath cannot be empty")
		}
	}
	return nil
}

// Print the configuration using the application logger.
func (t *TLSConfig) Print() {
	if t.LaunchSecureService {
		log.Info().Str("certPath", t.CertificatePath).Str("keyPath", t.PrivateKeyPath).Msg("secure gRPC server")
	} else {
		log.Info().Bool("launchSecureService", t.LaunchSecureService).Msg("secure gRPC server")
	}
}
