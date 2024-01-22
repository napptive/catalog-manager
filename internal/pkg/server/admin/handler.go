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

package admin

import (
	"context"
	"fmt"

	"github.com/napptive/grpc-catalog-common-go"
	"github.com/napptive/grpc-catalog-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	manager Manager
}

func NewHandler(manager Manager) *Handler {
	return &Handler{manager: manager}
}

// Delete a namespace so that the applications contained on it are not longer available.
func (h *Handler) Delete(_ context.Context, request *grpc_catalog_go.DeleteNamespaceRequest) (*grpc_catalog_common_go.OpResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}
	err := h.manager.DeleteNamespace(request.Namespace)
	if err != nil {
		log.Warn().Err(err).Str("namespace", request.Namespace).Msg("unable to delete namespace")
		return nil, err
	}
	return &grpc_catalog_common_go.OpResponse{
		Status:     grpc_catalog_common_go.OpStatus_SUCCESS,
		StatusName: grpc_catalog_common_go.OpStatus_SUCCESS.String(),
		UserInfo:   fmt.Sprintf("namespace %s has been removed", request.Namespace),
	}, nil
}

// DeleteApplication deletes an application
func (h *Handler) DeleteApplication(_ context.Context, request *grpc_catalog_go.RemoveApplicationRequest) (*grpc_catalog_common_go.OpResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	if err := h.manager.DeleteApplication(request.ApplicationId); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	return &grpc_catalog_common_go.OpResponse{
		Status:     grpc_catalog_common_go.OpStatus_SUCCESS,
		StatusName: grpc_catalog_common_go.OpStatus_SUCCESS.String(),
		UserInfo:   fmt.Sprintf("Application [%s] removed", request.ApplicationId),
	}, nil
}

// List all the applications in the catalog
func (h *Handler) List(_ context.Context, request *grpc_catalog_go.ListApplicationsRequest) (*grpc_catalog_go.ApplicationList, error) {
	returned, err := h.manager.List(request.Namespace)
	if err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	summaryList := make([]*grpc_catalog_go.ApplicationSummary, 0)
	for _, app := range returned {
		summaryList = append(summaryList, app.ToApplicationSummary())
	}

	return &grpc_catalog_go.ApplicationList{Applications: summaryList}, nil
}
