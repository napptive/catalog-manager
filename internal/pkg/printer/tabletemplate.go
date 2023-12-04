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

package printer

import (
	"reflect"

	"github.com/napptive/grpc-catalog-common-go"
	"github.com/napptive/grpc-catalog-go"
	"github.com/napptive/nerrors/pkg/nerrors"
)

const ApplicationListTemplate = `APPLICATION	NAME
{{range $other, $app := .Applications}}{{fromApplicationSummary $app}}{{end}}`

// OpResponseTemplate with the table representation of an OpResponse.
const OpResponseTemplate = `STATUS	INFO
{{.StatusName}}	{{.UserInfo}}
`

// structTemplates map associating type and template to print it.
var structTemplates = map[reflect.Type]string{
	reflect.TypeOf(&grpc_catalog_go.ApplicationList{}):   ApplicationListTemplate,
	reflect.TypeOf(&grpc_catalog_common_go.OpResponse{}): OpResponseTemplate,
}

// GetTemplate returns a template to print an arbitrary structure in table format.
func GetTemplate(result interface{}) (*string, error) {
	template, exists := structTemplates[reflect.TypeOf(result)]
	if !exists {
		return nil, nerrors.NewUnimplementedError("no template is available to print %s", reflect.TypeOf(result).String())
	}
	return &template, nil
}
