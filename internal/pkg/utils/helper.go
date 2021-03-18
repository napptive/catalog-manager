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
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/napptive/grpc-oam-go"
	"github.com/napptive/nerrors/pkg/nerrors"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const ComponentKindStr  = "kind: Component"
const ApiVersionStr = "apiVersion: core.oam.dev/v1alpha2"

// IsComponent check if a file contains a Component definition
func IsComponent(filepath string) (bool, *grpc_oam_go.Component, error) {

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return false, nil, nerrors.FromError(err)
	}

	// Find substring
	dataStr := string(data)
	ind := strings.Index( dataStr, ComponentKindStr)
	ind2 := strings.Index( dataStr, ApiVersionStr)
	if  ind != -1  && ind2 != -1{

		oam, err := DecodeComponent(data)
		if err != nil {
			log.Warn().Str("error", err.Error()).Str("file", filepath).Msg("error decoding component, decode it checking env variables")
			oam, err = DecodeComponentChecking(data)
			if err != nil {
				return true, nil, err
			}
		}
		return true, oam, nil
	}else{
		log.Debug().Str("file", filepath).Msg("Is not a component")
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

// check file extension and returns if is a yaml file
func IsYamlFile (filePath string) bool {
	return strings.Contains(filePath, ".yaml")
}



// DecodeComponentChecking try to returns an OAMComponent from a file checking the env variable value types
// Note: the env.value in component proto message is a string,
// if the value in the yaml file is an integer (a port for example)
// the unmarshall function does not works, this function checks the env types and convert it into string if needed
func DecodeComponentChecking(data []byte) (*grpc_oam_go.Component, error){

	// Find substring
	dataStr := string(data)
	// File > yaml k8s
	obj := &unstructured.Unstructured{}
	yamlDecoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(dataStr), 1024)
	err := yamlDecoder.Decode(obj)
	if err != nil {
		return nil, err
	}
	pp := obj.UnstructuredContent()

	// navigate thought the unstructured object
	interm, _ := pp["spec"].(map[string]interface{})
	interm, _ = interm["workload"].(map[string]interface{})
	interm, _ = interm["spec"].(map[string]interface{})
	containers, exists := interm["containers"]
	if exists {
		containerList := containers.([]interface{})
		for j := 0; j < len(containerList); j++ {
			envs, _ := containerList[j].(map[string]interface{})["env"].([]interface{})
			for i := 0; i < len(envs); i++ {
				env, _ := envs[i].(map[string]interface{})
				value := env["value"]
				switch value.(type) {
				case int32, int64:
					env["value"] = fmt.Sprintf("%d", value)
				case float32, float64:
					env["value"] = fmt.Sprintf("%f", value)
				}
			}
		}
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