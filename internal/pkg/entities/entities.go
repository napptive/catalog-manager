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
package entities

import (
	"fmt"

	grpc_catalog_go "github.com/napptive/grpc-catalog-go"
)

type AppSummary struct {
	// Namespace with the namespace of the application
	Namespace string
	// ApplicationName with the name of the application
	ApplicationName string
	// TagMetadataName with the MetadataName indexed by Tag
	TagMetadataName map[string]string
	// MetadataLogo with the ApplicationLogo indexed by Tag
	MetadataLogo map[string][]ApplicationLogo
}

// ToApplicationSummary converts the ApplicationSummary to grpc_catalog_go.ApplicationSummary
func (a *AppSummary) ToApplicationSummary() *grpc_catalog_go.ApplicationSummary {

	// map[string][]ApplicationLogo TO map[string]* Logo []*ApplicationLogo

	logoSummary := make(map[string]*grpc_catalog_go.ApplicationLogoList)
	for key, value := range a.MetadataLogo {
		logoList := make([]*grpc_catalog_go.ApplicationLogo, 0)
		for _, logo := range value {
			logoList = append(logoList, logo.ToGRPC())
		}
		logoSummary[key] = &grpc_catalog_go.ApplicationLogoList{Logo: logoList}
	}

	return &grpc_catalog_go.ApplicationSummary{
		Namespace:              a.Namespace,
		ApplicationName:        a.ApplicationName,
		TagMetadataName:        a.TagMetadataName,
		SummaryApplicationLogo: logoSummary,
	}
}

// Summary with catalog summary
type Summary struct {
	NumNamespaces   int
	NumApplications int
	NumTags         int
}

// ToSummaryResponse converts Summary to GRPC
func (s *Summary) ToSummaryResponse() *grpc_catalog_go.SummaryResponse {
	if s == nil {
		return nil
	}
	return &grpc_catalog_go.SummaryResponse{
		NumNamespaces:   int32(s.NumNamespaces),
		NumApplications: int32(s.NumApplications),
		NumTags:         int32(s.NumTags),
	}
}

type ExtendedAppSummary struct {
	// Namespace with the namespace of the application
	Namespace string
	// ApplicationName with the name of the application
	ApplicationName string
	// TagMetadataName with the MetadataName indexed by Tag
	Tag          string
	MetadataName string
	Metadata     string
}

// -- ApplicationMetadata

// ApplicationInfo with the metadata of application, this will be the application info showed
type ApplicationInfo struct {
	// CatalogID with an internal identifier
	CatalogID string
	// Namespace where the application is located.
	Namespace string
	// ApplicationName with the name of the application
	ApplicationName string
	// Tag with the tag/version of the application
	Tag string
	// Readme with the content of the README file
	Readme string
	// Metadata with the metadata.yaml file
	Metadata string
	//MedataName with the name defined in metadata file. This field is used to store it in elastic field and return it when listing
	MetadataName string
}

// ToApplicationID converts ApplicationSummary to ApplicationID
func (a *ApplicationInfo) ToApplicationID() *ApplicationID {
	return &ApplicationID{
		Namespace:       a.Namespace,
		ApplicationName: a.ApplicationName,
		Tag:             a.Tag,
	}
}

// --

// -- ApplicationID

// ApplicationID with the application identifier (catalogURL-Namespace-AppName-tag)
// these four fields must be unique
type ApplicationID struct {
	// Namespace associated with the application.
	Namespace string
	// ApplicationName with the name of the application
	ApplicationName string
	// Tag with the tag/version of the application
	Tag string
}

func (a *ApplicationID) String() string {
	return fmt.Sprintf("%s/%s:%s", a.Namespace, a.ApplicationName, a.Tag)
}

// --

// -- CatalogMetadata

// ApplicationLogo represents the application logo
type ApplicationLogo struct {
	// Src with the src URL
	Src string `yaml:"src"`
	// Type with the logo type (p.e: image/png)
	Type string `yaml:"type"`
	// Size with the logo size (p.e. 120x120)
	Size string `yaml:"size"`
}

// ToGRPC converts ApplicationLogo to grpc_catalog_go.ApplicationLogo
func (al *ApplicationLogo) ToGRPC() *grpc_catalog_go.ApplicationLogo {
	return &grpc_catalog_go.ApplicationLogo{
		Src:  al.Src,
		Type: al.Type,
		Size: al.Size,
	}
}

