package config

import (
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"strings"
)

type TeamConfig struct {
	Enabled          bool
	PrivilegedUsers  []string
	TeamRepositories []string
}

func NewTeamConfig (enabled bool, users string, repositories string) TeamConfig {
	var privilegedUsers []string
	var teamRepositories []string
	if users != "" {
		privilegedUsers = strings.Split(users, " ")
	}
	if repositories != "" {
		teamRepositories = strings.Split(repositories, " ")
	}
	return TeamConfig{
		Enabled:          enabled,
		PrivilegedUsers:  privilegedUsers,
		TeamRepositories: teamRepositories,
	}
}

func (t *TeamConfig) IsValid() error  {
	if t.Enabled {
		if len(t.PrivilegedUsers) == 0 || len(t.TeamRepositories) == 0 {
			return nerrors.NewFailedPreconditionError("team enabled needs privileged users and team repositories")
		}
	}
	return nil
}

func (t *TeamConfig) Print() {
	if t.Enabled {
		log.Info().Str("PrivilegedUsers", strings.Join(t.PrivilegedUsers, " ")).Msg("Privileged Users")
		log.Info().Str("TeamRepositories", strings.Join(t.TeamRepositories, " ")).Msg("Repositories of the Team")
	}
}