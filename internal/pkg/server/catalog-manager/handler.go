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

package catalog_manager

import (
	"context"

	"github.com/napptive/grpc-catalog-manager-go"
	"github.com/napptive/nerrors/pkg/nerrors"
)

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) *Handler {
	return &Handler{manager: manager}
}

// List the available components.
func (h *Handler) List(ctx context.Context, request *grpc_catalog_manager_go.ListComponentsRequest) (*grpc_catalog_manager_go.CatalogEntryListResponse, error){
	// Validate
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	return h.manager.List()
}

// Get retrieves a specific component
func (h *Handler) Get(ctx context.Context, request *grpc_catalog_manager_go.GetComponentRequest) (*grpc_catalog_manager_go.CatalogEntryResponse, error){

	// Validate
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	return h.manager.Get(request.CatalogId, request.EntryId)

}


