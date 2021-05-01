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

package utils

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Utils test", func() {

	ginkgo.Context("with the application identifier", func() {

		ginkgo.It("application identifier should be decomposed", func() {

			var testCases = map[string]struct {
				catalogURL      string
				namespace       string
				applicationName string
				tag             string
			}{
				"namespace/app:latest":         {"", "namespace", "app", "latest"},
				"namespace/app":                {"", "namespace", "app", "latest"},
				"catalog/namespace/app:latest": {"catalog", "namespace", "app", "latest"},
				"catalog/namespace/app":        {"catalog", "namespace", "app", "latest"},
			}

			for appID, expectedResult := range testCases {
				catalog, appID, err := DecomposeApplicationID(appID)
				gomega.Expect(err).To(gomega.Succeed())
				gomega.Expect(catalog).To(gomega.Equal(expectedResult.catalogURL))
				gomega.Expect(appID.Namespace).To(gomega.Equal(expectedResult.namespace))
				gomega.Expect(appID.ApplicationName).To(gomega.Equal(expectedResult.applicationName))
				gomega.Expect(appID.Tag).To(gomega.Equal(expectedResult.tag))
			}

		})

	})
})
