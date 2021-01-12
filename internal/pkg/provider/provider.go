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
	"github.com/napptive/grpc-oam-go"
)

// CatalogEntry with the entry information
type CatalogEntry struct {
	// EntryId with the entry identifier (for now, catalogName:pathFile)
	EntryId string
	// Component with the Component
	Component *grpc_oam_go.Component
}

// CatalogProvider with an interface that defines the provider methods
type CatalogProvider interface {
	// GetName returns the provider name
	GetName () string
	// GetComponents get the components from the repo and returns them
	GetComponents() ([]CatalogEntry, error)
	// EmptyCache removes the cloned repository
	EmptyCache() error
}

