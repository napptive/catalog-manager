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

package catalog_manager

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"io"

	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/server/resolver"
	"github.com/napptive/grpc-catalog-common-go"
	"github.com/napptive/grpc-catalog-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/napptive/njwt/pkg/interceptors"
	"github.com/rs/zerolog/log"
)

const appRemovedMsg = "%s removed from catalog"

type Handler struct {
	teamConfig config.TeamConfig
	manager    Manager
	// authEnabled is a boolean to check the user
	authEnabled bool
	resolver    resolver.PermissionResolver
}

// TODO: Check update/get concurrency

func NewHandler(manager Manager, authEnabled bool, teamConfig config.TeamConfig, resolver resolver.PermissionResolver) *Handler {
	return &Handler{manager: manager, authEnabled: authEnabled, teamConfig: teamConfig, resolver: resolver}
}

// Add a new application in the catalog
func (h *Handler) Add(server grpc_catalog_go.Catalog_AddServer) error {

	accountName := ""
	// if authentication is enabled -> Get the account name to filter all private apps by namespace
	if h.authEnabled {
		accountNameFromCtx, err := h.getAccountNameFromContext(server.Context())
		if err != nil {
			log.Error().Err(err).Msg("error getting account name from context")
			return err
		}
		accountName = *accountNameFromCtx
	}

	// TODO: create a map to load the files and avoid send a file twice
	applicationID := ""
	var applicationFiles []*entities.FileInfo
	private := false

	for {
		// From https://grpc.io/docs/languages/go/basics/#server-side-streaming-rpc-1
		request, err := server.Recv()
		if err == io.EOF {
			isPrivate, err := h.manager.Add(applicationID, applicationFiles, private, accountName)
			if err != nil {
				return nerrors.FromError(err).ToGRPC()
			} else {
				var message string
				if isPrivate {
					message = fmt.Sprintf("Private application %s added.", applicationID)
				} else {
					message = fmt.Sprintf("Public application %s added.", applicationID)
				}

				return server.SendAndClose(&grpc_catalog_common_go.OpResponse{
					Status:     grpc_catalog_common_go.OpStatus_SUCCESS,
					StatusName: grpc_catalog_common_go.OpStatus_SUCCESS.String(),
					UserInfo:   message,
				})
			}
		}
		if err != nil {
			return nerrors.FromError(err).ToGRPC()
		}

		// the first time save the application identifier
		if applicationID == "" {
			applicationID = request.ApplicationId
			private = request.Private
			// Also, validate the user.
			if vErr := h.validateUser(server.Context(), request.ApplicationId, "push", false); vErr != nil {
				return vErr
			}

			// if the application is private and the authentication is no enable -> error
			if private && !h.authEnabled {
				sErr := nerrors.NewFailedPreconditionError("enable authentication to make use of private apps")
				return nerrors.FromError(sErr).ToGRPC()
			}
		}

		// if the identifier is other than the saved one -> ERROR
		// For now, we do not allow the user to send different applications in the same stream
		if request.ApplicationId != applicationID {
			sErr := nerrors.NewFailedPreconditionError("cannot send different applications in the same stream")
			return nerrors.FromError(sErr).ToGRPC()
		}
		if request.Private != private {
			sErr := nerrors.NewFailedPreconditionError("cannot send different private flag in the same stream")
			return nerrors.FromError(sErr).ToGRPC()
		}
		// Append the files
		fileInfo := entities.NewFileInfo(request.File)
		if fileInfo != nil {
			applicationFiles = append(applicationFiles, fileInfo)
		}
	}
}

func (h *Handler) Upload(ctx context.Context, request *grpc_catalog_go.UploadApplicationRequest) (*grpc_catalog_common_go.OpResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	// check the user (check if after validation yto be sure the ApplicationId is filled
	if err := h.validateUser(ctx, request.ApplicationId, "remove", true); err != nil {
		return nil, err
	}

	accountName := ""
	// if authentication is enabled -> Get the account name to filter all private apps by namespace
	if h.authEnabled {
		accountNameFromCtx, err := h.getAccountNameFromContext(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error getting username from context")
			return nil, err
		}
		accountName = *accountNameFromCtx
	}

	files := make([]*entities.FileInfo, 0)
	for _, file := range request.Files {
		sDec, err := b64.StdEncoding.DecodeString(file.Data)
		if err != nil {
			log.Error().Err(err).Str("file", file.Path).Msg("error uploading catalog application. Error decoding application file")
			return nil, nerrors.NewInternalErrorFrom(err, "Error uploading catalog application, error decoding application file [%s]", file.Path)
		}
		files = append(files, &entities.FileInfo{
			Path: file.Path,
			Data: sDec,
		})
	}
	isPrivate, err := h.manager.Add(request.ApplicationId, files, request.Private, accountName)
	if err != nil {
		log.Error().Err(err).Str("applicationID", request.ApplicationId).Msg("error uploading application")
		return nil, nerrors.FromGRPC(err)
	}

	message := ""

	if isPrivate {
		message = fmt.Sprintf("Private application %s added.", request.ApplicationId)
	} else {
		message = fmt.Sprintf("Public application %s added.", request.ApplicationId)
	}

	return &grpc_catalog_common_go.OpResponse{
		Status:     grpc_catalog_common_go.OpStatus_SUCCESS,
		StatusName: grpc_catalog_common_go.OpStatus_SUCCESS.String(),
		UserInfo:   message,
	}, nil
}

