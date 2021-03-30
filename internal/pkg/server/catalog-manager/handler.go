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
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	grpc_catalog_common_go "github.com/napptive/grpc-catalog-common-go"
	grpc_catalog_go "github.com/napptive/grpc-catalog-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"io"
)


const appAddedMsg = "Application added successfully"

type Handler struct {
	manager Manager
}

// TODO: Check update/get concurrency

func NewHandler(manager Manager) *Handler {
	return &Handler{manager: manager}
}

// Add a new application in the catalog
func (h *Handler) Add(server grpc_catalog_go.Catalog_AddServer) error {

	// TODO: create a map to load the files and avoid send a file twice
	applicationName := ""
	var applicationFiles []*entities.FileInfo

	for {
		// From https://grpc.io/docs/languages/go/basics/#server-side-streaming-rpc-1
		request, err := server.Recv()
		log.Debug().Interface("request", request).Msg("Add")
		if err == io.EOF {
			if err := h.manager.Add(applicationName, applicationFiles); err != nil {
				return nerrors.FromError(err).ToGRPC()
			} else {
				return server.SendAndClose(&grpc_catalog_common_go.OpResponse{
					Status:     grpc_catalog_common_go.OpStatus_SUCCESS,
					StatusName: grpc_catalog_common_go.OpStatus_SUCCESS.String(),
					UserInfo:   appAddedMsg,
				})
			}
		}
		if err != nil {
			return nerrors.FromError(err).ToGRPC()
		}

		// the first time save the application name
		if applicationName == "" {
			applicationName = request.ApplicationName
		}

		// if the name is other than the saved one -> ERROR
		// it is not allowed sending different applications in the same stream
		if request.ApplicationName != applicationName {
			sErr := nerrors.NewFailedPreconditionError("not allowed sending different applications in the same stream")
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

	files, err := h.manager.Download(request)
	if err != nil {
		return nerrors.FromError(err).ToGRPC()
	}
	log.Debug().Interface("files", files).Msg("Files returned")

	for _, file := range files {
		if err := server.Send(file.ToGRPC()); err != nil {
			return nerrors.NewInternalErrorFrom(err, "unable to send the file").ToGRPC()
		}
	}

	return nil
}
