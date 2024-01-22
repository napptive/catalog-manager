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

package cli

import (
	"fmt"

	"github.com/napptive/catalog-manager/internal/pkg/server/users"
	"github.com/napptive/nerrors/pkg/nerrors"
)

type UserCli struct {
}

func NewUserCli() *UserCli {
	return &UserCli{}
}

// CreateNewUser adds new user in database
func (uc *UserCli) CreateNewUser(username string, password string, connString string) error {
	userManager := users.NewManager(connString)
	err := userManager.CreateUser(username, password)
	if err != nil {
		fmt.Println(nerrors.FromError(err).String())
	} else {
		fmt.Printf("User created")
	}
	return nil
}

// LoginUser check
func (uc *UserCli) LoginUser(username string, password string, connString string) error {
	userManager := users.NewManager(connString)
	err := userManager.LoginUser(username, password)
	if err != nil {
		fmt.Println(nerrors.FromError(err).String())
	} else {
		fmt.Printf("Login success")
	}
	return nil
}
