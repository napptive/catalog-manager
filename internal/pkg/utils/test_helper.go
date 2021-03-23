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
	"os"
	"syreclabs.com/go/faker"
)

// RunIntegrationTests checks whether integration tests should be executed.
func RunIntegrationTests(id string) bool {
	var runIntegration = os.Getenv("RUN_INTEGRATION_TEST")
	if runIntegration == "all" {
		return true
	}
	return runIntegration == id
}


func CreateApplicationMetadata () *entities.ApplicationMetadata {

	return &entities.ApplicationMetadata{
		Url:             faker.Internet().Url(),
		Repository:      faker.Name().FirstName(),
		ApplicationName: faker.App().Name(),
		Tag:             faker.App().Version(),
		Readme:          faker.Lorem().Paragraph(10),
		Metadata:        faker.Lorem().Paragraph(10),
	}
}