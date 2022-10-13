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

package catalog_manager

import (
	"context"
	"fmt"
	"io"

	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	grpc_catalog_common_go "github.com/napptive/grpc-catalog-common-go"
	grpc_catalog_go "github.com/napptive/grpc-catalog-go"
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
}

// TODO: Check update/get concurrency

func NewHandler(manager Manager, authEnabled bool, teamConfig config.TeamConfig) *Handler {
	return &Handler{manager: manager, authEnabled: authEnabled, teamConfig: teamConfig}
}

// Add a new application in the catalog
func (h *Handler) Add(server grpc_catalog_go.Catalog_AddServer) error {

	username := ""
	// if authentication is enabled -> Get the account name to filter all private apps by namespace
	if h.authEnabled {
		usernameFromCtx, err := h.getUsernameFromContext(server.Context())
		if err != nil {
			log.Error().Err(err).Msg("error getting username from context")
			return err
		}
		username = *usernameFromCtx
	}

	// TODO: create a map to load the files and avoid send a file twice
	applicationID := ""
	var applicationFiles []*entities.FileInfo
	private := false

	for {
		// From https://grpc.io/docs/languages/go/basics/#server-side-streaming-rpc-1
		request, err := server.Recv()
		if err == io.EOF {
			isPrivate, err := h.manager.Add(applicationID, applicationFiles, private, username)
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
		applicationFiles = append(applicationFiles, entities.NewFileInfo(request.File))
	}
}

// Download an application from catalog
func (h *Handler) Download(request *grpc_catalog_go.DownloadApplicationRequest, server grpc_catalog_go.Catalog_DownloadServer) error {
	if err := request.Validate(); err != nil {
		return nerrors.FromError(err).ToGRPC()
	}

	username := ""
	// if authentication is enabled -> Get the account name to filter all private apps by namespace
	if h.authEnabled {
		usernameFromCtx, err := h.getUsernameFromContext(server.Context())
		if err != nil {
			log.Error().Err(err).Msg("error getting username from context")
			return err
		}
		username = *usernameFromCtx
	}

	files, err := h.manager.Download(request.ApplicationId, request.Compressed, username)
	if err != nil {
		return nerrors.FromError(err).ToGRPC()
	}

	for _, file := range files {
		if err := server.Send(file.ToGRPC()); err != nil {
			return nerrors.NewInternalErrorFrom(err, "unable to send the file").ToGRPC()
		}
	}

	return nil
}

// Remove an application from the catalog
func (h *Handler) Remove(ctx context.Context, request *grpc_catalog_go.RemoveApplicationRequest) (*grpc_catalog_common_go.OpResponse, error) {

	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	// check the user (check if after validation yto be sure the ApplicationId is filled
	if err := h.validateUser(ctx, request.ApplicationId, "remove", true); err != nil {
		return nil, err
	}

	if err := h.manager.Remove(request.ApplicationId); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	return &grpc_catalog_common_go.OpResponse{
		Status:     grpc_catalog_common_go.OpStatus_SUCCESS,
		StatusName: grpc_catalog_common_go.OpStatus_SUCCESS.String(),
		UserInfo:   fmt.Sprintf(appRemovedMsg, request.ApplicationId),
	}, nil
}

// List returns a list with all the applications
func (h *Handler) List(ctx context.Context, request *grpc_catalog_go.ListApplicationsRequest) (*grpc_catalog_go.ApplicationList, error) {

	username := ""
	// if authentication is enabled -> Get the account name to filter all private apps by namespace
	if h.authEnabled {
		usernameFromCtx, err := h.getUsernameFromContext(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error getting username from context")
			return nil, err
		}
		username = *usernameFromCtx
	}

	returned, err := h.manager.List(request.Namespace, username)
	if err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	summaryList := make([]*grpc_catalog_go.ApplicationSummary, 0)
	for _, app := range returned {
		summaryList = append(summaryList, app.ToApplicationSummary())
	}

	return &grpc_catalog_go.ApplicationList{Applications: summaryList}, nil

}

// Info returns the detail of a given application
func (h *Handler) Info(ctx context.Context, request *grpc_catalog_go.InfoApplicationRequest) (*grpc_catalog_go.InfoApplicationResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}

	username := ""
	// if authentication is enabled -> Get the account name to filter all private apps by namespace
	if h.authEnabled {
		usernameFromCtx, err := h.getUsernameFromContext(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error getting username from context")
			return nil, err
		}
		username = *usernameFromCtx
	}

	retrieved, err := h.manager.Get(request.ApplicationId, username)
	if err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}
	return &grpc_catalog_go.InfoApplicationResponse{
		Namespace:       retrieved.Namespace,
		ApplicationName: retrieved.ApplicationName,
		Tag:             retrieved.Tag,
		MetadataFile:    []byte(retrieved.Metadata),
		ReadmeFile:      []byte(retrieved.Readme),
		Metadata:        retrieved.MetadataObj.ToGRPC(),
	}, nil
}

// Summary returns the summary of the catalog (#repositories, #applications and #tags)
func (h *Handler) Summary(ctx context.Context, request *grpc_catalog_common_go.EmptyRequest) (*grpc_catalog_go.SummaryResponse, error) {

	summary, err := h.manager.Summary()
	if err != nil {
		return nil, nerrors.FromError(err).ToGRPC()
	}
	return summary.ToSummaryResponse(), nil
}

// validateUser check if the user in the context is the same as the repo name
func (h *Handler) validateUser(ctx context.Context, appName string, action string, requireAdminPrivilege bool) error {

	// check the user (check if after validation to be sure the ApplicationName is filled
	if h.authEnabled {
		claim, err := interceptors.GetClaimFromContext(ctx)
		if err != nil {
			return err
		}
		log.Debug().Interface("user", claim).Msg("validating user")

		// get the repoName
		_, appID, err := utils.DecomposeApplicationID(appName)
		if err != nil {
			return err
		}

		// A user can only remove their apps
		if appID.Namespace != claim.Username {

			// Check target account
			if appID.Namespace == claim.AccountName {
				if requireAdminPrivilege {
					if claim.AccountAdmin {
						return nil
					} else {
						return nerrors.NewPermissionDeniedError("%s operation requires ADMIN privileges", action)
					}
				}
				return nil
			}

			// if the user is privileged and the repository is a team repository -> OK
			isPrivileged := h.isPrivilegedUser(claim.Username)
			isTeamRepo := h.isTeamNamespace(appID.Namespace)
			log.Debug().Str("repository", appID.Namespace).Str("user", claim.Username).
				Bool("isPrivileged", isPrivileged).Bool("isTeamRepo", isTeamRepo).Msg("checking privileges")
			if !h.isPrivilegedUser(claim.Username) || !h.isTeamNamespace(appID.Namespace) {
				return nerrors.NewPermissionDeniedError("A user can only %s their apps", action)
			}
		}

	}
	return nil
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

// getUsernameFromContext returns the username from the token
func (h *Handler) getUsernameFromContext(ctx context.Context) (*string, error) {
	claim, err := interceptors.GetClaimFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return &claim.Username, nil
}
