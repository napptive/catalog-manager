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
	grpc_catalog_go "github.com/napptive/grpc-catalog-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/yaml"
	"strings"
)

const (
	defaultVersion = "latest"
	readmeFile     = "readme.md"
	appFile        = "app_cfg.yaml"
	apiVersion     = "core.oam.dev/v1alpha2"
	kind           = "ApplicationMetadata"
)

type Manager struct {
	stManager *storage.StorageManager
	provider  provider.MetadataProvider
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

// getFile looks for a file by name in the array retrieved and return the data
func (m *Manager) getFile(fileName string, files []grpc_catalog_go.FileInfo) []byte {

	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file.Path), strings.ToLower(fileName)) {
			return file.Data
		}
	}

	return []byte{}
}

// AppHeader is a struct to load the kind and apiversion of a file to check if it is an applicationConfiguration
type AppHeader struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// getApplicationMetadataFile looks for the application metadata yaml file
func (m *Manager) getApplicationMetadataFile(files []grpc_catalog_go.FileInfo) []byte {

	for _, file := range files {
		// 1.- the files must have .yaml extension
		if strings.HasSuffix(strings.ToLower(file.Path), "yaml") {
			var a AppHeader
			// 2.- apiVersion: core.oam.dev/v1alpha2 and
			// 3.- kind: ApplicationConfiguration
			err := yaml.Unmarshal(file.Data, &a)
			if err == nil {
				log.Debug().Interface("App", a).Str("name", file.Path).Msg("appconfig")
				if a.APIVersion == apiVersion && a.Kind == kind {
					log.Debug().Interface("App", a).Str("name", file.Path).Msg("appconfig FOUND")
					return file.Data
				}
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
	readme := m.getFile(readmeFile, files)
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
