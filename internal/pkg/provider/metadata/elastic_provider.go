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
package metadata

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/nerrors/pkg/nerrors"

	"github.com/rs/zerolog/log"
)

const (
	// NamespaceField with the name of the field where we store the name of the repository
	NamespaceField = "Namespace"
	// ApplicationField with the name of the field where we store the name of the application
	ApplicationField = "ApplicationName"
	// TagField with the name of the field where we store the name of tag/version
	TagField = "Tag"
	// CatalogIDField with the name of the field where we store the internal ID
	CatalogIDField = "CatalogID"
	// CacheRefreshTime ick duration to update cache
	CacheRefreshTime = time.Second * 30
)

// mapping with the elastic-schema
var mapping = `{
    "mappings": {
        "properties": {
          "CatalogID":  		{ "type": "keyword" },
          "Namespace":  		{ "type": "keyword" },
          "ApplicationName":	{ "type": "keyword" },
          "Tag":         		{ "type": "keyword" },
          "Readme": 			{ "type": "text" },
          "Metadata":  			{ "type": "text" },
          "MetadataName":		{ "type": "text" }
      }
    }
}`

// responseWrapper is an struct used to load a search result
type responseWrapper struct {
	Took int
	Hits struct {
		Total struct {
			Value int
		}
		Hits []struct {
			ID         string          `json:"_id"`
			Source     json.RawMessage `json:"_source"`
			Highlights json.RawMessage `json:"highlight"`
			Sort       []interface{}   `json:"sort"`
		}
	}
}

type ElasticProvider struct {
	client    *elasticsearch.Client
	indexName string
	// appCache with a cache that contains all the catalog applications
	appCache []*entities.AppSummary
	// summaryCache with a cache that contains the catalog summary
	summaryCache *entities.Summary
	// Mutex to protect cache access
	sync.Mutex
	// invalidateCacheChan with a chan te send/receive message to fill Cache after remove or add an application
	invalidateCacheChan chan bool
}

// NewElasticProvider returns new Elastic provider
func NewElasticProvider(index string, address string) (*ElasticProvider, error) {

	conf := elasticsearch.Config{
		Addresses: []string{address},
	}
	es, err := elasticsearch.NewClient(conf)
	if err != nil {
		log.Err(err).Msg("error creating elastic client")
		return nil, err
	}
	return &ElasticProvider{
		client:              es,
		indexName:           index,
		appCache:            make([]*entities.AppSummary, 0),
		invalidateCacheChan: make(chan bool),
	}, nil
}

// Init creates the index and the necessary index
func (e *ElasticProvider) Init() error {
	log.Info().Msg("Initializing elastic provider")
	err := e.CreateIndex(mapping)
	if err != nil {
		return err
	}

	e.FillCache()

	go e.periodicCacheRefresh()

	return nil
}

// periodicCacheRefresh refresh the application cache
func (e *ElasticProvider) periodicCacheRefresh() {
	ticker := time.NewTicker(CacheRefreshTime)

	// Method executed in one thread to fill the cache every "CacheRefreshTime" time
	// or when a message is received through the "invalidateCacheChan" channel
	for {
		select {
		case val := <-e.invalidateCacheChan:
			if val {
				e.FillCache()
			} else {
				ticker.Stop()
				close(e.invalidateCacheChan)
				return
			}
		case <-ticker.C:
			e.FillCache()
		}
	}
}

// Finish method to exist in an orderly way
func (e *ElasticProvider) Finish() {
	// send a message to finish th timer and the close the channel
	e.invalidateCacheChan <- false
}

// IndexExists check if an index exists
func (e *ElasticProvider) IndexExists() (bool, error) {

	exists, err := esapi.IndicesExistsRequest{
		Index: []string{e.indexName},
	}.Do(context.Background(), e.client)

	if err != nil {
		return false, nerrors.FromError(err)
	}
	defer exists.Body.Close()

	if exists.IsError() {
		switch exists.StatusCode {
		case 404:
			return false, nil
		default:
			return false, nerrors.NewInternalError("error checking index. %s", exists.Status())
		}

	}
	return true, nil
}

