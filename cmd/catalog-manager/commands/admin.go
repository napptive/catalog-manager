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

var adminCmdLongHelp = `admin commands`
var adminCmdShortHelp = `admin commands`

var adminCmd = &cobra.Command{
	Use:   "admin",
	Long:  adminCmdLongHelp,
	Short: adminCmdShortHelp,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var listCmdLongHelp = `list the catalog applications`
var listCmdShortHelp = `list the catalog applications`
var listCmd = &cobra.Command{
	Use:     "list [namespace]",
	Long:    listCmdLongHelp,
	Short:   listCmdShortHelp,
	Aliases: []string{"ls"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace := ""
		if len(args) > 0 {
			namespace = args[0]
		}
		op, err := cli.NewApplicationCli(cfg.AdminGRPCPort)
		if err != nil {
			return err
		}
		return op.List(namespace)
	},
}

var deleteAppCmdLongHelp = `Delete applications from catalog. 
You can delete a namespace indicating the name as arg or delete only one application by passing the name as namespace/applicationName`
var deleteAppCmdShortHelp = `Delete catalog application`

var deleteAppCmd = &cobra.Command{
	Use:     "delete <applicationName>",
	Long:    deleteAppCmdLongHelp,
	Short:   deleteAppCmdShortHelp,
	Aliases: []string{"rm", "remove"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		op, err := cli.NewApplicationCli(cfg.AdminGRPCPort)
		if err != nil {
			return err
		}
		return op.Delete(args[0])
	},
}

func init() {
	rootCmd.AddCommand(adminCmd)

	adminCmd.AddCommand(deleteAppCmd)
	adminCmd.AddCommand(listCmd)

	adminCmd.PersistentFlags().IntVar(&cfg.CatalogManager.AdminGRPCPort, "adminGRPCPort", 7062, "gRPC Port to connect the Catalog-manager admin API")
}
