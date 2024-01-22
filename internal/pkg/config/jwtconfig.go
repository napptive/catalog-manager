package config

import (
	nwjtConfig "github.com/napptive/njwt/pkg/config"
	"github.com/rs/zerolog/log"
	"strings"
)

// JWTConfig struct with njwt configuration
type JWTConfig struct {
	// AuthEnabled is a flag indicating
	AuthEnabled bool
	// JWTConfig with the JWT specific configuration.
	nwjtConfig.JWTConfig
}

// IsValid checks if the configuration options are valid.
func (c *JWTConfig) IsValid() error {
	if c.AuthEnabled {
		return c.JWTConfig.IsValid()
	}

	return nil
}

// Print prints the configuration
func (c *JWTConfig) Print() error {
	if c.AuthEnabled {
		log.Info().Str("header", c.Header).
			Str("secret", strings.Repeat("*", len(c.Secret))).Msg("Authorization")
	}

	return nil
}
