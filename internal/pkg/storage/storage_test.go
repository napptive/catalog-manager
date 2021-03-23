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
package storage

import (
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	grpc_catalog_go "github.com/napptive/grpc-catalog-go"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"syreclabs.com/go/faker"
)

var _ = ginkgo.Describe("Elastic Provider test", func() {

	if !utils.RunIntegrationTests("storage")  {
		log.Warn().Msg("Storage manager tests are skipped")
		return
	}

	var basePath = "/tmp/repositories"

	ginkgo.It("should be able to create a repository", func() {
		manager := NewStorageManager(basePath)
		err := manager.CreateRepository(faker.Name().FirstName())
		gomega.Expect(err).Should(gomega.Succeed())
	})

	ginkgo.It("should be able to check if a repository does not exist", func() {
		manager := NewStorageManager(basePath)
		exist, err := manager.RepositoryExists(faker.Name().FirstName())
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(exist).ShouldNot(gomega.BeTrue())
	})

	ginkgo.It("should be able to check if a repository exists", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		err := manager.CreateRepository(repo)
		gomega.Expect(err).Should(gomega.Succeed())

		exist, err := manager.RepositoryExists(repo)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(exist).Should(gomega.BeTrue())
	})

	ginkgo.It("should be able to remove a repository ", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		err := manager.CreateRepository(repo)
		gomega.Expect(err).Should(gomega.Succeed())

		err = manager.RemoveRepository(repo)
		gomega.Expect(err).Should(gomega.Succeed())
	})

	ginkgo.It("Should be able to add an application", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		appName := faker.App().Name()
		files := []grpc_catalog_go.FileInfo {
			{Path: "app_config.yaml", Data: []byte("appconf")},
			{Path: "component1.yaml", Data: []byte("component1")},
			{Path: "component2.yaml", Data: []byte("component2")}}
		err := manager.StorageApplication(repo, appName, "latest", files)
		gomega.Expect(err).Should(gomega.Succeed())
	})

	ginkgo.It("Should be able to add two versions of an application", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		appName := faker.App().Name()
		files := []grpc_catalog_go.FileInfo {
			{Path: "app_config.yaml", Data: []byte("appconf")},
			{Path: "component1.yaml", Data: []byte("component1")},
			{Path: "component2.yaml", Data: []byte("component2")}}
		err := manager.StorageApplication(repo, appName, "latest", files)
		gomega.Expect(err).Should(gomega.Succeed())

		err = manager.StorageApplication(repo, appName, "v0.0.1", files)
		gomega.Expect(err).Should(gomega.Succeed())
	})

})