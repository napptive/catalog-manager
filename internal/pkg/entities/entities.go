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
}

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

// AppHeader is a struct to load the kind and api_version of a file to check if it is metadata file
type AppHeader struct {
	APIVersion  string `yaml:"apiVersion"`
	Kind        string `yaml:"kind"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

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
