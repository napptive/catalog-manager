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
package utils

import (
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/yaml"
	"strings"

	"encoding/json"
)

const (
	// apiMetadataVersion with the version of the metadata entity
	apiMetadataVersion = "core.oam.dev/v1alpha1"
	// appMetadataKind with the kind of the metadata entity
	appMetadataKind = "ApplicationMetadata"
)

// check file extension and returns if is a yaml file
func IsYamlFile(filePath string) bool {
	return strings.Contains(filePath, ".yaml") || strings.Contains(filePath, ".yml")
}

// ApplicationMetadataToJSON converts an ApplicationMetadata struct into a JSON
func ApplicationMetadataToJSON(metadata entities.ApplicationMetadata) (string, error) {

	bRes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(bRes), nil
}

// getFile looks for a file by name in the array retrieved and returns the data or nil if the file does not exist
func GetFile(relativeFileName string, files []*entities.FileInfo) []byte {

	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file.Path), strings.ToLower(relativeFileName)) {
			return file.Data
		}
	}

	return []byte{}
}

// IsMetadata checks if the file is metadata file and returns it
func IsMetadata(data []byte) (bool, *entities.CatalogMetadata, error) {
	var a entities.CatalogMetadata
	err := yaml.Unmarshal(data, &a)
	if err != nil {
		log.Err(err).Msg("error getting metadata")
		return false, nil, nerrors.FromError(err)
	}
	if a.APIVersion == apiMetadataVersion && a.Kind == appMetadataKind {
		return true, &a, nil
	}

	return false, nil, nil
}
