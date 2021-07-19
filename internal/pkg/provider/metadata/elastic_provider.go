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

// deleteResponseWrapper is an struct used to load a delete result
type deleteResponseWrapper struct {
	Batches int
	Deleted int
	Noops   int
	Retries struct {
		Bulk   int
		Search int
	}
	ThrottledMllis       int  `json:"throttled_millis"`
	ThrottledUntilMillis int  `json:"throttled_until_millis"`
	TimedOut             bool `json:"timed_out"`
	Took                 int
	Total                int
	VersionConflicts     int `json:"version_conflicts"`
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
	lastTime := time.Now().Add(-1 * CacheRefreshTime)
	for {
		select {
		case val := <-e.invalidateCacheChan:
			if val {
				// When less than x seconds have passed since the last update, we do not update
				if lastTime.Add(CacheRefreshTime / 3).Before(time.Now()) {
					e.FillCache()
					lastTime = time.Now()
				}

			} else {
				ticker.Stop()
				close(e.invalidateCacheChan)
				return
			}
		case <-ticker.C:
			// When less than x seconds have passed since the last update, we do not update
			if lastTime.Add(CacheRefreshTime / 3).Before(time.Now()) {
				e.FillCache()
				lastTime = time.Now()
			}
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

// CreateID generates the documentID to store the application metadata
func (e *ElasticProvider) CreateID(metadata entities.ApplicationInfo) string {
	id := fmt.Sprintf("%s/%s:%s", metadata.Namespace, metadata.ApplicationName, metadata.Tag)
	return id
}

// CreateIDFromAppID creates the documentID from ApplicationID
func (e *ElasticProvider) CreateIDFromAppID(metadata entities.ApplicationID) string {
	id := fmt.Sprintf("%s/%s:%s", metadata.Namespace, metadata.ApplicationName, metadata.Tag)
	return id
}

// buildTerm creates a term field to search
func (e *ElasticProvider) buildTerm(field string, value string) map[string]interface{} {
	return map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	}
}

// buildQuery creates the query to ask for an application metadata
// this method is used to ask about ALL the files (url, repo, appName and tag)
func (e *ElasticProvider) buildQuery(appID entities.ApplicationID) map[string]interface{} {

	repo := e.buildTerm(NamespaceField, appID.Namespace)
	appName := e.buildTerm(ApplicationField, appID.ApplicationName)
	tag := e.buildTerm(TagField, appID.Tag)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{repo, appName, tag},
			},
		},
	}

	return query
}

// Add stores new application metadata or updates it if it exists
func (e *ElasticProvider) Add(metadata *entities.ApplicationInfo) (*entities.ApplicationInfo, error) {

	// Check if application already exists -> remove it!
	appID := metadata.ToApplicationID()
	exists, err := e.Exists(appID)
	if err != nil {
		log.Err(err).Msg("error checking if application exists")
		return nil, nerrors.NewInternalError("Unable to add application Metadata, unable to check if application already exists")
	}

	if exists {
		if err = e.Remove(appID); err != nil {
			return nil, nerrors.NewInternalError("Unable to add application Metadata, unable to remove previous application")
		}
	}

	// Fill Internal ID
	metadata.CatalogID = e.CreateID(*metadata)
	//	metadata.MetadataName = metadata.MetadataObj.Name

	// convert the metadata to JSON
	metadataJSON, err := utils.ApplicationInfoToJSON(*metadata)
	if err != nil {
		log.Error().Err(err).Msg("error converting metadata to JSON")
		return nil, err
	}

	req := esapi.IndexRequest{
		Index:      e.indexName,
		Body:       strings.NewReader(metadataJSON),
		Timeout:    0,
		Pretty:     false,
		Human:      false,
		ErrorTrace: false,
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		log.Error().Err(err).Msg("error adding metadata")
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Warn().Str("err", res.Status()).Msg("Error indexing document")
		return nil, nerrors.NewInternalError("Error indexing document: [%s]", res.Status())

	}

	// update the cache
	e.invalidateCacheChan <- true

	return metadata, nil
}

