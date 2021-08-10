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
package commands

import (
	"github.com/napptive/catalog-manager/internal/app/cli"
	"github.com/spf13/cobra"
)

var appCmdLongHelp = `Manage apps`

var appCmdShortHelp = `Manage apps`

var appCmd = &cobra.Command{
	Use:   "apps",
	Long:  appCmdLongHelp,
	Short: appCmdShortHelp,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace := ""
		if len(args) > 0 {
			namespace = args[0]
		}
		op := cli.NewApplicationCli(cfg.Index, cfg.ElasticAddress, cfg.RepositoryPath, "")
		return op.List(namespace)
	},
}

var deleteAppCmdLongHelp = `Delete an application from catalog`

var deleteAppCmdShortHelp = `Delete catalog application`

var deleteAppCmd = &cobra.Command{
	Use:   "delete <applicationName>",
	Long:  deleteAppCmdLongHelp,
	Short: deleteAppCmdShortHelp,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		op := cli.NewApplicationCli(cfg.Index, cfg.ElasticAddress, cfg.RepositoryPath, "")
		return op.DeleteApplication(args[0])
	},
}
func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(deleteAppCmd)

	appCmd.PersistentFlags().StringVar(&cfg.Index, "index", "napptive", "Elastic Index to store the repositories")
	appCmd.PersistentFlags().StringVar(&cfg.RepositoryPath, "repositoryPath", "/napptive/repository/", "base path to store the repositories")
	appCmd.PersistentFlags().StringVar(&cfg.ElasticAddress, "elasticAddress", "http://localhost:9200", "address to connect to Elastic Search")
}
