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
package connection

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"

	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/nerrors/pkg/nerrors"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// GetConnectionToPlayground creates a new connection with the playground.
func GetConnectionToPlayground(cfg *config.PlaygroundConnection, targetPlaygroundApiURL string) (*grpc.ClientConn, error) {
	if cfg.UseTLS {
		return GetTLSConnection(cfg, targetPlaygroundApiURL)
	}
	return GetNonTLSConnection(targetPlaygroundApiURL)
}

// GetTLSConnection returns a TLS wrapped connection with the playground server.
func GetTLSConnection(cfg *config.PlaygroundConnection, address string) (*grpc.ClientConn, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipCertValidation,
	}
	if cfg.ClientCA != "" {
		cp := x509.NewCertPool()
		decoded, err := base64.StdEncoding.DecodeString(cfg.ClientCA)
		if err != nil {
			return nil, nerrors.NewInternalErrorFrom(err, "error decoding CA")
		}
		if !cp.AppendCertsFromPEM(decoded) {
			return nil, nerrors.NewInternalError("Error appending CA")
		}
		// add the CA as valid one
		tlsConfig.RootCAs = cp
	}
	tlsCredentials := credentials.NewTLS(tlsConfig)
	return grpc.Dial(address, grpc.WithTransportCredentials(tlsCredentials))
}

// GetNonTLSConnection returns a plain connection with the playground server.
func GetNonTLSConnection(address string) (*grpc.ClientConn, error) {
	log.Warn().Msg("using insecure connection with the playground")
	return grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
