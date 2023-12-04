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
	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/connection"
	"github.com/napptive/catalog-manager/internal/pkg/server/catalog-manager"
	"github.com/napptive/grpc-catalog-common-go"
	"github.com/napptive/grpc-catalog-go"
	"github.com/napptive/grpc-playground-apps-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/napptive/oam-utils/pkg/oam-utils"
	"github.com/rs/zerolog/log"
)

type Manager interface {
	// Deploy an application on a target Playground platform. This endpoint
	// will gather the application information and send it to the target
	// playground platform.
	Deploy(userToken string, applicationID string, targetEnvironmentQualifiedName string, targetPlaygroundApiURL string,
		instanceConfiguration map[string]*grpc_catalog_go.ApplicationInstanceConfiguration, allowed bool) (*grpc_catalog_common_go.OpResponse, error)
	// GetConfiguration returns the application configuration (name for now)
	GetConfiguration(applicationID string, allowed bool) (*grpc_catalog_go.GetConfigurationResponse, error)
}

// Manager for apps operations.
type manager struct {
	catalogManager catalog_manager.Manager
	contextHelper  *connection.ContextHelper
	cfg            *config.Config
}

// NewManager creates a new instance of the manager.
func NewManager(cfg *config.Config, catalogManager catalog_manager.Manager) Manager {
	contextHelper := connection.NewContextHelper(cfg)
	return &manager{
		catalogManager: catalogManager,
		contextHelper:  contextHelper,
		cfg:            cfg,
	}
}

// Deploy an application on a target Playground platform. This endpoint
// will gather the application information and send it to the target
// playground platform.
func (m *manager) Deploy(userToken string, applicationID string, targetEnvironmentQualifiedName string, targetPlaygroundApiURL string,
	instanceConfiguration map[string]*grpc_catalog_go.ApplicationInstanceConfiguration, allowed bool) (*grpc_catalog_common_go.OpResponse, error) {

	// Download the application
	app, err := m.catalogManager.Download(applicationID, true, allowed)
	if err != nil {
		log.Error().Err(err).Str("application_id", applicationID).Msg("error downloading the application, unable to deploy it")
		return nil, err
	}

	// GetConnection
	conn, err := connection.GetConnectionToPlayground(&m.cfg.PlaygroundConnection, targetPlaygroundApiURL)
	if err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "cannot establish connection with Playground server on %s", targetPlaygroundApiURL)
	}
	defer conn.Close()
	client := grpc_playground_apps_go.NewAppsClient(conn)

	// Create a connection with the same JWT token
	ctx, cancel := m.contextHelper.GetContext(userToken)
	defer cancel()

	response, err := client.Deploy(ctx, &grpc_playground_apps_go.DeployApplicationRequest{
		DeployFrom:                     grpc_playground_apps_go.DeploySource_FROM_DATA,
		ApplicationDataType:            grpc_playground_apps_go.AppDataType_TGZ,
		ApplicationData:                app[0].Data,
		TargetEnvironmentQualifiedName: targetEnvironmentQualifiedName,
		InstanceConfiguration:          m.toInstanceConfiguration(instanceConfiguration),
		RepoConf:                       nil,
	})
	if err != nil {
		return nil, nerrors.FromGRPC(err)
	}

	return &grpc_catalog_common_go.OpResponse{
		Status:     grpc_catalog_common_go.OpStatus_SUCCESS,
		StatusName: grpc_catalog_common_go.OpStatus_SUCCESS.String(),
		UserInfo:   response.Message,
	}, nil
}

func (m *manager) toInstanceConfiguration(instanceConfiguration map[string]*grpc_catalog_go.ApplicationInstanceConfiguration) map[string]*grpc_playground_apps_go.ApplicationInstanceConfiguration {
	newConf := make(map[string]*grpc_playground_apps_go.ApplicationInstanceConfiguration)
	for appName, conf := range instanceConfiguration {
		newConf[appName] = &grpc_playground_apps_go.ApplicationInstanceConfiguration{
			ApplicationName:   conf.ApplicationDefaultName,
			SpecComponentsRaw: conf.SpecComponentsRaw,
		}
	}
	return newConf
}

// GetConfiguration returns the application configuration
func (m *manager) GetConfiguration(applicationID string, allowed bool) (*grpc_catalog_go.GetConfigurationResponse, error) {

	log.Debug().Str("application_id", applicationID).Bool("allowed", allowed).Msg("getting configuration")
	// Download the application
	files, err := m.catalogManager.Download(applicationID, false, allowed)
	if err != nil {
		log.Error().Err(err).Str("application_id", applicationID).Msg("error downloading the application, unable to get application configuration")
		return nil, err
	}

	appFiles := make([]*oam_utils.ApplicationFile, 0)

	// Get Application configuration
	for _, file := range files {
		appFiles = append(appFiles, &oam_utils.ApplicationFile{FileName: file.Path, Content: file.Data})
	}

	app, err := oam_utils.NewApplication(appFiles)
	if err != nil {
		// check the error, perhaps the catalog application no correspond to an oam application
		log.Error().Err(err).Str("applicationID", applicationID).Msg("error getting application files")
		return nil, nerrors.NewInternalErrorFrom(err, "error getting application configuration")

	}

	appConfig, err := app.GetConfigurations()
	if err != nil {
		// check the error, perhaps the catalog application no correspond to an oam application
		log.Error().Err(err).Str("applicationID", applicationID).Msg("error getting application parameters")
		return nil, err

	}
	if len(appConfig) > 1 {
		log.Warn().Str("applicationID", applicationID).Int("applications", len(appConfig)).Msg("Application with more than one oam applications")
	}
	// There is no oam application, the catalog application contains another entities
	if len(appConfig) == 0 {
		return &grpc_catalog_go.GetConfigurationResponse{
			IsApplication:          false,
			ApplicationDefaultName: "",
			SpecComponentsRaw:      "",
		}, nil
	}
	defaultName := ""
	spec := ""
	for _, conf := range appConfig {
		defaultName = conf.Name
		spec = conf.ComponentSpec
		break
	}

	return &grpc_catalog_go.GetConfigurationResponse{
		IsApplication:          true,
		ApplicationDefaultName: defaultName,
		SpecComponentsRaw:      spec,
	}, nil
}
