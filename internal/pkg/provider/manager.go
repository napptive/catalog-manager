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
package provider

import (
	"encoding/json"
	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/grpc-catalog-manager-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"sync"
	"time"
)

// ManagerProvider defines the object responsible for catalog provider management
type ManagerProvider struct {
	// clonePath with the path where the catalog repositories have to be cloned
	clonePath string
	// Providers indexes by CatalogId (config.Name)
	providers map[string]CatalogProvider
	// conf with all the provider configurations
	conf      config.ConfList
	// catalog with a map wit all the components in the catalog
	catalog   map[string]grpc_catalog_manager_go.CatalogEntryResponse
	// TODO: Hace falta?
	sync.Mutex
}

// NewManagerProvider creates a new provider manager
func NewManagerProvider(configFile string,  cloneDir string) (*ManagerProvider, error) {

	reader, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Err(err).Msg("error reading the config")
		return nil, nerrors.FromError(err)
	}
	var configList config.ConfList
	err = json.Unmarshal(reader, &configList)
	if err != nil {
		log.Err(err).Msg("error getting the config list")
		return nil, nerrors.FromError(err)
	}

	return &ManagerProvider{
		conf: configList,
		clonePath: cloneDir,
		providers: map[string]CatalogProvider{},
	}, nil
}

// Init creates providers based on the list of configurations
func (mp *ManagerProvider) Init() error {
	for _, conf := range mp.conf {
		switch conf.ProviderType {
		case config.ProviderTypeGit:
			cp, err := NewGitProvider(conf.Name, conf.Url, conf.SSHPath, mp.clonePath)
			if err != nil {
				return err
			}
			mp.providers[conf.Name] = cp
		default:
			return nerrors.NewInvalidArgumentError("provider not supported [%s]", conf.ProviderType)
		}
	}
	return mp.loadComponents()
}

// GetComponents returns  the catalog
func (mp *ManagerProvider) GetCatalog() map[string]grpc_catalog_manager_go.CatalogEntryResponse {

	return mp.catalog
}

// LoadComponents get the components from all the providers and stored them into catalog
func (mp *ManagerProvider) loadComponents() error{
	// uses a catalogAux to avoid lock main time
	catalog  := map[string]grpc_catalog_manager_go.CatalogEntryResponse{}

	// get Components
	for _, provider := range mp.providers {
		componentsList, err := provider.GetComponents()
		if err != nil {
			return err
		}
		// The catalogId is the name of the catalog
		catalogId := provider.GetName()
		for _, component := range componentsList {
			catalog[component.EntryId] = *utils.ComponentToCatalogEntryResponse(catalogId, component.EntryId, component.Component)
		}
	}

	mp.Lock()
	defer mp.Unlock()

	mp.catalog = catalog

	log.Debug().Int("catalog", len(catalog)).Msg("Catalog")

	return nil
}

// EmptyRepositories removes the cloned repositories
func (mp *ManagerProvider) EmptyRepositories() error{
	for _, provider := range mp.providers {
		if err := provider.EmptyCache(); err != nil {
			return err
		}
	}
	return nil
}

// LaunchAutomaticRepoUpdates reload the catalog each X minutes
func (mp *ManagerProvider) LaunchAutomaticRepoUpdates(d time.Duration) {
	for now := range time.Tick(d) {
		log.Debug().Interface("time", now).Msg("LaunchAutomaticRepoUpdates")
		if err := mp.loadComponents(); err != nil {
			log.Error().Str("err", nerrors.FromError(err).StackTraceToString()).Msg("error in LaunchAutomaticRepoUpdates ")
		}
	}
}

func (mp *ManagerProvider) GetComponent(catalogId string, componentId string) (*grpc_catalog_manager_go.CatalogEntryResponse, error) {
	component, exists := mp.catalog[componentId]
	if !exists {
		return nil, nerrors.NewNotFoundError("component does not exit [%s]", componentId)
	}
	return &component, nil
}