// CreateIndex creates an index with the mapping received
func (e *ElasticProvider) CreateIndex(mapping string) error {

	exists, err := e.IndexExists()
	if err != nil {
		return err
	}
	// if not exist -> create it
	if !exists {
		res, err := e.client.Indices.Create(e.indexName, e.client.Indices.Create.WithBody(strings.NewReader(mapping)))
		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.IsError() {
			log.Warn().Str("err", res.String()).Msg("error creating index")
			return nerrors.NewInternalError("error creating index")
		}
	}

	return nil
}

// DeleteIndex removes a elastic index
func (e *ElasticProvider) DeleteIndex() error {
	resp, err := e.client.Indices.Delete([]string{e.indexName})
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}

// GenerateCatalogID generates the catalog ID (field stored in elastic) as namespace/appName:tag
func (e *ElasticProvider) GenerateCatalogID(namespace, appName, tag string) string {
	return fmt.Sprintf("%s/%s:%s", namespace, appName, tag)
}

// GenerateID generates the document _id
func (e *ElasticProvider) GenerateID(info *entities.ApplicationInfo) string {
	catalogID := e.GenerateCatalogID(info.Namespace, info.ApplicationName, info.Tag)
	id := md5.Sum([]byte(catalogID))
	return fmt.Sprintf("%x", id)
}

// GenerateIDFromAppID generates the document _id
func (e *ElasticProvider) GenerateIDFromAppID(metadata *entities.ApplicationID) string {
	catalogID := e.GenerateCatalogID(metadata.Namespace, metadata.ApplicationName, metadata.Tag)
	id := md5.Sum([]byte(catalogID))
	return fmt.Sprintf("%x", id)
}

func (e *ElasticProvider) chekElasticError(res *esapi.Response, operation string) error {

	if res.IsError() {
		log.Warn().Str("err", res.Status()).Str("operation", operation).Msg("Elastic error")
		return nerrors.NewInternalError("Error %s document: [%s]", operation, res.Status())
	}
	return nil
}

// Add stores new application metadata or updates it if it exists
func (e *ElasticProvider) Add(metadata *entities.ApplicationInfo) (*entities.ApplicationInfo, error) {

	// 1.- Generate _id
	id := e.GenerateID(metadata)

	// Fill Internal ID
	metadata.CatalogID = e.GenerateCatalogID(metadata.Namespace, metadata.ApplicationName, metadata.Tag)

	// convert the metadata to JSON
	metadataJSON, err := utils.ApplicationInfoToJSON(*metadata)
	if err != nil {
		log.Error().Err(err).Msg("error converting metadata to JSON")
		return nil, err
	}

	res, err := e.client.Index(e.indexName, strings.NewReader(metadataJSON),
		e.client.Index.WithRefresh("true"),
		e.client.Index.WithContext(context.Background()),
		e.client.Index.WithDocumentID(id))

	// Perform the request with the client.
	if err != nil {
		log.Error().Err(err).Msg("error adding metadata")
		return nil, err
	}
	defer res.Body.Close()

	if err = e.chekElasticError(res, "adding"); err != nil {
		return nil, err
	}

	// update the cache
	e.invalidateCacheChan <- true

	return metadata, nil
}

// Exists checks if the application Metadata already exists
func (e *ElasticProvider) Exists(appID *entities.ApplicationID) (bool, error) {

	id := e.GenerateIDFromAppID(appID)
	res, err := e.client.Exists(e.indexName, id, e.client.Exists.WithContext(context.Background()))
	if err != nil {
		log.Err(err).Msg("Error getting response")
		return false, nerrors.FromError(err)
	}
	defer res.Body.Close()

	return !res.IsError(), nil
}