// Download an application from catalog
func (h *Handler) Download(request *grpc_catalog_go.DownloadApplicationRequest, server grpc_catalog_go.Catalog_DownloadServer) error {
	// validate
	if err := request.Validate(); err != nil {
		return nerrors.FromError(err).ToGRPC()
	}

	// check user permission in the application namespace (for private apps)
	accountAllowed, err := h.resolver.CheckAccountPermissions(server.Context(), request.ApplicationId, false)
	if err != nil {
		log.Error().Err(err).Str("application_name", request.ApplicationId).Msg("error checking permission, unable to download the application")
		return nerrors.FromError(err).ToGRPC()
	}
	// download the application
	files, err := h.manager.Download(request.ApplicationId, request.Compressed, *accountAllowed)
	if err != nil {
		log.Error().Err(err).Str("application_name", request.ApplicationId).Msg("error downloading the application")
		return nerrors.FromError(err).ToGRPC()
	}
	// send the files
	for _, file := range files {
		if err := server.Send(file.ToGRPC()); err != nil {
			return nerrors.NewInternalErrorFrom(err, "unable to send the file").ToGRPC()
		}
	}
	return nil
}

// Remove an application from the catalog
func (h *Handler) Remove(ctx context.Context, request *grpc_catalog_go.RemoveApplicationRequest) (*grpc_catalog_common_go.OpResponse, error) {

	// validate
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	// check the user (check if after validation yto be sure the ApplicationId is filled
	if err := h.validateUser(ctx, request.ApplicationId, "remove", true); err != nil {
		log.Error().Err(err).Str("application_name", request.ApplicationId).Msg("error validating user, unable to remove the application")
		return nil, nerrors.FromError(err).ToGRPC()
	}

	// remove the application
	if err := h.manager.Remove(request.ApplicationId); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	// return response
	return &grpc_catalog_common_go.OpResponse{
		Status:     grpc_catalog_common_go.OpStatus_SUCCESS,
		StatusName: grpc_catalog_common_go.OpStatus_SUCCESS.String(),
		UserInfo:   fmt.Sprintf(appRemovedMsg, request.ApplicationId),
	}, nil
}

// List returns a list with all the required applications
func (h *Handler) List(ctx context.Context, request *grpc_catalog_go.ListApplicationsRequest) (*grpc_catalog_go.ApplicationList, error) {
	namespacesMap := make(map[string]*bool)

	// Boolean to indicate if list all the public applications
	var showPublicApps bool

	if request.Namespace != "" {
		showPublicApps = false
		// check user permission in the application namespace (for private apps)
		accountAllowed, err := h.resolver.CheckAccountPermissions(ctx, fmt.Sprintf("%s/dummy", request.Namespace), false)
		if err != nil {
			log.Error().Err(err).Str("namespace", request.Namespace).Msg("error checking permission, unable to list applications")
			return nil, nerrors.FromError(err).ToGRPC()
		}
		if *accountAllowed {
			accountAllowed = nil
		}
		namespacesMap[request.Namespace] = accountAllowed
	} else {
		ownApps := true
		showPublicApps = true

		// getClaim
		claim, err := interceptors.GetClaimFromContext(ctx)
		if err != nil {
			return nil, err
		}
		for _, account := range claim.Accounts {
			namespacesMap[account.Name] = &ownApps
		}
	}

	list, err := h.manager.List(namespacesMap, showPublicApps)
	if err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}
	summaryList := make([]*grpc_catalog_go.ApplicationSummary, 0)
	for _, app := range list {
		summaryList = append(summaryList, app.ToApplicationSummary())
	}

	return &grpc_catalog_go.ApplicationList{Applications: summaryList}, nil

}

