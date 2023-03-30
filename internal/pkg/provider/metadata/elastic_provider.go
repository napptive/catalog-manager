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

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/elastic/go-elasticsearch/v7/esapi"

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
	// MetadataNameField with the name of the field where we store the name in the metadata
	MetadataNameField = "MetadataName"
	// MetadataField with the name of the field where we store the application metadata
	MetadataField = "Metadata"
	// CatalogIDField with the name of the field where we store the internal ID
	CatalogIDField = "CatalogID"
	// PrivateField with the name of the field where we store the application scope
	PrivateField = "Private"
	// CacheRefreshTime ick duration to update cache
	// CacheRefreshTime = time.Second * 30
	CacheRefreshTime = time.Minute * 5
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
          "MetadataName":		{ "type": "text" },
          "Private": 			{ "type": "boolean" }
      }
    }
}`

// responseWrapper is a struct used to load a search result
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

type ElasticFilter interface {
	ToElasticQuery() map[string]interface{}
	ToString() string
}

// ToElasticQuery returns the search query for a ListFilter. Required to implement ElasticFilter interface
/*
curl -X GET "localhost:9200/napptive/_search?pretty" -H 'Content-Type: application/json' -d'

	{
	  "query": {
	    "bool" : {
	      "must" : {
	        "term" : { "Private" : <private> }
	      },
	      "filter": {
	        "term" : { "Namespace" : <namespace> }
	      }
	    }
	  }
	}

'
*/
func (f *ListFilter) ToElasticQuery() map[string]interface{} {
	var query map[string]interface{}

	if f == nil {
		return query
	}

	// Namespace && Private
	if f.Namespace != nil && *f.Namespace != "" && f.Private != nil {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": map[string]interface{}{
						"term": map[string]interface{}{PrivateField: *f.Private},
					},
					"filter": map[string]interface{}{
						"term": map[string]interface{}{NamespaceField: *f.Namespace},
					},
				},
			},
		} // Private
	} else if f.Private != nil && (f.Namespace == nil || *f.Namespace == "") {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"term": map[string]interface{}{
					PrivateField: *f.Private,
				},
			},
		}
		// Namespace
	} else if f.Namespace != nil && *f.Namespace != "" && f.Private == nil {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"term": map[string]interface{}{
					NamespaceField: *f.Namespace,
				},
			},
		}
	}
	return query
}

func (f *ListFilter) ToString() string {
	namespace := ""
	private := ""
	if f.Private != nil {
		private = fmt.Sprintf("%v", *f.Private)
	}
	if f.Namespace != nil {
		namespace = *f.Namespace
	}
	return fmt.Sprintf("ListFilter. Namespace [%s] - Private [%s]", namespace, private)
}

// ApplicationFilter struct to filter by namespace and applicationName
type ApplicationFilter struct {
	namespace   string
	application string
}

// ToElasticQuery returns the search query for a ApplicationFilter. Required to implement ElasticFilter interface
/*
curl -X GET "localhost:9200/napptive/_search?pretty" -H 'Content-Type: application/json' -d'

	{
	  "query": {
	    "bool" : {
	      "must" : {
	        "term" : { "Namespace" : <namespace> }
	      },
	      "filter": {
	        "term" : { "ApplicationName" : <applicationName> }
	      }
	    }
	  }
	}

'
*/
func (af *ApplicationFilter) ToElasticQuery() map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"term": map[string]interface{}{NamespaceField: af.namespace},
				},
				"filter": map[string]interface{}{
					"term": map[string]interface{}{ApplicationField: af.application},
				},
			},
		},
	}
}

func (af *ApplicationFilter) ToString() string {
	return fmt.Sprintf("ApplicationFilter. Namespace [%s] - Application [%s]", af.namespace, af.application)
}

// ElasticProvider a struct to manage elastic storage
type ElasticProvider struct {
	client    *elasticsearch.Client
	indexName string
	// appCache with a cache that contains all the catalog PUBLIC applications
	appCache []*entities.AppSummary
	// summaryCache with a cache that contains the catalog summary (with PUBLIC applications)
	summaryCache *entities.Summary
	// Mutex to protect cache access
	sync.Mutex
	// invalidateCacheChan with a chan te send/receive message to fill Cache after remove or add an application
	invalidateCacheChan chan bool
	// authEnable with a flag to indicate if the authorization is enabled
	authEnable bool
}

// NewElasticProvider returns new Elastic provider
func NewElasticProvider(index string, address string, authEnable bool) (*ElasticProvider, error) {

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
		authEnable:          authEnable,
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

func (e *ElasticProvider) checkElasticError(res *esapi.Response, operation string) error {

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

	if err = e.checkElasticError(res, "adding"); err != nil {
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

	if err = e.checkElasticError(res, "getting"); err != nil {
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
	log.Debug().Str("id", id).Msg("Remove app id")
	res, err := e.client.Delete(e.indexName, id, e.client.Delete.WithContext(context.Background()), e.client.Delete.WithRefresh("true"))

	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Error deleting metadata")
		return nerrors.NewInternalErrorFrom(err, "error deleting metadata by ID")
	}
	defer res.Body.Close()

	if err = e.checkElasticError(res, "removing"); err != nil {
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

	filter := &ListFilter{
		Namespace: &namespace,
		Private:   nil,
	}

	for query {
		r, err := e.listFromWithFilter(filter, lastReceived)
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

// FillCache refresh the cache with the applications
func (e *ElasticProvider) FillCache() {
	// ListSummary and fillCache
	var listFilter *ListFilter
	if e.authEnable {
		private := false
		listFilter = &ListFilter{
			Namespace: nil,
			Private:   &private,
		}
	} else {
		listFilter = &ListFilter{
			Namespace: nil,
			Private:   nil,
		}
	}
	summaryList, summary, err := e.listSummaryWithFilter(listFilter)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("error filling the cache")
	} else {
		e.Lock()
		defer e.Unlock()
		e.appCache = summaryList
		e.summaryCache = summary
		log.Debug().Int("len", len(e.appCache)).Msg("applications in cache")

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

// listFromWithFilter search applications in elastic with pagination
func (e *ElasticProvider) listFromWithFilter(filter ElasticFilter, lastReceived int, getFields ...string) (*responseWrapper, error) {

	log.Debug().Str("filter", filter.ToString()).Int("last received", lastReceived).Msg("listFromWithFilter")
	sortedBy := []string{NamespaceField, ApplicationField, TagField}
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

	query := filter.ToElasticQuery()
	if len(query) > 0 {

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

	if err = e.checkElasticError(res, "listing"); err != nil {
		return nil, err
	}

	var r responseWrapper
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, nerrors.FromError(err)
	}
	// Print the response status, number of results, and request duration.
	log.Debug().Str("filter", filter.ToString()).Str("Status", res.Status()).Int("total", r.Hits.Total.Value).Int("took(ms)", r.Took).Msg("List operation")

	return &r, nil
}

// ListSummaryWithFilter returns entities.AppSummary and entities.Summary applying a filter in the search method
func (e *ElasticProvider) ListSummaryWithFilter(filter *ListFilter) ([]*entities.AppSummary, *entities.Summary, error) {
	// if filtering == (public applications for all namespaces) -> return cache
	if filter != nil && (filter.Namespace == nil || *filter.Namespace == "") && (filter.Private == nil || !*filter.Private) {
		e.Lock()
		defer e.Unlock()
		return e.appCache, e.summaryCache, nil
	} else {
		return e.listSummaryWithFilter(filter)
	}
}

// ListSummaryWithFilter returns entities.AppSummary and entities.Summary applying a filter in the search method
func (e *ElasticProvider) listSummaryWithFilter(filter *ListFilter) ([]*entities.AppSummary, *entities.Summary, error) {
	lastReceived := 0
	query := true
	summaryList := make([]*entities.AppSummary, 0)
	var summary entities.Summary
	total := 0
	getFields := []string{NamespaceField, ApplicationField, TagField, MetadataNameField, MetadataField, PrivateField}

	for query {
		r, err := e.listFromWithFilter(filter, lastReceived, getFields...)
		if err != nil {
			return nil, nil, err
		}

		for _, app := range r.Hits.Hits {
			// var application entities.ExtendedAppSummary
			var application entities.ApplicationInfo
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
					newAppSummary := &entities.AppSummary{
						Namespace:       application.Namespace,
						ApplicationName: application.ApplicationName,
						TagMetadataName: map[string]string{application.Tag: application.MetadataName},
						MetadataLogo:    map[string][]entities.ApplicationLogo{},
						Private:         application.Private,
					}
					if metadataLogo != nil {
						newAppSummary.MetadataLogo[application.Tag] = metadataLogo
					}
					summaryList = append(summaryList, newAppSummary)
				}
			} else {
				// new namespace
				summary.NumNamespaces++
				// new application (new tag updated above)
				summary.NumApplications++
				newAppSummary := &entities.AppSummary{
					Namespace:       application.Namespace,
					ApplicationName: application.ApplicationName,
					TagMetadataName: map[string]string{application.Tag: application.MetadataName},
					MetadataLogo:    map[string][]entities.ApplicationLogo{},
					Private:         application.Private,
				}
				if metadataLogo != nil {
					newAppSummary.MetadataLogo[application.Tag] = metadataLogo
				}
				summaryList = append(summaryList, newAppSummary)

			}
			total++
		}
		lastReceived += len(r.Hits.Hits)
		query = r.Hits.Total.Value != total && len(r.Hits.Hits) != 0
	}

	return summaryList, &summary, nil
}

// GetApplicationVisibility returns the application visibility or error if the application does not exist
func (e *ElasticProvider) GetApplicationVisibility(namespace string, applicationName string) (*bool, error) {

	sortedBy := []string{NamespaceField, ApplicationField, TagField}
	searchFunctions := []func(*esapi.SearchRequest){
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.indexName),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithSize(1), // only one requested -> all the applications have the same visibility
		e.client.Search.WithSort(sortedBy...),
		e.client.Search.WithSourceIncludes(NamespaceField, ApplicationField, TagField, PrivateField),
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"term": map[string]interface{}{NamespaceField: namespace},
				},
				"filter": map[string]interface{}{
					"term": map[string]interface{}{ApplicationField: applicationName},
				},
			},
		},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Err(err).Msg("Error encoding namespaced query")
		return nil, nerrors.NewInternalErrorFrom(err, "error all applications tags")
	}
	searchFunctions = append(searchFunctions, e.client.Search.WithBody(&buf))

	// Perform the search request.
	res, err := e.client.Search(searchFunctions...)
	if err != nil {
		log.Err(err).Msg("Error getting response")
		return nil, nerrors.FromError(err)
	}
	defer res.Body.Close()

	if err = e.checkElasticError(res, "getting tag"); err != nil {
		return nil, err
	}

	var r responseWrapper
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, nerrors.FromError(err)
	}

	if len(r.Hits.Hits) <= 0 {
		return nil, nerrors.NewNotFoundError("application not found")
	}

	var application entities.ApplicationInfo
	if err := json.Unmarshal(r.Hits.Hits[0].Source, &application); err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
	}
	return &application.Private, nil

}

// getApplicationIds returns the internal identifiers of all the tags of an application and the application visibility (if it is private or public)
func (e *ElasticProvider) getApplicationIds(namespace string, application string) ([]string, error) {

	lastReceived := 0
	query := true
	ids := make([]string, 0)
	for query {
		r, err := e.listFromWithFilter(&ApplicationFilter{
			namespace:   namespace,
			application: application,
		}, lastReceived)
		if err != nil {
			return nil, err
		}

		log.Debug().Int("hits received", len(r.Hits.Hits)).Msg("received")
		if len(r.Hits.Hits) > 0 {
			var application entities.ApplicationInfo
			if err := json.Unmarshal(r.Hits.Hits[0].Source, &application); err != nil {
				return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
			}
		}
		for _, app := range r.Hits.Hits {
			ids = append(ids, app.ID)
		}
		lastReceived += len(r.Hits.Hits)
		query = r.Hits.Total.Value != len(ids) && len(r.Hits.Hits) != 0
	}

	return ids, nil

}

// updateVisibilityStruct struct required to update application visibility
type updateVisibilityStruct struct {
	Private bool
}

// UpdateApplicationVisibility changes the application visibility
func (e *ElasticProvider) UpdateApplicationVisibility(namespace string, applicationName string, isPrivate bool) error {
	log.Debug().Msg("UpdateApplicationVisibility")

	// Get all the catalogIDs for namespace, applicationName
	// Foreach:
	// update the data

	ids, err := e.getApplicationIds(namespace, applicationName)
	if err != nil {
		log.Error().Err(err).Msg("error getting application tags")
		return err
	}

	if len(ids) == 0 {
		log.Error().Str("namespace", namespace).Str("application", applicationName).
			Msg("error changing application visibility, no applications found")
		return nerrors.NewNotFoundError("unable to update application visibility. Application not found")
	}

	data := &updateVisibilityStruct{
		Private: isPrivate,
	}

	// convert the metadata to JSON
	metadataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	for _, id := range ids {
		res, err := esapi.UpdateRequest{
			Refresh:    "true",
			Index:      "napptive",
			DocumentID: id,
			Body:       bytes.NewReader([]byte(fmt.Sprintf(`{"doc":%s}`, string(metadataJSON)))),
		}.Do(context.Background(), e.client)
		if err != nil {
			log.Error().Err(err).Msg("error updating metadata")
			return err
		}
		defer res.Body.Close()

		if err = e.checkElasticError(res, "updating visibility"); err != nil {
			return err
		}
	}

	// update the cache
	e.invalidateCacheChan <- true

	return nil
}
