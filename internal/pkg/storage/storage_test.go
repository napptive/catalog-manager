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
	"fmt"
	"os"

	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"syreclabs.com/go/faker"
)

var _ = ginkgo.Describe("Storage test", func() {

	if !utils.RunIntegrationTests("storage") {
		log.Warn().Msg("Storage manager tests are skipped")
		return
	}

	var basePath = os.Getenv("REPO_BASE_PATH")
	if basePath == "" {
		basePath = "/tmp/cmtest"
	}

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
		files := []*entities.FileInfo{
			{Path: "app_config.yaml", Data: []byte("appconf")},
			{Path: "component1.yaml", Data: []byte("component1")},
			{Path: "component2.yaml", Data: []byte("component2")}}
		err := manager.StoreApplication(repo, appName, "latest", files)
		gomega.Expect(err).Should(gomega.Succeed())
	})

	ginkgo.It("Should be able to add two versions of an application", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		appName := faker.App().Name()
		files := []*entities.FileInfo{
			{Path: "app_config.yaml", Data: []byte("appconf")},
			{Path: "component1.yaml", Data: []byte("component1")},
			{Path: "component2.yaml", Data: []byte("component2")}}
		err := manager.StoreApplication(repo, appName, "latest", files)
		gomega.Expect(err).Should(gomega.Succeed())

		err = manager.StoreApplication(repo, appName, "v0.0.1", files)
		gomega.Expect(err).Should(gomega.Succeed())
	})

	ginkgo.It("Should be able to find a repository", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		appName := faker.App().Name()
		files := []*entities.FileInfo{
			{Path: "app_config.yaml", Data: []byte("appconf")},
			{Path: "component1.yaml", Data: []byte("component1")},
			{Path: "component2.yaml", Data: []byte("component2")}}
		err := manager.StoreApplication(repo, appName, "latest", files)
		gomega.Expect(err).Should(gomega.Succeed())

		returned, err := manager.GetApplication(repo, appName, "latest", false)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(returned).ShouldNot(gomega.BeNil())
		gomega.Expect(returned).ShouldNot(gomega.BeEmpty())
		gomega.Expect(len(returned)).Should(gomega.Equal(len(files)))
	})

	ginkgo.It("Should be able to check if an application exists", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		appName := faker.App().Name()
		version := "latest"
		files := []*entities.FileInfo{
			{Path: "app_config.yaml", Data: []byte("appconf")},
			{Path: "component1.yaml", Data: []byte("component1")},
			{Path: "component2.yaml", Data: []byte("component2")}}
		err := manager.StoreApplication(repo, appName, version, files)
		gomega.Expect(err).Should(gomega.Succeed())

		exists, err := manager.ApplicationExists(repo, appName, version)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(exists).Should(gomega.BeTrue())

	})

	ginkgo.It("Should be able to check if an application does not exist", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		appName := faker.App().Name()
		version := "latest"

		exists, err := manager.ApplicationExists(repo, appName, version)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(exists).ShouldNot(gomega.BeTrue())
	})

	ginkgo.It("should be able to delete an application", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		appName := faker.App().Name()
		version := "latest"
		files := []*entities.FileInfo{
			{Path: "app_config.yaml", Data: []byte("appconf")},
			{Path: "component1.yaml", Data: []byte("component1")},
			{Path: "component2.yaml", Data: []byte("component2")}}
		err := manager.StoreApplication(repo, appName, version, files)
		gomega.Expect(err).Should(gomega.Succeed())

		err = manager.RemoveApplication(repo, appName, version)
		gomega.Expect(err).Should(gomega.Succeed())

		exists, err := manager.ApplicationExists(repo, appName, version)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(exists).ShouldNot(gomega.BeTrue())
	})

	ginkgo.It("should not be able to delete an application if it does not exists", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		appName := faker.App().Name()
		version := "latest"

		err := manager.RemoveApplication(repo, appName, version)
		gomega.Expect(err).ShouldNot(gomega.Succeed())

	})

	ginkgo.It("Should be able to download an application in tgz", func() {
		manager := NewStorageManager(basePath)
		repo := faker.Name().FirstName()
		appName := faker.App().Name()
		version := "latest"
		files := []*entities.FileInfo{
			{Path: "app_config.yaml", Data: []byte("appconf")},
			{Path: "component1.yaml", Data: []byte("component1")},
			{Path: "component2.yaml", Data: []byte("component2")}}
		err := manager.StoreApplication(repo, appName, version, files)
		gomega.Expect(err).Should(gomega.Succeed())

		entity, err := manager.(*storageManager).loadAppFileTgz(appName, fmt.Sprintf("%s/%s/%s/%s", basePath, repo, appName, version))
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(entity).ShouldNot(gomega.BeNil())

		// save file
		file, err := os.Create(fmt.Sprintf("%s/%s.tgz", basePath, appName))
		gomega.Expect(err).Should(gomega.Succeed())

		defer file.Close()

		_, err = file.Write(entity[0].Data)
		gomega.Expect(err).Should(gomega.Succeed())

	})
})
