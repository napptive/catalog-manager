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

package commands

import (
	"time"

	catalog_manager "github.com/napptive/catalog-manager/internal/app/catalog-manager"
	"github.com/spf13/cobra"
)

var runCmdLongHelp = "Launch the catalog-manager service"
var runCmdShortHelp = "Launch the service"
var runCmdExample = `$ catalog-manager run`
var runCmdUse = "run"

var runCmd = &cobra.Command{
	Use:     runCmdUse,
	Long:    runCmdLongHelp,
	Example: runCmdExample,
	Short:   runCmdShortHelp,
	Run: func(cmd *cobra.Command, args []string) {
		cfg.Debug = debugLevel
		s := catalog_manager.NewService(cfg)
		s.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().IntVar(&cfg.CatalogManager.GRPCPort, "grpcPort", 7060, "gRPC Port to launch the Catalog-manager")
	runCmd.Flags().IntVar(&cfg.CatalogManager.HTTPPort, "httpPort", 7061, "HTTP Port to launch the Catalog-manager")
	runCmd.Flags().StringVar(&cfg.ElasticAddress, "elasticAddress", "http://localhost:9200", "address to connect to Elastic Search")
	runCmd.Flags().StringVar(&cfg.Index, "index", "napptive", "Elastic Index to store the repositories")
	runCmd.Flags().StringVar(&cfg.RepositoryPath, "repositoryPath", "/napptive/repository/", "base path to store the repositories")
	runCmd.Flags().StringVar(&cfg.CatalogUrl, "repositoryUrl", "", "Repository URL")
	runCmd.Flags().BoolVar(&cfg.JWTConfig.AuthEnabled, "authEnabled", false, "Enable Authentication")
	runCmd.Flags().StringVar(&cfg.JWTConfig.Header, "authHeader", "authorization", "Authorization header name")
	runCmd.Flags().StringVar(&cfg.JWTConfig.Secret, "authSecret", "secret", "Authorization secret to validate JWT signatures")
	runCmd.Flags().BoolVar(&cfg.TeamConfig.Enabled, "teamPrivileges", false, "Enable Team Privileges")
	runCmd.Flags().StringSliceVar(&cfg.PrivilegedUsers, "teamUsers", nil, "Privileged Users")
	runCmd.Flags().StringSliceVar(&cfg.TeamNamespaces, "teamRepositories", nil, "Team Repositories")

	runCmd.Flags().BoolVar(&cfg.BQConfig.Enabled, "analyticsEnabled", false, "Analytics enabled")
	runCmd.Flags().StringVar(&cfg.BQConfig.Config.ProjectID, "analyticsProjectID", "", "GKE project for analytics")
	runCmd.Flags().StringVar(&cfg.BQConfig.Config.CredentialsPath, "analyticsCredentialsPath", "/analytics/credentials.json", "credentials path for analytics")
	runCmd.Flags().StringVar(&cfg.BQConfig.Config.Schema, "analyticsSchema", "analytics", "analytics schema")
	runCmd.Flags().StringVar(&cfg.BQConfig.Config.Table, "analyticsTable", "operation", "analytics table")
	runCmd.Flags().DurationVar(&cfg.BQConfig.Config.SendingTime, "analyticsLoop", 5*time.Second, "time to send the data to analytics provider")

}
