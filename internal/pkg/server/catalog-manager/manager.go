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
	"fmt"
	"regexp"
	"strings"

	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/provider/metadata"
	"github.com/napptive/catalog-manager/internal/pkg/storage"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
)

const (
	// readmeFile with the name of the readme file
	readmeFile = "readme.md"
	// NamespaceRegex with the regular expression to match namespaces.
	// * Must not contain two consecutive hyphens
	// * Must be lowercase
	// * Cannot start or end with hyphen
	NamespaceRegex = "^[a-z0-9]+([a-z0-9-{1}][a-z0-9]+)+[a-z0-9]?$"
)

// validNamespace regex parser.
var validNamespace = regexp.MustCompile(NamespaceRegex)

type Manager interface {
	// Add stores a new application in the repository.
	Add(requestedAppID string, files []*entities.FileInfo, isPrivate bool, username string) (bool, error)
	// Download returns the files of an application
	Download(requestedAppID string, compressed bool, username string) ([]*entities.FileInfo, error)
	// Remove removes an application from the repository
	Remove(requestedAppID string) error
	// Get returns a given application metadata
	Get(requestedAppID string, username string) (*entities.ExtendedApplicationMetadata, error)
	// List returns a list of applications (without metadata and readme content)
	List(namespace string, username string) ([]*entities.AppSummary, error)
	// Summary returns catalog summary
	Summary() (*entities.Summary, error)
}

type manager struct {
	stManager storage.StorageManager
	provider  metadata.MetadataProvider
	// catalogURL is the URL of the repository managed by this catalog
	catalogURL string
}

// NewManager returns a new object of manager
func NewManager(stManager storage.StorageManager, provider metadata.MetadataProvider, catalogURL string) Manager {
	return &manager{
		stManager:  stManager,
		provider:   provider,
		catalogURL: catalogURL,
	}
}

// getApplicationMetadataFile looks for the application metadata yaml file
func (m *manager) getApplicationMetadataFile(files []*entities.FileInfo) ([]byte, *entities.ApplicationMetadata, error) {
	for _, file := range files {
		// the files must have .yaml extension
		if utils.IsYamlFile(strings.ToLower(file.Path)) {
			// 2.- Get Metadata
			isMetadata, metadataObj, err := utils.IsMetadata(file.Data)
			if err != nil {
				log.Error().Err(err).Str("file", file.Path).Msg("Error looking for the metadata file")
				return nil, nil, nerrors.NewInternalErrorFrom(err, "error in %s file", file.Path)
			}
			if isMetadata {
				return file.Data, metadataObj, nil
			}
		}
	}
	return nil, nil, nil
}

// Add stores a new application in the repository returning the application visibility
func (m *manager) Add(requestedAppID string, files []*entities.FileInfo, isPrivate bool, username string) (bool, error) {

	// TODO: here, validate the application

	// Store metadata into the provider
	// Locate README and metadata
	// Store in the provider
	url, appID, err := utils.DecomposeApplicationID(requestedAppID)
	if err != nil {
		log.Err(err).Str("name", requestedAppID).Msg("Error decomposing the application identifier")
		return false, err
	}

	// Validate namespace
	if !validNamespace.Match([]byte(appID.Namespace)) {
		return false, nerrors.NewFailedPreconditionError("Invalid namespace, must contain lowercase letters, can contain single hyphens and numbers.")
	}

	// if catalogURL is not empty, check it!
	if m.catalogURL != "" {
		// check that the url of the application matches the url of the catalog
		// we avoid including applications in catalogs that do not correspond
		if url != "" && url != m.catalogURL {
			log.Err(err).Str("name", requestedAppID).Msg("Error adding application. The application url does not match the one in the catalog")
			return false, nerrors.NewInternalError("The application url does not match the one in the catalog")
		}
	}

	readme := utils.GetFile(readmeFile, files)
	appMetadata, header, err := m.getApplicationMetadataFile(files)
	if err != nil {
		return false, nerrors.NewInternalErrorFrom(err, "Unable to add the application. Error getting metadata")
	}
	// the metadata file is required, if is not in the Files -> return an error
	if appMetadata == nil {
		return false, nerrors.NewNotFoundError("Unable to add the application. Metadata file is required.")
	}
	// Metadata Name is required too
	if header == nil || header.Name == "" {
		return false, nerrors.NewFailedPreconditionError("Unable to add the application. Metadata name is required.")
	}

	// If authentication is enabled
	if username != "" {
		// Check Application visibility
		private, err := m.provider.GetApplicationVisibility(appID.Namespace, appID.ApplicationName)
		if err != nil {
			return false, nerrors.NewInternalErrorFrom(err, "Unable to add the application.")
		}

		if private != nil {
			log.Debug().Bool("application visibility", *private).Bool("new app visibility", isPrivate).Msg("checking application visibility")
			// the application stored is public and the user wants to store another version PRIVATE -> error
			if *private == false && isPrivate {
				return false, nerrors.NewInternalError("error adding application. There is already a public application, change the visibility before adding a private one.")
			} else {
				isPrivate = *private
			}
		} else {
			log.Debug().Bool("new app visibility", isPrivate).Msg("There is no applications previously")
		}
	}

	if _, err := m.provider.Add(&entities.ApplicationInfo{
		Namespace:       appID.Namespace,
		ApplicationName: appID.ApplicationName,
		Tag:             appID.Tag,
		Readme:          string(readme),
		Metadata:        string(appMetadata),
		MetadataName:    header.Name,
		Private:         isPrivate,
	}); err != nil {
		log.Err(err).Str("name", requestedAppID).Msg("Error storing application metadata")
		return false, err
	}

	// store the files into the repository storage
	if err = m.stManager.StoreApplication(appID.Namespace, appID.ApplicationName, appID.Tag, files); err != nil {
		log.Err(err).Str("name", requestedAppID).Msg("Error storing application")
		// rollback operation
		if rErr := m.provider.Remove(appID); rErr != nil {
			log.Err(err).Interface("appID", appID).Msg("Error in rollback operation, metadata can not be removed")
		}
		return false, err
	}

	return isPrivate, nil
}

