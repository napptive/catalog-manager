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
package provider

import (
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Decode tests", func() {

	ginkgo.It("should be able to decode a component with a int in a environment variables", func() {

		oam, err := utils.DecodeComponentChecking(component)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(oam).ShouldNot(gomega.BeNil())

	})

	ginkgo.It("should not be able to decode a component with a int in a environment variables", func() {

		_, err := utils.DecodeComponent(component)
		gomega.Expect(err).ShouldNot(gomega.Succeed())

	})
})