// Get returns the application metadata requested
func (e *ElasticProvider) Get(appID *entities.ApplicationID) (*entities.ApplicationInfo, error) {

	id := e.GenerateIDFromAppID(appID)
	res, err := e.client.Get(e.indexName, id, e.client.Get.WithContext(context.Background()))
	if err != nil {
		log.Err(err).Msg("Error getting response")
		return nil, nerrors.NewInternalErrorFrom(err, "Error getting application")
	}

	defer res.Body.Close()

	if err = e.chekElasticError(res, "getting"); err != nil {
		return nil, err
	}

	var p map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, nerrors.FromError(err)
	}
	data, exists := p["_source"]
	if !exists {
		return nil, nerrors.NewInternalError("Error getting application, no _source found")
	}

	var application entities.ApplicationInfo
	other, err := json.Marshal(data)
	if err != nil {
		return nil, nerrors.FromError(err)
	}
	err = json.Unmarshal(other, &application)
	if err != nil {
		return nil, nerrors.FromError(err)
	}
	return &application, nil

}

// Remove deletes an application from the catalog
func (e *ElasticProvider) Remove(appID *entities.ApplicationID) error {
	id := e.GenerateIDFromAppID(appID)
	res, err := e.client.Delete(e.indexName, id, e.client.Delete.WithContext(context.Background()), e.client.Delete.WithRefresh("true"))

	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Error deleting metadata")
		return nerrors.NewInternalErrorFrom(err, "error deleting metadata by ID")
	}
	defer res.Body.Close()

	if err = e.chekElasticError(res, "removing"); err != nil {
		return err
	}

	e.invalidateCacheChan <- true

	return nil
}

// List returns all the applications stored
func (e *ElasticProvider) List(namespace string) ([]*entities.ApplicationInfo, error) {

	lastReceived := 0
	query := true
	applications := make([]*entities.ApplicationInfo, 0)

	for query {
		r, err := e.listFrom(namespace, lastReceived)
		if err != nil {
			return nil, err
		}

		log.Debug().Int("hits received", len(r.Hits.Hits)).Msg("received")
		for _, app := range r.Hits.Hits {
			var application entities.ApplicationInfo
			if err := json.Unmarshal(app.Source, &application); err != nil {
				return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
			}
			applications = append(applications, &application)
		}
		lastReceived += len(r.Hits.Hits)
		query = r.Hits.Total.Value != len(applications) && len(r.Hits.Hits) != 0
	}

	return applications, nil

}

// listFrom returns applications from last received
func (e *ElasticProvider) listFrom(namespace string, lastReceived int, getFields ...string) (*responseWrapper, error) {

	sortedBy := []string{"Namespace", "ApplicationName", "Tag"}
	searchFunctions := []func(*esapi.SearchRequest){
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.indexName),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithFrom(lastReceived),
		e.client.Search.WithSort(sortedBy...),
	}

	if len(getFields) > 0 {
		searchFunctions = append(searchFunctions, e.client.Search.WithSourceIncludes(getFields...))
	}

	if namespace != "" {
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					NamespaceField: namespace,
				},
			},
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Err(err).Msg("Error encoding namespaced query")
			return nil, nerrors.NewInternalErrorFrom(err, "error creating query to list by namespace")
		}
		searchFunctions = append(searchFunctions, e.client.Search.WithBody(&buf))

	}

	// Perform the search request.
	res, err := e.client.Search(searchFunctions...)
	if err != nil {
		log.Err(err).Msg("Error getting response")
		return nil, nerrors.FromError(err)
	}
	defer res.Body.Close()

	if err = e.chekElasticError(res, "listing"); err != nil {
		return nil, err
	}

	var r responseWrapper
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, nerrors.FromError(err)
	}
	// Print the response status, number of results, and request duration.
	log.Debug().Str("Status", res.Status()).Int("total", r.Hits.Total.Value).Int("took(ms)", r.Took).Msg("List operation")

	return &r, nil
}

