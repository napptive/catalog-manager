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
	"github.com/napptive/mockup-generator/pkg/mockups"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var metadataFile = `
apiVersion: core.napptive.com/v1alpha1
kind: ApplicationMetadata
# Name of the application, not necessarily a valid k8s name.
name: "My App Name 2"
version: 1.0
description: Short description for searchs. Long one plus how to goes into the README.md
# Keywords facilitate searches on the catalog
keywords:
  - "key1"
  - "key2"
  - "key3"
license: "Apache License Version 2.0"
url: "https://..."
doc: "https://..."
apiVersion: core.napptive.com/v1alpha1
kind: ApplicationMetadata
# Name of the application, not necessarily a valid k8s name.
name: "My App Name 2"
version: 1.0
description: Short description for searchs. Long one plus how to goes into the README.md
# Keywords facilitate searches on the catalog
keywords:
  - "key1"
  - "key2"
  - "key3"
license: "Apache License Version 2.0"
url: "https://..."
doc: "https://..."
# Requires gives a list of entities that are needed to launch the application.
requires:
  traits:
    - my.custom.trait
    - my.custom.trait2
  scopes:
    - my.custom.scope
`

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
			namespace := "namespace"
			appName := "appName"

			filesReturned := []*entities.FileInfo{
				{
					Path: "./app.yaml",
					Data: []byte("app"),
				}, {
					Path: "./metadata.yaml",
					Data: []byte("metadata"),
				}}

			storageProvider.EXPECT().GetApplication(namespace, appName, "latest", false).Return(filesReturned, nil)

			manager := NewManager(storageProvider, metadataProvider, "")
			files, err := manager.Download(fmt.Sprintf("%s/%s", namespace, appName), false)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(files).ShouldNot(gomega.BeEmpty())
			gomega.Expect(files).ShouldNot(gomega.BeNil())
		})
		ginkgo.It("should not be able to download an application with a wrong name", func() {
			appName := "appName"

			manager := NewManager(storageProvider, metadataProvider, "")
			_, err := manager.Download(appName, false)
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})
		ginkgo.It("should not be able to download an application if there is an error in the storage", func() {
			namespace := "namespace"
			appName := "appName"

			storageProvider.EXPECT().GetApplication(namespace, appName, "latest", false).Return(nil, nerrors.NewInternalError("error reading repository"))

			manager := NewManager(storageProvider, metadataProvider, "")
			_, err := manager.Download(fmt.Sprintf("%s/%s", namespace, appName), false)
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})
	})

	ginkgo.Context("Getting application", func() {
		ginkgo.It("should be able to get an application", func() {
			namespace := "namespace"
			appName := "appName"

			matcher := mockups.NewStructMatcher(map[string]interface{}{"Namespace": namespace, "ApplicationName": appName})
			metadataProvider.EXPECT().Get(matcher).Return(&entities.ApplicationInfo{
				CatalogID:       "",
				Namespace:       namespace,
				ApplicationName: appName,
				Tag:             "latest",
				MetadataName:    "My App",
				Metadata:        metadataFile,
			}, nil)

			manager := NewManager(storageProvider, metadataProvider, "")
			metadata, err := manager.Get(fmt.Sprintf("%s/%s", namespace, appName))
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(metadata).ShouldNot(gomega.BeNil())
			gomega.Expect(metadata.MetadataObj.Name).ShouldNot(gomega.BeEmpty())

		})
		ginkgo.It("should not be able to return an application if it does not exist", func() {
			namespace := "namespace"
			appName := "appName"

			matcher := mockups.NewStructMatcher(map[string]interface{}{"Namespace": namespace, "ApplicationName": appName})
			metadataProvider.EXPECT().Get(matcher).Return(nil, nerrors.NewNotFoundError("not found"))

			manager := NewManager(storageProvider, metadataProvider, "")
			_, err := manager.Get(fmt.Sprintf("%s/%s", namespace, appName))
			gomega.Expect(err).ShouldNot(gomega.Succeed())

		})
		ginkgo.It("should not be able to return a invalid application", func() {
			manager := NewManager(storageProvider, metadataProvider, "")
			_, err := manager.Get("invalidApp")
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})
	})

	ginkgo.Context("Listing applications", func() {
		ginkgo.It("should be able to list applications", func() {

			returned := []*entities.AppSummary{
				{
					Namespace:       "ns1",
					ApplicationName: "app1",
					TagMetadataName: map[string]string{"tag1": "my app v1"},
				},
				{
					Namespace:       "ns1",
					ApplicationName: "app1",
					TagMetadataName: map[string]string{"tag2": "my app v2"},
				},
			}

			metadataProvider.EXPECT().ListSummary(gomock.Any()).Return(returned, nil)

			manager := NewManager(storageProvider, metadataProvider, "")
			received, err := manager.List("")
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(received).ShouldNot(gomega.BeEmpty())
			gomega.Expect(len(received)).ShouldNot(gomega.BeZero())
		})
		ginkgo.It("should be able to list applications from a selected namespace", func() {
			returned := []*entities.AppSummary{
				{
					Namespace:       "ns1",
					ApplicationName: "app1",
					TagMetadataName: map[string]string{"tag1": "my app v1"},
				},
				{
					Namespace:       "ns1",
					ApplicationName: "app1",
					TagMetadataName: map[string]string{"tag2": "my app v2"},
				},
			}
			metadataProvider.EXPECT().ListSummary(gomock.Any()).Return(returned, nil)
			manager := NewManager(storageProvider, metadataProvider, "")
			received, err := manager.List("ns1")
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(received).ShouldNot(gomega.BeEmpty())
			gomega.Expect(len(received)).Should(gomega.Equal(len(returned)))

		})
		ginkgo.It("should be able to return an empty list of applications", func() {

			returned := make([]*entities.AppSummary, 0)

			metadataProvider.EXPECT().ListSummary(gomock.Any()).Return(returned, nil)

			manager := NewManager(storageProvider, metadataProvider, "")
			received, err := manager.List("")
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(received).Should(gomega.BeEmpty())
		})
	})
})