// SearchByApplicationID returns the application metadata requested
// Right now it is not used neither tested
func (e *ElasticProvider) SearchByApplicationID(appID entities.ApplicationID) (*entities.ApplicationInfo, error) {

	var buf bytes.Buffer

	query := e.buildQuery(appID)
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Err(err).Msg("Error encoding query")
	}

	// Perform the search request.
	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.indexName),
		e.client.Search.WithBody(&buf),
		e.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		log.Err(err).Msg("Error getting response")
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Err(err).Msg("Error parsing the response body")
			return nil, nerrors.NewInternalError("Error getting application. Error parsing the response body")
		} else {
			// Print the response status and error information.
			log.Err(err).Str("status", res.Status()).Msg("error")
			return nil, nerrors.NewInternalError(res.Status())
		}
	}

	var r responseWrapper
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, nerrors.FromError(err)
	}

	// Print the response status, number of results, and request duration.
	log.Debug().Str("Status", res.Status()).Int("total", r.Hits.Total.Value).Int("took(ms)", r.Took).Msg("SearchByApplicationID operation")

	if r.Hits.Total.Value == 0 {
		return nil, nerrors.NewNotFoundError("%s metadata is not available", appID.String())
	}

	var application entities.ApplicationInfo
	if err := json.Unmarshal(r.Hits.Hits[0].Source, &application); err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
	}

	return &application, nil
}

// Exists checks if the application Metadata already exists
func (e *ElasticProvider) Exists(appID *entities.ApplicationID) (bool, error) {

	catalogID := e.CreateIDFromAppID(*appID)

	var buf bytes.Buffer

	// Query Field
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				CatalogIDField: catalogID,
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Err(err).Msg("Error encoding query")
		return false, nerrors.NewInternalErrorFrom(err, "error getting metadata by ID")
	}

	// Perform the search request.
	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.indexName),
		e.client.Search.WithBody(&buf),
		e.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		log.Err(err).Msg("Error getting response")
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Err(err).Msg("Error parsing the response body")
			return false, nerrors.NewInternalError("Error getting application. Error parsing the response body")
		} else {
			// Print the response status and error information.
			log.Err(err).Str("status", res.Status()).Msg("error")
			return false, nerrors.NewInternalError(res.Status())
		}
	}

	var r responseWrapper
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return false, nerrors.FromError(err)
	}

	return r.Hits.Total.Value != 0, nil

}

// Get returns the application metadata requested
func (e *ElasticProvider) Get(appID entities.ApplicationID) (*entities.ApplicationInfo, error) {
	catalogID := e.CreateIDFromAppID(appID)
	var buf bytes.Buffer

	// Query Field
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				CatalogIDField: catalogID,
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Err(err).Msg("Error encoding query")
		return nil, nerrors.NewInternalErrorFrom(err, "error getting metadata by ID")
	}

	// Perform the search request.
	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.indexName),
		e.client.Search.WithBody(&buf),
		e.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		log.Err(err).Msg("Error getting response")
		return nil, nerrors.NewInternalErrorFrom(err, "Error getting application")
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Err(err).Msg("Error parsing the response body")
			return nil, nerrors.NewInternalError("Error getting application. Error parsing the response body")
		} else {
			// Print the response status and error information.
			log.Err(err).Str("status", res.Status()).Msg("error")
			return nil, nerrors.NewInternalError(res.Status())
		}
	}

	var r responseWrapper
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, nerrors.FromError(err)
	}

	// Print the response status, number of results, and request duration.
	log.Debug().Str("Status", res.Status()).Int("total", r.Hits.Total.Value).Int("took(ms)", r.Took).Msg("Get operation")

	if r.Hits.Total.Value == 0 {
		return nil, nerrors.NewNotFoundError("%s metadata is not available", appID.String())
	}
	if r.Hits.Total.Value != 1 {
		log.Error().Str("catalogID", catalogID).Msg("Error getting application metadata, duplicated entries")
		return nil, nerrors.NewInternalError("Error getting application metadata, duplicated entries")
	}

	var application entities.ApplicationInfo
	if err := json.Unmarshal(r.Hits.Hits[0].Source, &application); err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
	}

	return &application, nil

}

