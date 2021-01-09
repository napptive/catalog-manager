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
	"github.com/napptive/grpc-catalog-manager-go"
	"github.com/napptive/grpc-oam-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"google.golang.org/protobuf/encoding/protojson"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	"strings"
)

const ComponentKindStr  = "kind: Component"

// IsComponent check if a file contains a Component definition
func IsComponent(filepath string) (bool, *grpc_oam_go.Component, error) {

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return false, nil, nerrors.FromError(err)
	}

	// Find substring
	dataStr := string(data)
	if ind := strings.Index( dataStr, ComponentKindStr); ind != -1 {
		oam, err := DecodeComponent(data)
		if err != nil {
			return true, nil, err
		}
		return true, oam, nil
	}

	return false, nil, nil
}

// DecodeComponent read a yaml file and returns a OAMComponent
func DecodeComponent (data []byte) (*grpc_oam_go.Component, error) {
	// File > yaml k8s
	obj := &unstructured.Unstructured{}
	yamlDecoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(data)), 1024)
	err := yamlDecoder.Decode(obj)
	if err != nil {
		return nil, err
	}
	// yaml k8s > json

	to, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var result grpc_oam_go.Component
	customUnmarshaler := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}
	err = customUnmarshaler.Unmarshal(to, &result)
	if err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "cannot unmarshal OAM Component message")
	}

	return &result, nil
}

func IsYamlFile (filePath string) bool {
	return strings.Index(filePath, ".yaml") != -1
}

// OAMComponentToCatalogEntryResponse converts a Component to a CatalogEntryResponse
func ComponentToCatalogEntryResponse(catalogId string, entryId string, component grpc_oam_go.Component) *grpc_catalog_manager_go.CatalogEntryResponse{

	var name string
	if component.Metadata != nil {
		name = component.Metadata.Name
	}
	var image string
	if component.Spec != nil && component.Spec.Workload != nil  && component.Spec.Workload.Spec != nil && len(component.Spec.Workload.Spec.Containers) > 0{
		image = component.Spec.Workload.Spec.Containers[0].Image
	} else {
		if component.Spec != nil && component.Spec.Settings != nil && len(component.Spec.Settings.Containers) > 0 {
			image = component.Spec.Settings.Containers[0].Image
		}
	}

	return &grpc_catalog_manager_go.CatalogEntryResponse{
		CatalogId: catalogId,
		EntryId:   entryId,
		Name:      name,
		Image:     image,
		Version:   component.ApiVersion,
		Added:     false,
		Component: &component,
	}

}
