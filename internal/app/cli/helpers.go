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
package cli

import (
	"fmt"
	"github.com/napptive/catalog-manager/internal/pkg/printer"

	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog"
)

// PrintResultOrError prints the result using a given printer or the error.
func PrintResultOrError(result interface{}, err error) {
	if err != nil {
		if zerolog.GlobalLevel() == zerolog.DebugLevel {
			fmt.Println(nerrors.FromError(err).StackTraceToString())
		} else {
			fmt.Println(err.Error())
		}
	} else {
		if pErr := printer.NewTablePrinter().Print(result); pErr != nil {
			if zerolog.GlobalLevel() == zerolog.DebugLevel {
				fmt.Println(nerrors.FromError(pErr).StackTraceToString())
			} else {
				fmt.Println(pErr.Error())
			}
		}
	}
}