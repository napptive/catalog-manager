package config

import (
	"strings"

	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
)

// TeamConfig with the configuration related to the priviledged team access.
type TeamConfig struct {
	Enabled         bool
	PrivilegedUsers []string
	TeamNamespaces  []string
}

func NewTeamConfig(enabled bool, users string, repositories string) TeamConfig {
	var privilegedUsers []string
	var teamNamespaces []string
	if users != "" {
		privilegedUsers = strings.Split(users, " ")
	}
	if repositories != "" {
		teamNamespaces = strings.Split(repositories, " ")
	}
	return TeamConfig{
		Enabled:         enabled,
		PrivilegedUsers: privilegedUsers,
		TeamNamespaces:  teamNamespaces,
	}
}

func (t *TeamConfig) IsValid() error {
	if t.Enabled {
		if len(t.PrivilegedUsers) == 0 || len(t.TeamNamespaces) == 0 {
			return nerrors.NewFailedPreconditionError("team enabled needs privileged users and team repositories")
		}
	}
	return nil
}

func (t *TeamConfig) Print() {
	if t.Enabled {
		log.Info().Str("PrivilegedUsers", strings.Join(t.PrivilegedUsers, " ")).Msg("Privileged Users")
		log.Info().Str("namespaces", strings.Join(t.TeamNamespaces, " ")).Msg("Team")
	}
}
