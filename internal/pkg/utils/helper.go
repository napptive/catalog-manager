/**
 * Copyright 2023 Napptive
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
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"sigs.k8s.io/yaml"
	"strings"

	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

const (
	// defaultVersion with the version ofa an application if it is no filled
	defaultVersion = "latest"
)

// metadataGKV with a map associating group/version with the object kind. This map contains all the version of a metadata
// file that are supported by the catalog. Notice that OAM has not yet accepted the ApplicationMetadata proposal.
var metadataGKV = map[string]string{
	"core.napptive.com/v1alpha1": "ApplicationMetadata",
	"core.oam.dev/v1alpha1":      "ApplicationMetadata"}

// IsYamlFile check file extension and returns if is a yaml file
func IsYamlFile(filePath string) bool {
	return strings.Contains(filePath, ".yaml") || strings.Contains(filePath, ".yml")
}

// ApplicationInfoToJSON converts an ApplicationInfo struct into a JSON
func ApplicationInfoToJSON(metadata entities.ApplicationInfo) (string, error) {
	bRes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(bRes), nil
}

// GetFile looks for a file by name in the array retrieved and returns the data or nil if the file does not exist
func GetFile(relativeFileName string, files []*entities.FileInfo) []byte {
	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file.Path), strings.ToLower(relativeFileName)) {
			return file.Data
		}
	}
	return []byte{}
}

// IsMetadata checks if the file is metadata file and returns it
func IsMetadata(data []byte) (bool, *entities.ApplicationMetadata, error) {
	var a entities.ApplicationMetadata
	err := yaml.Unmarshal(data, &a)
	if err != nil {
		log.Err(err).Msg("error getting metadata")
		return false, nil, err
	}
	// Iterate through the list of valid metadata file definitions.
	for gv, k := range metadataGKV {
		if a.APIVersion == gv && a.Kind == k {
			return true, &a, nil
		}
	}
	return false, nil, nil
}

// GenerateRandomString is a method to generate a random string with a determinate length
func GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b)[:length], nil
}

// DecomposeApplicationID extracts the catalog URL, namespace, and application name
// from an application identifier in the form of:
// [catalogURL/]namespace/appName[:tag]
func DecomposeApplicationID(applicationID string) (string, *entities.ApplicationID, error) {
	var version string
	var applicationName string
	var namespace string
	catalogURL := ""

	elements := strings.Split(applicationID, "/")
	if len(elements) != 2 && len(elements) != 3 {
		return "", nil, nerrors.NewFailedPreconditionError(
			"incorrect format for application name. [catalogURL/]namespace/appName[:tag]")
	}

	// if len == 2 -> no url informed.
	if len(elements) == 3 {
		catalogURL = elements[0]
	}
	namespace = elements[len(elements)-2]

	// get the version -> appName[:tag]
	sp := strings.Split(elements[len(elements)-1], ":")
	if len(sp) == 1 {
		applicationName = sp[0]
		version = defaultVersion
	} else if len(sp) == 2 {
		applicationName = sp[0]
		version = sp[1]
		if strings.Trim(version, " ") == "" {
			version = defaultVersion
		}
	} else {
		return "", nil, nerrors.NewFailedPreconditionError(
			"incorrect format for application name. [catalogURL/]namespace/appName[:tag]")
	}

	return catalogURL, &entities.ApplicationID{
		Namespace:       namespace,
		ApplicationName: applicationName,
		Tag:             version,
	}, nil
}

// GetGvk returns the GroupVersionKind from a yaml file
func GetGvk(inputYAML []byte) (*schema.GroupVersionKind, error) {
	// - Decode YAML manifest into unstructured.Unstructured
	var decUnstructured = k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(inputYAML, nil, obj)
	if err != nil {
		return nil, err
	}
	return gvk, nil
}
