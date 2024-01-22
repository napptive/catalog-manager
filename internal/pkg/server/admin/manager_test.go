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

package admin

import (
	"github.com/golang/mock/gomock"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/mock-extensions/pkg/matcher"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func createTestApps(namespace string, num int) []*entities.ApplicationInfo {
	result := make([]*entities.ApplicationInfo, 0)
	for appIndex := 0; appIndex < num; appIndex++ {
		toAdd := utils.CreateTestApplicationInfo()
		toAdd.Namespace = namespace
		result = append(result, toAdd)
	}
	return result
}

var _ = ginkgo.Describe("Catalog handler test", func() {

	var ctrl *gomock.Controller
	var storageProvider *MockStorageManager
	var metadataProvider *MockMetadataProvider
	var manager Manager

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		storageProvider = NewMockStorageManager(ctrl)
		metadataProvider = NewMockMetadataProvider(ctrl)
		manager = NewManager(storageProvider, metadataProvider)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.It("should be able to delete an existing user account", func() {
		namespace := "valid"
		numApps := 10
		apps := createTestApps(namespace, numApps)
		metadataProvider.EXPECT().List(namespace).Return(apps, nil)
		namespaceMatcher := matcher.NewStructMatcher(map[string]interface{}{"Namespace": namespace})
		metadataProvider.EXPECT().Remove(namespaceMatcher).Times(numApps).Return(nil)
		storageProvider.EXPECT().RemoveRepository(namespace).Return(nil)
		err := manager.DeleteNamespace(namespace)
		gomega.Expect(err).Should(gomega.Succeed())
	})
	ginkgo.It("should not return error when trying to delete a non existing user account", func() {
		namespace := "valid_not_existing"
		numApps := 0
		apps := createTestApps(namespace, numApps)
		metadataProvider.EXPECT().List(namespace).Return(apps, nil)
		err := manager.DeleteNamespace(namespace)
		gomega.Expect(err).Should(gomega.Succeed())
	})

})
