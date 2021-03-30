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

package catalog_manager

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Catalog handler test", func() {

	var ctrl *gomock.Controller
	var storageProvider *MockStorageManager
	var metadataProvider *MockMetadataProvider

	ginkgo.BeforeSuite(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		storageProvider = NewMockStorageManager(ctrl)
		metadataProvider = NewMockMetadataProvider(ctrl)
	})

	ginkgo.AfterSuite(func() {
		ctrl.Finish()
	})

	ginkgo.Context("Downloading applications", func() {
		ginkgo.It("Should be able to download an application", func() {
			repoName := "repoName"
			appName := "appName"

			filesReturned := []*entities.FileInfo{
				&entities.FileInfo{
					Path: "./app.yaml",
					Data: []byte("app"),
				}, &entities.FileInfo{
					Path: "./metadata.yaml",
					Data: []byte("metadata"),
				}}

			storageProvider.EXPECT().GetApplication(repoName, appName, "latest").Return(filesReturned, nil)

			manager := NewManager(storageProvider, metadataProvider)
			files, err := manager.Download(fmt.Sprintf("%s/%s", repoName, appName))
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(files).ShouldNot(gomega.BeEmpty())
			gomega.Expect(files).ShouldNot(gomega.BeNil())
		})
		ginkgo.It("should not be able to download an application with a wrong name", func (){
			appName := "appName"

			manager := NewManager(storageProvider, metadataProvider)
			_, err := manager.Download(appName)
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})
		ginkgo.It("should not be able to download an application if there is an error in the storage", func (){
			repoName := "repoName"
			appName := "appName"

			storageProvider.EXPECT().GetApplication(repoName, appName, "latest").Return(nil, nerrors.NewInternalError("error reading repository"))

			manager := NewManager(storageProvider, metadataProvider)
			_, err := manager.Download(fmt.Sprintf("%s/%s", repoName, appName))
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})
	})
})
