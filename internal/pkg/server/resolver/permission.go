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

package resolver

import (
	"context"

	"github.com/napptive/catalog-manager/internal/pkg/config"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/njwt/pkg/interceptors"
	"github.com/rs/zerolog/log"
)

// PermissionResolver with a struct to manage account permissions
type PermissionResolver struct {
	authEnabled bool
	teamConfig  config.TeamConfig
}

// NewPermissionResolver returns a PermissionResolver object
func NewPermissionResolver(authEnabled bool, teamConfig config.TeamConfig) *PermissionResolver {
	return &PermissionResolver{
		authEnabled: authEnabled,
		teamConfig:  teamConfig,
	}
}

// CheckAccountPermissions checks if a user belongs to an account, this is required with the private applications or with certain operations with the public ones
func (pr *PermissionResolver) CheckAccountPermissions(ctx context.Context, applicationName string, requireAdminPrivilege bool) (*bool, error) {
	allowed := true
	notAllowed := false

	if !pr.authEnabled {
		return &allowed, nil
	}

	// get the repoName
	_, applicationInfo, err := utils.DecomposeApplicationID(applicationName)
	if err != nil {
		log.Error().Err(err).Str("application_name", applicationName).Msg("error getting application name, unable to check permissions")
		return &notAllowed, err
	}

	// getClaim
	claim, err := interceptors.GetClaimFromContext(ctx)
	if err != nil {
		return &notAllowed, err
	}

	namespace := applicationInfo.Namespace

	// If the user is an admin and the namespace is a teamNamespace -> return true
	if pr.isPrivilegedUser(claim.Username) && pr.isTeamNamespace(namespace) {
		log.Debug().Str("user_id", claim.UserID).Str("username", claim.Username).Str("namespace", namespace).
			Msg("Privileged user can operate in a team namespace")
		return &allowed, nil
	}

	// Check if the user can operate in the requested account
	isAuthorized := claim.IsAuthorized(namespace, requireAdminPrivilege)
	if isAuthorized {
		log.Debug().Str("user_id", claim.UserID).Str("username", claim.Username).Str("namespace", namespace).
			Msg("user can operate in the namespace")
	} else {
		log.Debug().Str("user_id", claim.UserID).Str("username", claim.Username).Str("namespace", namespace).
			Msg("user can NOT operate in the namespace")
	}
	return &isAuthorized, nil
}

// isTeamNamespace checks the namespace role
func (pr *PermissionResolver) isTeamNamespace(repoName string) bool {
	if !pr.teamConfig.Enabled {
		return false
	}
	for _, repo := range pr.teamConfig.TeamNamespaces {
		if repo == repoName {
			return true
		}
	}

	return false
}

// isPrivilegedUser checks the user role
func (pr *PermissionResolver) isPrivilegedUser(userName string) bool {
	if !pr.teamConfig.Enabled {
		return false
	}
	for _, user := range pr.teamConfig.PrivilegedUsers {
		if user == userName {
			return true
		}
	}

	return false
}
