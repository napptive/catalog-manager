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

package connection

import (
	"context"
	"time"

	"github.com/napptive/catalog-manager/internal/pkg/config"
	"google.golang.org/grpc/metadata"
)

const (
	// ContextTimeout with the default timeout for operations with the playground.
	ContextTimeout = 1 * time.Minute
	// AuthorizationHeader with the key name for the authorization payload.
	AuthorizationHeader = "authorization"
	// AgentHeader with the key name for the agent payload.
	AgentHeader = "agent"
	// AgentValue with the value for the agent payload.
	AgentValue = "catalog-manager"
	// VersionHeader with the key name for the version payload.
	VersionHeader = "version"
)

// ContextHelper structure to faciliate the generation of secure contexts.
type ContextHelper struct {
	// Version of the application sending the request.
	Version string
}

// NewContextHelper creates a ContextHelper with a given configuration.
func NewContextHelper(cfg *config.Config) *ContextHelper {
	return &ContextHelper{
		Version: cfg.Version,
	}
}

// GetContext returns a valid gRPC context with the appropiate authorization header.
func (ch *ContextHelper) GetContext(JWT string) (context.Context, context.CancelFunc) {
	md := metadata.New(map[string]string{AuthorizationHeader: JWT, AgentHeader: AgentValue, VersionHeader: ch.Version})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	return context.WithTimeout(ctx, ContextTimeout)
}