// Download returns the files of an application
func (m *manager) Download(requestedAppID string, compressed bool, username string) ([]*entities.FileInfo, error) {

	_, appID, err := utils.DecomposeApplicationID(requestedAppID)
	if err != nil {
		return nil, nerrors.NewFailedPreconditionErrorFrom(err, "unable to download the application")
	}
	if username != "" {
		// If the application is private and the username is the application owner -> error
		app, err := m.provider.Get(appID)
		if err != nil {
			log.Error().Err(err).Str("application", requestedAppID).Str("username", username).Msg("error getting application to check its visibility")
			return nil, nerrors.NewInternalErrorFrom(err, "Error downloading application")
		}
		if app.Private {
			if app.Namespace != username {
				log.Error().Err(err).Str("application", requestedAppID).Str("username", username).
					Msg("error downloading application. The application is private and the user is not the owner")
				return nil, nerrors.NewPermissionDeniedError("Error downloading application")
			}
		}
	}
	return m.stManager.GetApplication(appID.Namespace, appID.ApplicationName, appID.Tag, compressed)
}

// Remove removes an application from the repository
func (m *manager) Remove(requestedAppID string) error {

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

// Get returns the application metadata for a given application
func (m *manager) Get(requestedAppID string, username string) (*entities.ExtendedApplicationMetadata, error) {

	_, appID, err := utils.DecomposeApplicationID(requestedAppID)
	if err != nil {
		return nil, err
	}

	app, err := m.provider.Get(appID)
	if err != nil {
		return nil, err
	}

	if username != "" && app.Private && username != app.Namespace {
		log.Error().Str("username", username).Str("application", requestedAppID).Msg("User trying to get info of a private app")
		return nil, nerrors.NewPermissionDeniedError("error getting application")
	}

	_, metadata, err := utils.IsMetadata([]byte(app.Metadata))
	if err != nil {
		return nil, err
	}
	var obj entities.ApplicationMetadata
	if metadata != nil {
		obj = *metadata
	}

	return &entities.ExtendedApplicationMetadata{
		CatalogID:       app.CatalogID,
		Namespace:       app.Namespace,
		ApplicationName: app.ApplicationName,
		Tag:             app.Tag,
		Readme:          app.Readme,
		Metadata:        app.Metadata,
		MetadataObj:     obj,
	}, nil
}

// List returns a list of applications (without metadata and readme content)
// List ([catalogURL/]namespace)
func (m *manager) List(namespace string, username string) ([]*entities.AppSummary, error) {
	// TODO: Check if the catalogURL matches with repositoryName
	// DecomposeApplicationID needs [catalogURL/]namespace/appName[:tag]
	// in this case we have no appName, uses dummyAppName to simulate it
	_, appID, err := utils.DecomposeApplicationID(fmt.Sprintf("%s/dummyAppName", namespace))
	if err != nil {
		return nil, nerrors.NewFailedPreconditionErrorFrom(err, "unable to list applications")
	}

	log.Debug().Str("required namespace", appID.Namespace).Str("username", username).Msg("application list")

	// no authentication enabled
	if username == "" {
		if namespace == "" {
			return m.provider.GetPublicApps(), nil
		} else {
			apps, _, err := m.provider.ListSummaryWithFilter(&metadata.ListFilter{
				Namespace: &appID.Namespace,
				Private:   nil,
			})
			if err != nil {
				log.Error().Err(err).Str("required namespace", appID.Namespace).Str("username", username).Msg("error getting public apps")
				return nil, err
			}
			return apps, nil
		}
	}

	// Authentication enabled
	// - namespace empty -> All public applications and his private ones
	// - namespace != empty
	//   - namespace == username -> All the applications in namespace (public and private)
	//   - namespace != username -> Public applications in namespace
	if appID.Namespace == "" {
		// All public apps + own privates (only if authEnabled username != "")
		private := true
		ownApps, _, err := m.provider.ListSummaryWithFilter(&metadata.ListFilter{
			Namespace: &username,
			Private:   &private,
		})
		if err != nil {
			log.Error().Err(err).Str("required namespace", appID.Namespace).Str("username", username).Msg("error getting public apps")
			return nil, err
		}
		return append(m.provider.GetPublicApps(), ownApps...), nil

	} else {
		if appID.Namespace == username {
			// the required namespace is the user namespace or authentication is not enabled -> Return all the applications in the namespace (public and private)
			apps, _, err := m.provider.ListSummaryWithFilter(&metadata.ListFilter{
				Namespace: &appID.Namespace,
				Private:   nil,
			})
			if err != nil {
				log.Error().Err(err).Str("required namespace", appID.Namespace).Str("username", username).Msg("error getting public apps")
				return nil, err
			}
			return apps, nil
		} else {
			// return only the public application in the namespace
			private := false
			apps, _, err := m.provider.ListSummaryWithFilter(&metadata.ListFilter{
				Namespace: &appID.Namespace,
				Private:   &private,
			})
			if err != nil {
				log.Error().Err(err).Str("required namespace", appID.Namespace).Str("username", username).Msg("error getting public apps")
				return nil, err
			}
			return apps, nil
		}
	}
}

// Summary returns catalog summary
func (m *manager) Summary() (*entities.Summary, error) {
	return m.provider.GetSummary()
}
