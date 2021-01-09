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
	"github.com/rs/zerolog/log"
)


var _ = ginkgo.Describe("Handler test on manager", func() {
	var manager *ManagerProvider

	if !utils.RunIntegrationTests("provider") {
		log.Warn().Msg("Manager Provider tests are skipped")
		return
	}

	ginkgo.It("Should be able to create the catalog providers", func() {
		var err error
		dir := "./tmp/"
		cmFilePath := "./config_test"
		manager, err = NewManagerProvider(cmFilePath, dir)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(manager).NotTo(gomega.BeNil())

		err = manager.Init()
		gomega.Expect(err).Should(gomega.Succeed())

		components := manager.GetCatalog()
		gomega.Expect(components).ShouldNot(gomega.BeNil())

		err = manager.EmptyRepositories()
		gomega.Expect(err).Should(gomega.Succeed())
	})

})
