/**
 * Copyright 2023 Napptive
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
	analytics "github.com/napptive/analytics/pkg/provider"
	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/provider/metadata"
	"github.com/napptive/catalog-manager/internal/pkg/storage"
)

// Providers with all the providers needed
type Providers struct {
	// elasticProvider with a elastic provider to store metadata
	elasticProvider metadata.MetadataProvider
	// repoStorage to store the applications
	repoStorage storage.StorageManager
	// analyticsProvider to store operation metrics
	analyticsProvider analytics.Provider
}

// GetProviders creates and initializes all the providers
func GetProviders(cfg *config.Config) (*Providers, error) {
	pr, err := metadata.NewElasticProvider(cfg.Index, cfg.ElasticAddress, cfg.AuthEnabled)
	if err != nil {
		return nil, err
	}
	err = pr.Init()
	if err != nil {
		return nil, err
	}

	if cfg.BQConfig.Enabled {
		provider, err := analytics.NewBigQueryProvider(cfg.BQConfig.Config)
		if err != nil {
			return nil, err
		}
		return &Providers{
			elasticProvider:   pr,
			repoStorage:       storage.NewStorageManager(cfg.RepositoryPath),
			analyticsProvider: provider}, nil
	}
	// ! s.cfg.BQConfig.Enabled
	return &Providers{elasticProvider: pr,
		repoStorage: storage.NewStorageManager(cfg.RepositoryPath)}, nil

}