// Info returns the detail of a given application
func (h *Handler) Info(ctx context.Context, request *grpc_catalog_go.InfoApplicationRequest) (*grpc_catalog_go.InfoApplicationResponse, error) {

	// validate
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	// check user permission in the application namespace (for private apps)
	accountAllowed, err := h.resolver.CheckAccountPermissions(ctx, request.ApplicationId, false)
	if err != nil {
		log.Error().Err(err).Str("application_name", request.ApplicationId).Msg("error checking permission, get application info")
		return nil, nerrors.FromError(err).ToGRPC()
	}

	// get the application info
	retrieved, err := h.manager.Get(request.ApplicationId, *accountAllowed)
	if err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	// return the response
	return &grpc_catalog_go.InfoApplicationResponse{
		Namespace:       retrieved.Namespace,
		ApplicationName: retrieved.ApplicationName,
		Tag:             retrieved.Tag,
		MetadataFile:    []byte(retrieved.Metadata),
		ReadmeFile:      []byte(retrieved.Readme),
		Metadata:        retrieved.MetadataObj.ToGRPC(),
		Private:         retrieved.Private,
	}, nil
}

// Summary returns the summary of the catalog (#repositories, #applications and #tags)
func (h *Handler) Summary(_ context.Context, _ *grpc_catalog_common_go.EmptyRequest) (*grpc_catalog_go.SummaryResponse, error) {

	summary, err := h.manager.Summary()
	if err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}
	return summary.ToSummaryResponse(), nil
}

func (h *Handler) Update(ctx context.Context, request *grpc_catalog_go.UpdateRequest) (*grpc_catalog_common_go.OpResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}
	if !h.authEnabled {
		sErr := nerrors.NewFailedPreconditionError("enable authentication to make use of private apps")
		return nil, nerrors.FromError(sErr).ToGRPC()
	}
	// check the user (check if after validation yto be sure the ApplicationId is filled
	appName := fmt.Sprintf("%s/%s", request.Namespace, request.ApplicationName)
	if err := h.validateUser(ctx, appName, "change visibility", true); err != nil {
		log.Error().Err(err).Str("application_name", appName).Msg("error validating user, unable to change the application visibility")
		return nil, nerrors.FromError(err).ToGRPC()
	}

	if err := h.manager.UpdateApplicationVisibility(request.Namespace, request.ApplicationName, request.Private); err != nil {
		log.Error().Err(err).Str("namespace", request.Namespace).
			Str("application", request.ApplicationName).Msg("error changing application visibility")
		return nil, nerrors.FromError(err).ToGRPC()
	}

	privateStr := "public"
	if request.Private {
		privateStr = "private"
	}

	return &grpc_catalog_common_go.OpResponse{
		Status:     grpc_catalog_common_go.OpStatus_SUCCESS,
		StatusName: grpc_catalog_common_go.OpStatus_SUCCESS.String(),
		UserInfo:   fmt.Sprintf("Application %s/%s changed to %s", request.Namespace, request.ApplicationName, privateStr),
	}, nil
}

// validateUser check if the user in the context is the same as the repo name
func (h *Handler) validateUser(ctx context.Context, appName string, action string, requireAdminPrivilege bool) error {
	allowed, err := h.resolver.CheckAccountPermissions(ctx, appName, requireAdminPrivilege)
	if err != nil {
		return err
	}
	if *allowed {
		return nil
	}
	// TODO: change the error message
	return nerrors.NewPermissionDeniedError("operation not allowed")
}

// isPrivilegedUser checks the user role
func (h *Handler) isPrivilegedUser(userName string) bool {
	if !h.teamConfig.Enabled {
		return false
	}
	for _, user := range h.teamConfig.PrivilegedUsers {
		if user == userName {
			return true
		}
	}

	return false
}

// isTeamNamespace checks the namespace role
func (h *Handler) isTeamNamespace(repoName string) bool {
	if !h.teamConfig.Enabled {
		return false
	}
	for _, repo := range h.teamConfig.TeamNamespaces {
		if repo == repoName {
			return true
		}
	}

	return false
}

// getAccountNameFromContext returns the username from the token
func (h *Handler) getAccountNameFromContext(ctx context.Context) (*string, error) {
	claim, err := interceptors.GetClaimFromContext(ctx)
	if err != nil {
		return nil, err
	}

	log.Debug().Interface("accounts", claim).Msg("borrar")

	return claim.GetCurrentAccountName()
}
