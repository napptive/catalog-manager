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
	"github.com/napptive/catalog-manager/internal/pkg/entities"
)

type ListFilter struct {
	Namespace *string
	Private   *bool
}

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

// MetadataProvider is an interface with the methods of a metadata provider must implement
type MetadataProvider interface {
	// Add stores new application metadata or updates it if it exists
	Add(metadata *entities.ApplicationInfo) (*entities.ApplicationInfo, error)
	// Get returns the application metadata requested or an error if it does not exist
	Get(appID *entities.ApplicationID) (*entities.ApplicationInfo, error)
	// Exists checks if an application metadata exists
	Exists(appID *entities.ApplicationID) (bool, error)
	// Remove removes an application metadata
	Remove(appID *entities.ApplicationID) error
	// List returns the applications stored
	List(namespace string) ([]*entities.ApplicationInfo, error)
	// ListSummary returns a list of application summaries
	ListSummary(namespace string) ([]*entities.AppSummary, error)
	// GetSummary returns the catalog summary
	GetSummary() (*entities.Summary, error)

	ListSummaryWithFilter(filter *ListFilter) ([]*entities.AppSummary, *entities.Summary, error)
	GetPublicApps() []*entities.AppSummary
	// GetApplicationVisibility returns is an application is private or not or nil if the application does not exist
	GetApplicationVisibility(namespace string, applicationName string) (*bool, error)
}
