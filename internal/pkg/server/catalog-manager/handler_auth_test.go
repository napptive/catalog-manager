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
	"context"
	"fmt"
	"io"

	"github.com/golang/mock/gomock"
	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/grpc-catalog-common-go"
	"github.com/napptive/grpc-catalog-go"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var (
	validUsername             = "username"
	validAccountName          = "accountName"
	unauthorizedApplicationID = "unauthorized/test:latest"
)

func GetTestMemberContext() context.Context {
	return utils.CreateTestJWTAuthIncomingContext(validUsername, validAccountName, false, "authorization", "jwt")
}
func GetTestAdminContext() context.Context {
	return utils.CreateTestJWTAuthIncomingContext(validUsername, validAccountName, true, "authorization", "jwt")
}
func GetTestMemberApplicationId() string {
	return fmt.Sprintf("%s/test:latest", validAccountName)
}
func GetTestAccountApplicationId() string {
	return fmt.Sprintf("%s/test:latest", validAccountName)
}

func MockApplicationUpload(addServerStream *MockCatalog_AddServer, applicationID string, ctx context.Context) {
	request := &grpc_catalog_go.AddApplicationRequest{
		ApplicationId: applicationID,
		File:          &grpc_catalog_go.FileInfo{},
	}
	addServerStream.EXPECT().Recv().Return(request, nil)
	addServerStream.EXPECT().Context().Return(ctx).AnyTimes()
	addServerStream.EXPECT().Recv().Return(nil, io.EOF)
	addServerStream.EXPECT().SendAndClose(gomock.Any()).Return(nil)
}

func MockApplicationUploadAuthorizationFailed(addServerStream *MockCatalog_AddServer, applicationID string, ctx context.Context) {
	request := &grpc_catalog_go.AddApplicationRequest{
		ApplicationId: applicationID,
		File:          &grpc_catalog_go.FileInfo{},
	}
	addServerStream.EXPECT().Recv().Return(request, nil)
	addServerStream.EXPECT().Context().Return(ctx).AnyTimes()
}

var _ = ginkgo.Describe("Catalog handler test with auth enabled by JWT", func() {

	var ctrl *gomock.Controller
	var handler *Handler
	var manager *MockManager
	var addServerStream *MockCatalog_AddServer

	var teamConfig = config.TeamConfig{}

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		manager = NewMockManager(ctrl)
		addServerStream = NewMockCatalog_AddServer(ctrl)
		handler = NewHandler(manager, true, teamConfig)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("user can create/remove/update applications on its catalog username", func() {
		ginkgo.It("should allow the user to create/update an application in his namespace", func() {
			appID := GetTestMemberApplicationId()
			MockApplicationUpload(addServerStream, appID, GetTestMemberContext())
			manager.EXPECT().Add(appID, gomock.Any(), false, validAccountName).Return(false, nil)
			err := handler.Add(addServerStream)
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should allow the user to delete an application from his namespace", func() {
			appID := GetTestMemberApplicationId()
			request := &grpc_catalog_go.RemoveApplicationRequest{
				ApplicationId: appID,
			}
			manager.EXPECT().Remove(appID).Return(nil)
			opResponse, err := handler.Remove(GetTestAdminContext(), request)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(opResponse.Status).Should(gomega.Equal(grpc_catalog_common_go.OpStatus_SUCCESS))
		})
		ginkgo.It("should fail if the user creates/updates an application in another namespace", func() {
			appID := unauthorizedApplicationID
			MockApplicationUploadAuthorizationFailed(addServerStream, appID, GetTestMemberContext())
			err := handler.Add(addServerStream)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should fail if the user deletes an application in another namespace", func() {
			appID := unauthorizedApplicationID
			request := &grpc_catalog_go.RemoveApplicationRequest{
				ApplicationId: appID,
			}
			_, err := handler.Remove(GetTestMemberContext(), request)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Context("members & admins can upload applications to the account name", func() {
		ginkgo.It("should allow the user to create/update an application in his account name being a member", func() {
			appID := GetTestMemberApplicationId()
			MockApplicationUpload(addServerStream, appID, GetTestMemberContext())
			manager.EXPECT().Add(appID, gomock.Any(), false, validAccountName).Return(false, nil)
			err := handler.Add(addServerStream)
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should allow the user to create/update an application in his account name being an admin", func() {
			appID := GetTestMemberApplicationId()
			MockApplicationUpload(addServerStream, appID, GetTestAdminContext())
			manager.EXPECT().Add(appID, gomock.Any(), false, validAccountName).Return(false, nil)
			err := handler.Add(addServerStream)
			gomega.Expect(err).To(gomega.Succeed())
		})
		ginkgo.It("should fail if the user deletes an application in his account name being a member", func() {
			appID := GetTestAccountApplicationId()
			request := &grpc_catalog_go.RemoveApplicationRequest{
				ApplicationId: appID,
			}
			_, err := handler.Remove(GetTestMemberContext(), request)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should allow the user to delete an application in his account name being an admin", func() {
			appID := GetTestAccountApplicationId()
			request := &grpc_catalog_go.RemoveApplicationRequest{
				ApplicationId: appID,
			}
			manager.EXPECT().Remove(appID).Return(nil)
			opResponse, err := handler.Remove(GetTestAdminContext(), request)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(opResponse.Status).Should(gomega.Equal(grpc_catalog_common_go.OpStatus_SUCCESS))
		})
		ginkgo.It("should fail if the user creates/updates an application in another account name being a member", func() {
			appID := unauthorizedApplicationID
			MockApplicationUploadAuthorizationFailed(addServerStream, appID, GetTestMemberContext())
			err := handler.Add(addServerStream)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should fail if the user creates/updates an application in another account name being an admin", func() {
			appID := unauthorizedApplicationID
			MockApplicationUploadAuthorizationFailed(addServerStream, appID, GetTestAdminContext())
			err := handler.Add(addServerStream)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})

	})
	ginkgo.Context("admins can delete applications from the account name", func() {
		ginkgo.It("should fail if the user deletes an application in another account name being a member", func() {
			appID := unauthorizedApplicationID
			request := &grpc_catalog_go.RemoveApplicationRequest{
				ApplicationId: appID,
			}
			_, err := handler.Remove(GetTestMemberContext(), request)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should fail if the user deletes an application in another account name being an admin", func() {
			appID := unauthorizedApplicationID
			request := &grpc_catalog_go.RemoveApplicationRequest{
				ApplicationId: appID,
			}
			_, err := handler.Remove(GetTestAdminContext(), request)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})

})
