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

package admin

import (
	"github.com/napptive/catalog-manager/internal/pkg/provider/metadata"
	"github.com/napptive/catalog-manager/internal/pkg/storage"
)

type Manager interface {
	// DeleteNamespace deletes a namespace so that the applications contained on it are not longer available.
	DeleteNamespace(namespace string) error
}

type manager struct {
	stManager storage.StorageManager
	provider  metadata.MetadataProvider
}

func NewManager(stManager storage.StorageManager, metadataProvider metadata.MetadataProvider) Manager {
	return &manager{
		stManager: stManager,
		provider:  metadataProvider,
	}
}

// DeleteNamespace deletes a namespace so that the applications contained on it are not longer available.
func (m *manager) DeleteNamespace(namespace string) error {
	namespaceApps, err := m.provider.List(namespace)
	if err != nil {
		return err
	}
	// If the user does not have any application on its namespace, exit as success.
	if len(namespaceApps) == 0 {
		return nil
	}
	// First, remove the metadata.
	for _, app := range namespaceApps {
		if err := m.provider.Remove(app.ToApplicationID()); err != nil {
			return err
		}
	}
	// Finally, delete the repository
	return m.stManager.RemoveRepository(namespace)
}
