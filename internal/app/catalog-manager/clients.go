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

package catalog_manager

import (
	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/grpc-jwt-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients to communicate with other services of the platform.
type Clients struct {
	JWTSecretsClient grpc_jwt_go.SecretsClient
}

// GetClients builds the set of clients to connect to the different internal services.
func GetClients(cfg *config.Config) (*Clients, error) {
	var jwtSecretsClient grpc_jwt_go.SecretsClient
	if cfg.CatalogManager.UseZoneAwareInterceptors {
		secretsConn, err := grpc.Dial(cfg.CatalogManager.SecretsProviderAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, nerrors.NewInternalErrorFrom(err, "cannot connect to JWT secrets provider")
		}
		jwtSecretsClient = grpc_jwt_go.NewSecretsClient(secretsConn)
	}

	return &Clients{
		JWTSecretsClient: jwtSecretsClient,
	}, nil
}
