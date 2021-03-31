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

import "github.com/napptive/grpc-catalog-go"

// -- ApplicationMetadata

// ApplicationMetadata with the metadata of application, this will be the application info showed
type ApplicationMetadata struct {
	// CatalogID with an internal identifier
	CatalogID string
	// Repository with the repository name
	Repository string
	// ApplicationName with the name of the application
	ApplicationName string
	// Tag with the tag/version of the application
	Tag string
	// Readme with the content of the README file
	Readme string
	// Metadata with the metadata.yaml file
	Metadata string
	//MedataName with the name defined in metadata file
	MetadataName string
	// MetadataObj with the metadata object
	MetadataObj CatalogMetadata
}

func (a *ApplicationMetadata) ToApplicationSummary() *grpc_catalog_go.ApplicationSummary {
	return &grpc_catalog_go.ApplicationSummary{
		RepositoryName:  a.Repository,
		ApplicationName: a.ApplicationName,
		Tag:             a.Tag,
		MetadataName:    a.MetadataName,
	}
}

func (a *ApplicationMetadata) ToApplicationID() *ApplicationID {
	return &ApplicationID{
		Repository:      a.Repository,
		ApplicationName: a.ApplicationName,
		Tag:             a.Tag,
	}
}

// --

// -- ApplicationID

// ApplicationID with the application identifier (URL-Repo-AppName-tag)
// these four fields must be unique
type ApplicationID struct {
	// Repository with the repository name
	Repository string
	// ApplicationName with the name of the application
	ApplicationName string
	// Tag with the tag/version of the application
	Tag string
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

func (k *KubernetesEntities) ToGRPC() *grpc_catalog_go.KubernetesEntities {
	return &grpc_catalog_go.KubernetesEntities{
		ApiVersion: k.ApiVersion,
		Kind:       k.Kind,
		Name:       k.Name,
	}
}

// CatalogRequirement with the application requirements
type CatalogRequirement struct {
	// Traits with the application traits
	Traits []string `yaml:"traits"`
	// Scopes with the application scopes
	Scopes []string `yaml:"scopes"`
	// K8s with all the K8s entities needed
	K8s []KubernetesEntities `yaml:"k8s"`
}

func (cr *CatalogRequirement) ToGRPC() *grpc_catalog_go.CatalogRequirement {
	k8s := make ([]*grpc_catalog_go.KubernetesEntities, 0)
	for _, entity := range cr.K8s {
		k8s = append(k8s, entity.ToGRPC())
	}

	return &grpc_catalog_go.CatalogRequirement{
		Traits: cr.Traits,
		Scopes: cr.Scopes,
		K8S:    k8s,
	}
}

// CatalogMetadata is a struct to load the kind and api_version of a file to check if it is metadata file
type CatalogMetadata struct {
	APIVersion  string             `yaml:"apiVersion"`
	Kind        string             `yaml:"kind"`
	Name        string             `yaml:"name"`
	Version     string             `yaml:"version"`
	Description string             `yaml:"description"`
	Tags        []string           `yaml:"tags"`
	License     string             `yaml:"license"`
	Url         string             `yaml:"url"`
	Doc         string             `yaml:"doc"`
	Requires    CatalogRequirement `yaml:"requires"`
	Logo        []ApplicationLogo  `yaml:"logo"`
}

func (c *CatalogMetadata) ToGRPC() *grpc_catalog_go.CatalogMetadata {
	logos := make ([]*grpc_catalog_go.ApplicationLogo, 0)
	for _, logo := range c.Logo {
		logos = append(logos, logo.ToGRPC())
	}
	return &grpc_catalog_go.CatalogMetadata{
		ApiVersion:  c.APIVersion,
		Kind:        c.Kind,
		Name:        c.Name,
		Version:     c.Version,
		Description: c.Description,
		Tags:        c.Tags,
		License:     c.License,
		Url:         c.Url,
		Doc:         c.Doc,
		Requires:    c.Requires.ToGRPC(),
		Logo:        logos,
	}
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

func NewFileInfo(info *grpc_catalog_go.FileInfo) *FileInfo {
	return &FileInfo{
		Path: info.Path,
		Data: info.Data,
	}
}

func (fi *FileInfo) ToGRPC() *grpc_catalog_go.FileInfo {
	return &grpc_catalog_go.FileInfo{
		Path: fi.Path,
		Data: fi.Data,
	}
}

// --
