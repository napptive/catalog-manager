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
package cli

import (
	"fmt"
	grpc_catalog_go "github.com/napptive/grpc-catalog-go"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"strings"
	"time"
)

type ApplicationCli struct {
	// adminClient to connect to Admin interface
	adminClient grpc_catalog_go.NamespaceAdministrationClient
}

func NewApplicationCli(adminPort int) (*ApplicationCli, error) {
	dir := fmt.Sprintf(":%d", adminPort)
	log.Info().Str("dir", dir).Msg("admin direction")
	conn, err := grpc.Dial(dir, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := grpc_catalog_go.NewNamespaceAdministrationClient(conn)
	return &ApplicationCli{
		adminClient: client,
	}, nil
}

// Delete removes applications from the catalog
// if app is an applicationName removes it and
// if app is a namespace removes all the namespace applications
func (ac *ApplicationCli) Delete(app string) error {
	log.Debug().Str("appName", app).Msg("Delete")

	if app == "" {
		return nerrors.NewFailedPreconditionError("applicationName or namespace mut be filled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// check if app is an applicationName or is a namespace
	// applicationName -> namespace/appName
	// namespace -> namespace
	if strings.Index(app, "/") != -1 {
		response, err := ac.adminClient.DeleteApplication(ctx, &grpc_catalog_go.RemoveApplicationRequest{ApplicationId: app})
		PrintResultOrError(response, err)
	} else {
		// is a namespace
		response, err := ac.adminClient.Delete(ctx, &grpc_catalog_go.DeleteNamespaceRequest{Namespace: app})
		PrintResultOrError(response, err)
	}

	return nil
}

func (ac *ApplicationCli) List(namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	response, err := ac.adminClient.List(ctx, &grpc_catalog_go.ListApplicationsRequest{Namespace: namespace})
	PrintResultOrError(response, err)

	return nil
}