// KubernetesEntities with the application K8s entities
type KubernetesEntities struct {
	// ApiVersion with the entity version
	ApiVersion string `yaml:"apiVersion"`
	// Kind with the entity type
	Kind string `yaml:"kind"`
	// Name with the entity name
	Name string `yaml:"name"`
}

// ToGRPC converts KubernetesEntities to *grpc_catalog_go.KubernetesEntities
func (k *KubernetesEntities) ToGRPC() *grpc_catalog_go.KubernetesEntities {
	return &grpc_catalog_go.KubernetesEntities{
		ApiVersion: k.ApiVersion,
		Kind:       k.Kind,
		Name:       k.Name,
	}
}

// ApplicationRequirement with the application requirements
type ApplicationRequirement struct {
	// Traits with the application traits
	Traits []string `yaml:"traits"`
	// Scopes with the application scopes
	Scopes []string `yaml:"scopes"`
	// K8s with all the K8s entities needed
	K8s []KubernetesEntities `yaml:"k8s"`
}

// ToGRPC converts CatalogRequirement to grpc_catalog_go.CatalogRequirement
func (ar *ApplicationRequirement) ToGRPC() *grpc_catalog_go.ApplicationRequirement {
	k8s := make([]*grpc_catalog_go.KubernetesEntities, 0)
	for _, entity := range ar.K8s {
		k8s = append(k8s, entity.ToGRPC())
	}

	return &grpc_catalog_go.ApplicationRequirement{
		Traits: ar.Traits,
		Scopes: ar.Scopes,
		K8S:    k8s,
	}
}

// ApplicationMetadata is a struct to load the kind and api_version of a file to check if it is metadata file
type ApplicationMetadata struct {
	APIVersion  string                 `yaml:"apiVersion"`
	Kind        string                 `yaml:"kind"`
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Description string                 `yaml:"description"`
	Keywords    []string               `yaml:"keywords"`
	License     string                 `yaml:"license"`
	Url         string                 `yaml:"url"`
	Doc         string                 `yaml:"doc"`
	Requires    ApplicationRequirement `yaml:"requires"`
	Logo        []ApplicationLogo      `yaml:"logo"`
}

// ToGRPC converts ApplicationMetadata to grpc_catalog_go.ApplicationMetadata
func (am *ApplicationMetadata) ToGRPC() *grpc_catalog_go.ApplicationMetadata {
	logos := make([]*grpc_catalog_go.ApplicationLogo, 0)
	for _, logo := range am.Logo {
		logos = append(logos, logo.ToGRPC())
	}
	return &grpc_catalog_go.ApplicationMetadata{
		ApiVersion:  am.APIVersion,
		Kind:        am.Kind,
		Name:        am.Name,
		Version:     am.Version,
		Description: am.Description,
		Keywords:    am.Keywords,
		License:     am.License,
		Url:         am.Url,
		Doc:         am.Doc,
		Requires:    am.Requires.ToGRPC(),
		Logo:        logos,
	}
}

// -- ExtendedApplicationMetadata
// ExtendedApplicationMetadata is an object with ApplicationMetadata and a field with metadata parsed
type ExtendedApplicationMetadata struct {
	// CatalogID with an internal identifier
	CatalogID string
	// Namespace associated with the application.
	Namespace string
	// ApplicationName with the name of the application
	ApplicationName string
	// Tag with the tag/version of the application
	Tag string
	// Readme with the content of the README file
	Readme string
	// Metadata with the metadata.yaml file
	Metadata string
	// MetadataObj with the metadata object
	MetadataObj ApplicationMetadata
}

// --

// -- FileInfo

// FileInfo represents a file
type FileInfo struct {
	// path with the File path
	Path string
	// data with the content of the file
	Data []byte
}

// NewFileInfo creates FileInfo from *grpc_catalog_go.FileInfo
func NewFileInfo(info *grpc_catalog_go.FileInfo) *FileInfo {
	return &FileInfo{
		Path: info.Path,
		Data: info.Data,
	}
}

// ToGRPC converts FileInfo to grpc_catalog_go.FileInfo
func (fi *FileInfo) ToGRPC() *grpc_catalog_go.FileInfo {
	return &grpc_catalog_go.FileInfo{
		Path: fi.Path,
		Data: fi.Data,
	}
}

// --
