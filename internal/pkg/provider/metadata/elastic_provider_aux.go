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
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
)

func (e *ElasticProvider) listFromWithFilter(filter *ListFilter, lastReceived int, getFields ...string) (*responseWrapper, error) {

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
	log.Debug().Str("Status", res.Status()).Int("total", r.Hits.Total.Value).Int("took(ms)", r.Took).Msg("List operation")

	return &r, nil
}

func (e *ElasticProvider) ListSummaryWithFilter(filter *ListFilter) ([]*entities.AppSummary, *entities.Summary, error) {
	lastReceived := 0
	query := true
	summaryList := make([]*entities.AppSummary, 0)
	var summary entities.Summary
	total := 0
	getFields := []string{"Namespace", "ApplicationName", "Tag", "MetadataName", "Metadata", "Private"}

	for query {
		r, err := e.listFromWithFilter(filter, lastReceived, getFields...)
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

func (e *ElasticProvider) GetPublicApps() []*entities.AppSummary {
	e.Lock()
	defer e.Unlock()
	return e.appCache
}

func (e *ElasticProvider) GetApplicationVisibility(namespace string, applicationName string) (*bool, error) {

	sortedBy := []string{"Namespace", "ApplicationName", "Tag"}
	searchFunctions := []func(*esapi.SearchRequest){
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.indexName),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithSize(1), // only one requested -> all the applications have the same visibility
		e.client.Search.WithSort(sortedBy...),
		e.client.Search.WithSourceIncludes("Namespace", "ApplicationName", "Tag", "Private"),
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
		return nil, nil
	}

	var application entities.ApplicationInfo
	if err := json.Unmarshal(r.Hits.Hits[0].Source, &application); err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "error unmarshalling application metadata")
	}
	return &application.Private, nil

}
