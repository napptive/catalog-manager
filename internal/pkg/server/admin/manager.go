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
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/provider/metadata"
	"github.com/napptive/catalog-manager/internal/pkg/storage"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
)

type Manager interface {
	// DeleteNamespace deletes a namespace so that the applications contained on it is no longer available.
	DeleteNamespace(namespace string) error
	// DeleteApplication removes an application from the repository
	DeleteApplication(requestedAppID string) error
	// List returns a list of applications (without metadata and readme content)
	List(namespace string) ([]*entities.AppSummary, error)
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

// DeleteNamespace deletes a namespace so that the applications contained on it is no longer available.
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

// DeleteApplication removes an application from the repository
func (m *manager) DeleteApplication(requestedAppID string) error {

	// - Validate the appName
	_, appID, err := utils.DecomposeApplicationID(requestedAppID)
	if err != nil {
		return nerrors.NewFailedPreconditionErrorFrom(err, "unable to remove application, wrong name")
	}

	// - Remove from metadata provider
	err = m.provider.Remove(appID)
	if err != nil {
		log.Err(err).Str("appName", requestedAppID).Msg("Unable to remove application metadata")
		return err
	}

	// - Remove from storage
	err = m.stManager.RemoveApplication(appID.Namespace, appID.ApplicationName, appID.Tag)
	if err != nil {
		log.Err(err).Str("requestedAppID", requestedAppID).Msg("Unable to remove application")
		return err
	}

	return nil
}

// List returns a list of applications (without metadata and readme content)
func (m *manager) List(namespace string) ([]*entities.AppSummary, error) {
	return m.provider.ListSummary(namespace)
}
