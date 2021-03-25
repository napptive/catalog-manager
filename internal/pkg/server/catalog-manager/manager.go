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

package catalog_manager

import (
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/provider"
	"github.com/napptive/catalog-manager/internal/pkg/storage"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	grpc_catalog_go "github.com/napptive/grpc-catalog-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"strings"
)

const (
	defaultVersion = "latest"
	readmeFile     = "readme.md"
	apiVersion     = "core.oam.dev/v1alpha2"
	kind           = "ApplicationMetadata"
)

type Manager struct {
	stManager *storage.StorageManager
	provider  provider.MetadataProvider
	// repositoryURL is the URL of the repository managed by this catalog
	repositoryURL string
}

// NewManager returns a new object of manager
func NewManager(manager *storage.StorageManager, provider provider.MetadataProvider) *Manager {
	return &Manager{
		stManager: manager,
		provider:  provider,
	}
}

// decomposeRepositoryName gets the url, repo, application and version from repository name
// [repoURL/]repoName/appName[:tag]
func (m *Manager) decomposeRepositoryName(name string) (*entities.ApplicationID, error) {
	var version string
	var applicationName string
	var repoName string
	urlName := ""

	names := strings.Split(name, "/")
	if len(names) != 2 && len(names) != 3 {
		return nil, nerrors.NewFailedPreconditionError(
			"incorrect format for application name. [repoURL/]repoName/appName[:tag]")
	}

	// if len == 2 -> no url informed.
	if len(names) == 3 {
		urlName = names[0]
	}
	repoName = names[len(names)-2]

	// get the version -> appName[:tag]
	sp := strings.Split(names[len(names)-1], ":")
	if len(sp) == 1 {
		applicationName = sp[0]
		version = defaultVersion
	} else if len(sp) == 2 {
		applicationName = sp[0]
		version = sp[1]
	} else {
		return nil, nerrors.NewFailedPreconditionError(
			"incorrect format for application name. [repoURL/]repoName/appName[:tag]")
	}

	return &entities.ApplicationID{
		Url:             urlName,
		Repository:      repoName,
		ApplicationName: applicationName,
		Tag:             version,
	}, nil
}

// getApplicationMetadataFile looks for the application metadata yaml file
func (m *Manager) getApplicationMetadataFile(files []grpc_catalog_go.FileInfo) []byte {

	for _, file := range files {
		// 1.- the files must have .yaml extension
		if utils.IsYamlFile(strings.ToLower(file.Path)) {
			// 2.- Get Metadata
			isMetadata := utils.CheckKindAndVersion(file.Data, apiVersion, kind)
			if isMetadata {
				log.Debug().Str("name", file.Path).Msg("Metadata found")
				return file.Data
			}
		}
	}
	return nil
}

// Add Adds a new application in the repository.
func (m *Manager) Add(name string, files []grpc_catalog_go.FileInfo) error {

	// TODO: here, validate the application

	// 1.- Store metadata into the provider
	// Locate README and appConfig
	// Store in the provider
	appID, err := m.decomposeRepositoryName(name)
	if err != nil {
		log.Err(err).Str("name", name).Msg("Error decomposing the application name")
		return err
	}

	// check that the url of the application matches the url of the catalog
	// we avoid including applications in catalogs that do not correspond
	if appID.Url != m.repositoryURL {
		log.Err(err).Str("name", name).Msg("Error adding application. The application url does not match the one in the catalog")
		return nerrors.NewInternalError("The application url does not match the one in the catalog")
	}

	readme := utils.GetFile(readmeFile, files)
	// appConfig := m.getFile(appFile, files)
	appMetadata := m.getApplicationMetadataFile(files)

	if err := m.provider.Add(entities.ApplicationMetadata{
		Url:             appID.Url,
		Repository:      appID.Repository,
		ApplicationName: appID.ApplicationName,
		Tag:             appID.Tag,
		Readme:          string(readme),
		Metadata:        string(appMetadata),
	}); err != nil {
		log.Err(err).Str("name", name).Msg("Error storing application metadata")
		return err
	}

	// 2.- store the files into the repository storage
	if err = m.stManager.StorageApplication(appID.Repository, appID.ApplicationName, appID.Tag, files); err != nil {
		log.Err(err).Str("name", name).Msg("Error storing application")
		return err
	}

	return nil
}
