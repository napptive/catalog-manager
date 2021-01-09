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
package entities

type Component struct {
	ApiVersion string `protobuf:"bytes,1,opt,name=api_version,json=apiVersion,proto3" json:"api_version,omitempty"`
	Kind       string `protobuf:"bytes,1,opt,name=kind,json=kind,proto3" json:"kind,omitempty"`
	//Metadata   string `protobuf:"bytes,1,opt,name=metadata,json=metadata,proto3" json:"metadata,omitempty"`
	//Spec       string `protobuf:"bytes,1,opt,name=spec,json=spec,proto3" json:"spec,omitempty"`
}

type ComponentEntry struct {
	Id        string
	Component Component
}
