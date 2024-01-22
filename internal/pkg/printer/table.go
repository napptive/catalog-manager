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
	"fmt"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/napptive/grpc-catalog-go"
	"github.com/napptive/nerrors/pkg/nerrors"
)

const (
	// MinWidth is the minimal cell width including any padding.
	MinWidth = 8
	// TabWidth is the width of tab characters (equivalent number of spaces)
	TabWidth = 4
	// Padding added to a cell before computing its width
	Padding = 4
	// PaddingChar with the ASCII char used for padding
	PaddingChar = ' '
	// TabWriterFlags with the formatting options.
	TabWriterFlags = 0
)

// TablePrinter structure with the implementation required to print in a human readable table format a given result.
type TablePrinter struct {
}

// NewTablePrinter builds a new ResultPrinter whose output is a human readable table-like representation of the object.
func NewTablePrinter() ResultPrinter {
	return &TablePrinter{}
}

func (tp *TablePrinter) toString(content []byte) string {
	return string(content)
}

// fromApplicationSummary composes the application in a catalog as
// namespace/appName:tag Medatada_Name
func (tp *TablePrinter) fromApplicationSummary(app *grpc_catalog_go.ApplicationSummary) string {

	var result string
	for version, name := range app.TagMetadataName {
		result += fmt.Sprintf("%s/%s:%s\t%s\n", app.Namespace, app.ApplicationName, version, name)
	}
	return result
}

// Print the result.
func (tp *TablePrinter) Print(result interface{}) error {
	associatedTemplate, err := GetTemplate(result)
	if err != nil {
		return err
	}
	t := template.New("TablePrinter").Funcs(template.FuncMap{
		"toString":               tp.toString,
		"fromApplicationSummary": tp.fromApplicationSummary,
	})
	t, err = t.Parse(*associatedTemplate)
	if err != nil {
		return nerrors.NewInternalErrorFrom(err, "cannot apply template")
	}
	w := tabwriter.NewWriter(os.Stdout, MinWidth, TabWidth, Padding, PaddingChar, TabWriterFlags)
	if err := t.Execute(w, result); err != nil {
		return err
	}
	w.Flush()
	return nil
}
