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
	"github.com/napptive/catalog-manager/internal/pkg/server/resolver"

	"github.com/golang/mock/gomock"
	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/grpc-catalog-common-go"
	"github.com/napptive/grpc-catalog-go"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	testAppName           = "repo/app:tag"
	testTargetEnvironment = "account/env"
	testTargetPlayground  = "target.playground"
)

var _ = ginkgo.Describe("Apps handler test with auth enabled by JWT", func() {

	var ctrl *gomock.Controller
	var handler *Handler
	var manager *MockManager

	var handlerConfig config.Config

	ginkgo.BeforeEach(func() {
		handlerConfig.AuthEnabled = true
		handlerConfig.Header = "authorization"

		ctrl = gomock.NewController(ginkgo.GinkgoT())
		manager = NewMockManager(ctrl)
		permissionResolver := resolver.NewPermissionResolver(true, config.NewTeamConfig(false, "", ""))
		handler = NewHandler(&handlerConfig.JWTConfig, manager, *permissionResolver)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("when the user does not provide a JWT", func() {
		ginkgo.It("should not be possible to deploy an application", func() {
			ctx := utils.CreateTestJWTAuthIncomingContext("user", "account", true, "authorization", "")
			_, err := handler.Deploy(ctx, &grpc_catalog_go.DeployApplicationRequest{
				ApplicationId:                  testAppName,
				TargetEnvironmentQualifiedName: testTargetEnvironment,
				TargetPlaygroundApiUrl:         testTargetPlayground,
			})
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Context("with a JWT", func() {
		ginkgo.It("should be able to deploy apps", func() {
			ctx := utils.CreateTestJWTAuthIncomingContext("user", "account", true, "authorization", "jwt")
			manager.EXPECT().Deploy(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&grpc_catalog_common_go.OpResponse{}, nil)
			response, err := handler.Deploy(ctx, &grpc_catalog_go.DeployApplicationRequest{
				ApplicationId:                  testAppName,
				TargetEnvironmentQualifiedName: testTargetEnvironment,
				TargetPlaygroundApiUrl:         testTargetPlayground,
			})
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(response).ShouldNot(gomega.BeNil())
		})
	})

	ginkgo.Context("getting application configuration", func() {
		ginkgo.It("Should be able to return application configuration", func() {
			appID := fmt.Sprintf("%s/%s", "username", "application")
			ctx := utils.CreateTestJWTAuthIncomingContext("user", "account", true, "authorization", "jwt")

			manager.EXPECT().GetConfiguration(appID, gomock.Any()).Return(&grpc_catalog_go.GetConfigurationResponse{
				IsApplication:          true,
				ApplicationDefaultName: "name",
				SpecComponentsRaw:      "",
			}, nil)

			conf, err := handler.GetConfiguration(ctx, &grpc_catalog_go.GetConfigurationRequest{
				ApplicationId: appID,
			})
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(conf).ShouldNot(gomega.BeNil())
		})
		ginkgo.It("Should be able to return application configuration when the catalog application is not an oam application", func() {
			appID := fmt.Sprintf("%s/%s", "username", "application")
			ctx := utils.CreateTestJWTAuthIncomingContext("user", "account", true, "authorization", "jwt")

			manager.EXPECT().GetConfiguration(appID, gomock.Any()).Return(&grpc_catalog_go.GetConfigurationResponse{
				IsApplication:          false,
				ApplicationDefaultName: "",
				SpecComponentsRaw:      "",
			}, nil)

			conf, err := handler.GetConfiguration(ctx, &grpc_catalog_go.GetConfigurationRequest{
				ApplicationId: appID,
			})
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(conf).ShouldNot(gomega.BeNil())
		})
	})

})
