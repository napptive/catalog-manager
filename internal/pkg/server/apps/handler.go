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

package apps

import (
	"context"

	"github.com/napptive/catalog-manager/internal/pkg/config"
	grpc_catalog_common_go "github.com/napptive/grpc-catalog-common-go"
	grpc_catalog_go "github.com/napptive/grpc-catalog-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

// Handler for apps operations.
type Handler struct {
	manager Manager
	cfg     *config.JWTConfig
}

// NewHandler creates a new instance of the handler.
func NewHandler(cfg *config.JWTConfig, manager Manager) *Handler {
	return &Handler{
		manager: manager,
		cfg:     cfg,
	}
}

// extractIncomingJWT retrieves the JWT used by the user in this request. This token will be forwarded to the playground for request authentication.
func (h *Handler) extractIncomingJWT(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", nerrors.NewUnauthenticatedError("retrieving metadata failed")
	}
	token, ok := md[h.cfg.Header]
	if !ok {
		return "", nerrors.NewUnauthenticatedError("no auth details supplied")
	}
	if token[0] == "" {
		return "", nerrors.NewUnauthenticatedError("token not found, log into the playground before proceeding")
	}
	return token[0], nil
}

// Deploy an application on a target Playground platform. This endpoint
// will gather the application information and send it to the target
// playground platform.
func (h *Handler) Deploy(ctx context.Context, request *grpc_catalog_go.DeployApplicationRequest) (*grpc_catalog_common_go.OpResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}
	jwt, err := h.extractIncomingJWT(ctx)
	if err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}
	response, err := h.manager.Deploy(jwt, request.ApplicationId, request.TargetEnvironmentQualifiedName, request.TargetPlaygroundApiUrl, request.InstanceConfiguration)
	if err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}
	return response, nil
}

func (h *Handler) GetConfiguration(ctx context.Context, request *grpc_catalog_go.GetConfigurationRequest) (*grpc_catalog_go.GetConfigurationResponse, error) {

	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	conf, err := h.manager.GetConfiguration(request.ApplicationId)
	if err != nil {
		log.Error().Err(err).Msg("error getting application configuration")
		return nil, nerrors.FromError(err).ToGRPC()
	}
	return conf, nil
}
