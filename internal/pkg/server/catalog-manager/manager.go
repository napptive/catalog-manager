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
	"github.com/napptive/catalog-manager/internal/pkg/provider"
	"github.com/napptive/grpc-catalog-manager-go"
)

type Manager struct {
	ManagerProvider *provider.ManagerProvider
}

// NewManager returns a new object of manager
func NewManager(provider *provider.ManagerProvider) *Manager {
	return &Manager{ManagerProvider: provider}
}

func (m *Manager) List() (*grpc_catalog_manager_go.CatalogEntryListResponse, error) {
	catalog := m.ManagerProvider.GetCatalog()
	list := make([]*grpc_catalog_manager_go.CatalogEntryResponse, 0)

	for _, component := range catalog {
		list = append(list, component)
	}

	return &grpc_catalog_manager_go.CatalogEntryListResponse{
		Entries: list,
	}, nil
}
func (m *Manager) Get(catalogId string, entryId string) (*grpc_catalog_manager_go.CatalogEntryResponse, error) {
	return m.ManagerProvider.GetComponent(catalogId, entryId)
}
