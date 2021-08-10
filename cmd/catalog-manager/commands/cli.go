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
	cli2 "github.com/napptive/catalog-manager/internal/app/cli"
	"github.com/spf13/cobra"
)

var connString string

var userCmdLongHelp = `Manage users`

var userCmdShortHelp = `Manage users`

var userCmd = &cobra.Command{
	Use:   "user",
	Long:  userCmdLongHelp,
	Short: userCmdShortHelp,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var createUserCmdLongHelp = `Create new user to be able to log in to the catalog `

var createUserCmdShortHelp = `Create new user`

var createUserCmd = &cobra.Command{
	Use:   "create <username> <password>",
	Long:  createUserCmdLongHelp,
	Short: createUserCmdShortHelp,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cli := cli2.NewUserCli()
		return cli.CreateNewUser(args[0], args[1], connString)
	},
}

var loginUserCmdLongHelp = `Login into Napptive Catalog`

var loginUserCmdShortHelp = `Login into Napptive Catalog`

var loginUserCmd = &cobra.Command{
	Use:   "login <username> <password>",
	Long:  loginUserCmdLongHelp,
	Short: loginUserCmdShortHelp,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cli := cli2.NewUserCli()
		return cli.LoginUser(args[0], args[1], connString)
	},
}

func init() {

	userCmd.PersistentFlags().StringVarP(&connString, "connectionString", "c", "host=postgres user=postgres password=postgres port=5432","connection string to connect postgres database")

	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(createUserCmd)
	userCmd.AddCommand(loginUserCmd)
}
