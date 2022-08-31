/*
Copyright 2022 Napptive

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package apps

import (
	"fmt"

	gomock "github.com/golang/mock/gomock"
	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/server/apps/mocks"
	"github.com/napptive/mockup-generator/pkg/mockups"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var application = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-test
data:
  cpu: "0.50"
  memory: "250Mi"
---
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: nginx-app
  annotations:
    version: v1.0.0
    description: "Customized version of nginx"
spec:
  components:
    - name: nginx
      type: webservice
      properties:
        image: nginx:1.20.0
        ports:
        - port: 80
          expose: true
`

const cm = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-test
data:
  cpu: "0.50"
  memory: "250Mi"
`

var _ = ginkgo.Describe("Apps manager test", func() {

	var ctrl *gomock.Controller
	var catalogManager *mocks.MockManager

	var manager Manager

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		catalogManager = mocks.NewMockManager(ctrl)
		manager = NewManager(&config.Config{}, catalogManager)
	})

	ginkgo.Context("Getting application config", func() {
		ginkgo.It("Should be able to get application configuration", func() {

			appID := fmt.Sprintf("%s/%s", mockups.GetUserName(), "application")

			catalogManager.EXPECT().Download(appID, false).Return([]*entities.FileInfo{{
				Path: "/../..",
				Data: []byte(application),
			}}, nil)

			config, err := manager.GetConfiguration(appID)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(config).ShouldNot(gomega.BeNil())
			gomega.Expect(config.IsApplication).Should(gomega.BeTrue())

		})
		ginkgo.It("Should be able to get application configuration, when the files do not containt any applications", func() {

			appID := fmt.Sprintf("%s/%s", mockups.GetUserName(), "application")

			catalogManager.EXPECT().Download(appID, false).Return([]*entities.FileInfo{{
				Path: "/../..",
				Data: []byte(cm),
			}}, nil)

			config, err := manager.GetConfiguration(appID)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(config).ShouldNot(gomega.BeNil())
			gomega.Expect(config.IsApplication).ShouldNot(gomega.BeTrue())
		})

		ginkgo.It("Should not be able to get application configuration if the application does not exists", func() {

			appID := fmt.Sprintf("%s/%s", mockups.GetUserName(), "application")

			catalogManager.EXPECT().Download(appID, false).Return([]*entities.FileInfo{}, nerrors.NewNotFoundError("Application not found"))

			_, err := manager.GetConfiguration(appID)
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})

	})

})
