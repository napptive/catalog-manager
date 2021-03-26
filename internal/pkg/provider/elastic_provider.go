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
	"net/url"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/nerrors/pkg/nerrors"

	"github.com/rs/zerolog/log"
)

const (
	urlField         = "Url"
	RepositoryField  = "Repository"
	ApplicationField = "ApplicationName"
)

// envelopeResponse is an struct used to load a search result
type envelopeResponse struct {
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

// CreateIndex creates an index with the mapping received
func (e *ElasticProvider) CreateIndex(mapping string) error {

	exists, err := esapi.IndicesExistsRequest{
		Index: []string{e.indexName},
	}.Do(context.Background(), e.client)

	if err != nil {
		return err
	}
	if exists.IsError() {
		switch exists.StatusCode {
		case 404:
			res, err := e.client.Indices.Create(e.indexName, e.client.Indices.Create.WithBody(strings.NewReader(mapping)))
			if err != nil {
				return err
			}
			if res.IsError() {
				log.Warn().Str("err", res.String()).Msg("error creating index")
				return nerrors.NewInternalError("error creating index")
			}
			defer res.Body.Close()
		default:
			return nerrors.NewInternalError("error checking index. %s", exists.Status())
		}

	}
	defer exists.Body.Close()

	return nil
}

// DeleteIndex removes a elastic index
func (e *ElasticProvider) DeleteIndex() error {
	if _, err := e.client.Indices.Delete([]string{e.indexName}); err != nil {
		return err
	}
	return nil
}

// CreateID generates the documentID to store the application metadata
// The url must be escaped
func (e *ElasticProvider) CreateID(metadata entities.ApplicationMetadata) string {
	id := fmt.Sprintf("%s%s%s%s", url.PathEscape(metadata.Url), metadata.Repository, metadata.ApplicationName, metadata.Tag)
	return id
}

// CreateIDFromAppID creates the documentID from ApplicationID
func (e *ElasticProvider) CreateIDFromAppID(metadata entities.ApplicationID) string {
	id := fmt.Sprintf("%s%s%s%s", url.PathEscape(metadata.Url), metadata.Repository, metadata.ApplicationName, metadata.Tag)
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

	url := e.buildTerm(urlField, appID.Url)
	repo := e.buildTerm(RepositoryField, appID.Repository)
	appName := e.buildTerm(ApplicationField, appID.ApplicationName)
	tag := e.buildTerm("Tag", appID.Tag)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{url, repo, appName, tag},
			},
		},
	}
	return query
}

// Add stores new application metadata or updates it if it exists
func (e *ElasticProvider) Add(metadata entities.ApplicationMetadata) error {

	// convert the metadata to JSON
	metadataJSON, err := utils.ApplicationMetadataToJSON(metadata)
	if err != nil {
		log.Error().Err(err).Msg("error converting metadata to JSON")
		return err
	}
	req := esapi.IndexRequest{
		Index:      e.indexName,
		DocumentID: e.CreateID(metadata),
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
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Warn().Str("err", res.Status()).Msg("Error indexing document")
		return nerrors.NewInternalError("Error indexing document: [%s]", res.Status())

	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}

	return nil
}

// Get returns the application metadata requested
func (e *ElasticProvider) Get(appID entities.ApplicationID) (*entities.ApplicationMetadata, error) {

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

	var r envelopeResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	// Print the response status, number of results, and request duration.
	log.Debug().Str("Status", res.Status()).Int("total", r.Hits.Total.Value).Int("took(ms)", r.Took).Msg("request operation")

	if r.Hits.Total.Value == 0 {
		return nil, nerrors.NewNotFoundError("application metadata not found")
	}

	var application entities.ApplicationMetadata
	if err := json.Unmarshal(r.Hits.Hits[0].Source, &application); err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
	}

	// Print the ID and document source for each hit.
	for _, hit := range r.Hits.Hits {
		log.Debug().Str("ID", hit.ID).Str("source", string(hit.Source)).Msg("Application")
	}

	return &application, nil
}

// Exists checks if the application Metadata already exists
func (e *ElasticProvider) Exists(appID entities.ApplicationID) (bool, error) {
	res, err := e.client.Exists(e.indexName, e.CreateIDFromAppID(appID))
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

// GetByID return the application metadata requested
func (e *ElasticProvider) GetByID(appID entities.ApplicationID) (*entities.ApplicationMetadata, error) {

	var buf bytes.Buffer

	query := e.buildQuery(appID)
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Err(err).Msg("Error encoding query")
	}

	// Perform the search request.
	var req = esapi.GetRequest{Index: e.indexName, DocumentID: e.CreateIDFromAppID(appID)}
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		log.Err(err).Msg("Error getting response")
		return nil, err
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

	type envelopeResponse struct {
		ID     string          `json:"_id"`
		Source json.RawMessage `json:"_source"`
	}

	//var r  map[string]interface{}
	var r envelopeResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Err(err).Msg("Error parsing the response body")
		return nil, err
	}
	var application entities.ApplicationMetadata
	if err := json.Unmarshal(r.Source, &application); err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
	}

	log.Debug().Interface("application", application).Msg("---")

	return &application, nil

}
