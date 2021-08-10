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
package cli

import (
	"github.com/napptive/catalog-manager/internal/pkg/provider/metadata"
	catalog_manager "github.com/napptive/catalog-manager/internal/pkg/server/catalog-manager"
	"github.com/napptive/catalog-manager/internal/pkg/storage"
	"github.com/rs/zerolog/log"
)

type ApplicationCli struct {
	// Index with the name of the elastic index
	Index string
	// ElasticAddress with the address to connect to Elastic
	ElasticAddress string
	// RepositoryPath with the path of the repository
	RepositoryPath string
	//CatalogUrl with the url of the repository (napptive repository must be nil)
	CatalogUrl string
}

func NewApplicationCli(index string, elasticAddress string, repoPat string, catalogUrl string) *ApplicationCli {
	return &ApplicationCli{
		Index:          index,
		ElasticAddress: elasticAddress,
		RepositoryPath: repoPat,
		CatalogUrl:     catalogUrl,
	}
}

func (ac *ApplicationCli) DeleteApplication(appName string) error {
	log.Debug().Str("appName", appName).Msg("DeleteApplication")

	pr, err := metadata.NewElasticProvider(ac.Index, ac.ElasticAddress)
	if err != nil {
		return err
	}
	err = pr.Init()
	if err != nil {
		return err
	}
	st := storage.NewStorageManager(ac.RepositoryPath)

	manager := catalog_manager.NewManager(st, pr, ac.CatalogUrl)

	return manager.Remove(appName)

}

func (ac *ApplicationCli) List (namespace string) error {
	pr, err := metadata.NewElasticProvider(ac.Index, ac.ElasticAddress)
	if err != nil {
		return err
	}
	err = pr.Init()
	if err != nil {
		return err
	}
	st := storage.NewStorageManager(ac.RepositoryPath)

	manager := catalog_manager.NewManager(st, pr, ac.CatalogUrl)
	result, err := manager.List(namespace)
	PrintResultOrError(result, err)
	return nil
}