// Remove deletes an application from the catalog
func (e *ElasticProvider) Remove(appID *entities.ApplicationID) error {

	catalogID := e.CreateIDFromAppID(*appID)
	var buf bytes.Buffer

	// Query Field
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				CatalogIDField: catalogID,
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Err(err).Msg("Error deleting metadata. Error encoding query")
		return nerrors.NewInternalErrorFrom(err, "error deleting metadata by ID")
	}

	req := esapi.DeleteByQueryRequest{
		Index: []string{e.indexName},
		Body:  strings.NewReader(buf.String()),
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		log.Err(err).Msg("Error deleting metadata")
		return nerrors.NewInternalErrorFrom(err, "error deleting metadata by ID")
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Err(err).Str("status", res.Status()).Msg("Error deleting metadata")
		return nerrors.NewInternalError("error deleting metadata by ID [%s]", res.Status())
	}

	var d deleteResponseWrapper
	if err := json.NewDecoder(res.Body).Decode(&d); err != nil {
		return nerrors.FromError(err)
	}

	if d.Deleted == 0 {
		return nerrors.NewNotFoundError("%s metadata is not available", appID.String())
	}

	// Print the response status, number of results, and request duration.
	log.Debug().Str("Status", res.Status()).Int("total", d.Total).Int("took(ms)", d.Took).Msg("Delete operation")

	e.invalidateCacheChan <- true

	return nil
}

// List returns all the applications stored
func (e *ElasticProvider) List(namespace string) ([]*entities.ApplicationInfo, error) {

	lastReceived := 0
	query := true
	applications := make([]*entities.ApplicationInfo, 0)

	for query {
		r, err := e.list(namespace, lastReceived)
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

// list lists the applications from one retrieved
func (e *ElasticProvider) list(namespace string, lastReceived int) (*responseWrapper, error) {

	sortedBy := []string{"Namespace", "ApplicationName", "Tag"}
	searchFunctions := []func(*esapi.SearchRequest){
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.indexName),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithFrom(lastReceived),
		e.client.Search.WithSort(sortedBy...),
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
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Err(err).Msg("Error parsing the response body")
			return nil, nerrors.NewInternalError("Error getting application. Error parsing the response body")
		} else {
			// Print the response status and error information.
			log.Err(err).Str("status", res.Status()).Msg("error")
			return nil, nerrors.NewInternalError(res.Status())
		}
	}

	var r responseWrapper
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, nerrors.FromError(err)
	}
	// Print the response status, number of results, and request duration.
	log.Debug().Str("Status", res.Status()).Int("total", r.Hits.Total.Value).Int("took(ms)", r.Took).Msg("List operation")

	return &r, nil
}

// listFrom returns applications from last received
func (e *ElasticProvider) listFrom(namespace string, lastReceived int) (*responseWrapper, error) {

	sortedBy := []string{"Namespace", "ApplicationName", "Tag"}
	getFields := []string{"Namespace", "ApplicationName", "Tag", "MetadataName", "Metadata"}
	searchFunctions := []func(*esapi.SearchRequest){
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.indexName),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithFrom(lastReceived),
		e.client.Search.WithSort(sortedBy...),
		e.client.Search.WithSourceIncludes(getFields...),
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
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Err(err).Msg("Error parsing the response body")
			return nil, nerrors.NewInternalError("Error getting application. Error parsing the response body")
		} else {
			// Print the response status and error information.
			log.Err(err).Str("status", res.Status()).Msg("error")
			return nil, nerrors.NewInternalError(res.Status())
		}
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
	for query {
		r, err := e.listFrom(namespace, lastReceived)
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
						MetadataLogo: map[string][]entities.ApplicationLogo{},

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
					MetadataLogo: map[string][]entities.ApplicationLogo{},
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