func (e *ElasticProvider) getSummaryList(namespace string) ([]*entities.AppSummary, *entities.Summary, error) {

	lastReceived := 0
	query := true
	summaryList := make([]*entities.AppSummary, 0)
	var summary entities.Summary
	total := 0
	getFields := []string{"Namespace", "ApplicationName", "Tag", "MetadataName", "Metadata"}
	for query {
		r, err := e.listFrom(namespace, lastReceived, getFields...)
		if err != nil {
			return nil, nil, err
		}

		for _, app := range r.Hits.Hits {
			var application entities.ExtendedAppSummary
			if err := json.Unmarshal(app.Source, &application); err != nil {
				return nil, nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
			}

			// new version
			summary.NumTags++

			var metadataLogo []entities.ApplicationLogo
			_, metadata, err := utils.IsMetadata([]byte(application.Metadata))
			if err != nil {
				// If returns the error, the catalog could be inaccessible. It could be better not return an error and allows to continue listing
				//return nil, nil, err
				log.Warn().Str("error", err.Error()).Msg("error getting metadata")
			} else {
				metadataLogo = metadata.Logo
			}

			// check if the last entry has the same namespace and applicationName as the newer one
			if len(summaryList) > 0 {
				last := summaryList[len(summaryList)-1]
				if last.Namespace != application.Namespace {
					// new namespace
					summary.NumNamespaces++
				}
				if last.Namespace == application.Namespace && last.ApplicationName == application.ApplicationName {
					summaryList[len(summaryList)-1].TagMetadataName[application.Tag] = application.MetadataName
					if metadataLogo != nil {
						summaryList[len(summaryList)-1].MetadataLogo[application.Tag] = metadataLogo
					}
				} else {
					// new application
					summary.NumApplications++
					newSumm := &entities.AppSummary{
						Namespace:       application.Namespace,
						ApplicationName: application.ApplicationName,
						TagMetadataName: map[string]string{application.Tag: application.MetadataName},
						MetadataLogo:    map[string][]entities.ApplicationLogo{},
					}
					if metadataLogo != nil {
						newSumm.MetadataLogo[application.Tag] = metadataLogo
					}
					summaryList = append(summaryList, newSumm)
				}
			} else {
				// new namespace
				summary.NumNamespaces++
				// new application (new tag updated above)
				summary.NumApplications++
				newSumm := &entities.AppSummary{
					Namespace:       application.Namespace,
					ApplicationName: application.ApplicationName,
					TagMetadataName: map[string]string{application.Tag: application.MetadataName},
					MetadataLogo:    map[string][]entities.ApplicationLogo{},
				}
				if metadataLogo != nil {
					newSumm.MetadataLogo[application.Tag] = metadataLogo
				}
				summaryList = append(summaryList, newSumm)

			}
			total++
		}
		lastReceived += len(r.Hits.Hits)
		query = r.Hits.Total.Value != total && len(r.Hits.Hits) != 0
	}

	return summaryList, &summary, nil
}

// ListSummary returns all the catalog applications.
// if it doesn't ask for namespace -> returns cache
// if namespace is filled -> elastic query
func (e *ElasticProvider) ListSummary(namespace string) ([]*entities.AppSummary, error) {

	if namespace == "" {
		e.Lock()
		defer e.Unlock()
		return e.appCache, nil
	}
	summaryList, _, err := e.getSummaryList(namespace)

	return summaryList, err
}

// FillCache refresh the cache with the applications
func (e *ElasticProvider) FillCache() {
	// ListSummary and fillCache
	summaryList, summary, err := e.getSummaryList("")
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("error filling the cache")
	} else {
		e.Lock()
		defer e.Unlock()

		e.appCache = summaryList
		e.summaryCache = summary
	}
}

// GetSummary returns the catalog summary
func (e *ElasticProvider) GetSummary() (*entities.Summary, error) {
	e.Lock()
	defer e.Unlock()
	if e.summaryCache == nil {
		return nil, nerrors.NewInternalError("error getting catalog summary")
	}
	return e.summaryCache, nil
}
