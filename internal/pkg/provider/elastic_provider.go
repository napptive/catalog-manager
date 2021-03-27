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
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/nerrors/pkg/nerrors"

	"github.com/rs/zerolog/log"
)

const (
	// RepositoryField with the name of the field where we store the name of the repository
	RepositoryField = "Repository"
	// ApplicationField with the name of the field where we store the name of the application
	ApplicationField = "ApplicationName"
	// TagField with the name of the field where we store the name of tag/version
	TagField = "Tag"
	// CatalogIDField with the name of the field where we store the internal ID
	CatalogIDField = "CatalogID"
)

// mapping with the elastic-schema
var mapping = `{
    "mappings": {
        "properties": {
          "CatalogID":  		{ "type": "keyword" },
          "Repository":  		{ "type": "keyword" },
          "ApplicationName":	{ "type": "keyword" },
          "Tag":         		{ "type": "keyword" },
          "Readme": 			{ "type": "text" },
          "Metadata":  			{ "type": "text" }
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
		Batches           int
		Deleted           int
		Noops             int
		Retries           struct {
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
}

// NewElasticProvider returns new Elastic provider
func NewElasticProvider(index string, address string) (*ElasticProvider, error) {
	// TODO: Change DefaultClient to NewClient(cfg Config)
	conf := elasticsearch.Config{
		Addresses: []string{address},
	}
	es, err := elasticsearch.NewClient(conf)
	if err != nil {
		log.Err(err).Msg("error creating elastic client")
		return nil, err
	}
	return &ElasticProvider{
		client:    es,
		indexName: index,
	}, nil
}

// Init creates the index and the necessary index
func (e *ElasticProvider) Init() error {
	log.Info().Msg("Initializing elastic provider")
	return e.CreateIndex(mapping)
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
func (e *ElasticProvider) CreateID(metadata entities.ApplicationMetadata) string {
	id := fmt.Sprintf("%s/%s:%s", metadata.Repository, metadata.ApplicationName, metadata.Tag)
	return id
}

// CreateIDFromAppID creates the documentID from ApplicationID
func (e *ElasticProvider) CreateIDFromAppID(metadata entities.ApplicationID) string {
	id := fmt.Sprintf("%s/%s:%s", metadata.Repository, metadata.ApplicationName, metadata.Tag)
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

	repo := e.buildTerm(RepositoryField, appID.Repository)
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
func (e *ElasticProvider) Add(metadata *entities.ApplicationMetadata) (*entities.ApplicationMetadata, error) {

	// Fill Internal ID
	metadata.CatalogID = e.CreateID(*metadata)

	// convert the metadata to JSON
	metadataJSON, err := utils.ApplicationMetadataToJSON(*metadata)
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

	return metadata, nil
}

// SearchByApplicationID returns the application metadata requested
// Right now it is not used neither tested
func (e *ElasticProvider) SearchByApplicationID(appID entities.ApplicationID) (*entities.ApplicationMetadata, error) {

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
		return nil, nerrors.NewNotFoundError("application metadata not found")
	}

	var application entities.ApplicationMetadata
	if err := json.Unmarshal(r.Hits.Hits[0].Source, &application); err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
	}

	return &application, nil
}

// Exists checks if the application Metadata already exists
func (e *ElasticProvider) Exists(appID *entities.ApplicationID) (bool, error) {
	res, err := e.client.Exists(e.indexName, e.CreateIDFromAppID(*appID))
	if err != nil {
		return false, err
	}

	switch res.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, nerrors.NewInternalError(res.Status())
	}

}

// Get returns the application metadata requested
func (e *ElasticProvider) Get(appID entities.ApplicationID) (*entities.ApplicationMetadata, error) {
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
		return nil, nerrors.NewNotFoundError("application metadata not found")
	}
	if r.Hits.Total.Value != 1 {
		log.Error().Str("catalogID", catalogID).Msg("Error getting application metadata, duplicated entries")
		return nil, nerrors.NewInternalError("Error getting application metadata, duplicated entries")
	}

	var application entities.ApplicationMetadata
	if err := json.Unmarshal(r.Hits.Hits[0].Source, &application); err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
	}

	return &application, nil

}

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
			return nerrors.NewNotFoundError("unable to delete the application metadata")
		}

		// Print the response status, number of results, and request duration.
		log.Debug().Str("Status", res.Status()).Int("total", d.Total).Int("took(ms)", d.Took).Msg("Delete operation")


	return nil
}
