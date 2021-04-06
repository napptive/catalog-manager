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

package users

import (
	"encoding/base64"
	"fmt"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/provider/user-provider"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/napptive/rdbms/pkg/rdbms"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"syreclabs.com/go/faker"
	"time"
)

type Manager struct{
	connString string
}

func NewManager (connString string) *Manager {
	return &Manager{connString: connString}
}


// generateSaltedPassword generates and returns the salted_password
// - Hash(concatenate salt and password)
func (m *Manager) generateSaltedPassword (password string, salt string) (string, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("%s%s", password, salt)), bcrypt.DefaultCost)
	if err != nil {
		return "", nerrors.NewInternalErrorFrom(err, "error hashing the password")
	}
	b64 := base64.StdEncoding.EncodeToString(hashedPassword)
	return b64, nil
}

// CreateUser stores new user
func (m *Manager) CreateUser (username string, password string) error{
	salt := faker.Lorem().Characters(10)
	// Generate
	saltedPassword, err := m.generateSaltedPassword(password, salt)
	if err != nil {
		return err
	}

	// Create Provider
	conn, err := rdbms.NewRDBMS().PoolConnect(context.Background(), m.connString)
	if err != nil {
		return err
	}
	provider := user_provider.NewUserProvider(conn, time.Second *10)

	_, err = provider.Add(&entities.User{
		Username:       username,
		Salt:           salt,
		SaltedPassword: saltedPassword,
	})
	if err != nil {
		return err
	}
	return nil
}

// CheckPassword compares a hashed password with a specif password.
func (m *Manager) CheckPassword(b64Password string, password string) error {

	hashedPassword, err := base64.StdEncoding.DecodeString(b64Password)
	if err != nil {
		return nerrors.NewInvalidArgumentError("error checking password")
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return nerrors.NewUnauthenticatedError("password is not valid")
	}
	return nil
}

// TODO: returns JWT token
// LoginUser gets the user and checks the credentials
func (m *Manager) LoginUser (username string, password string) error{
	// Create Provider
	conn, err := rdbms.NewRDBMS().PoolConnect(context.Background(), m.connString)
	if err != nil {
		return err
	}
	provider := user_provider.NewUserProvider(conn, time.Second *10)

	user, err := provider.Get(username)
	if err != nil {
		return err
	}

	if err = m.CheckPassword(user.SaltedPassword, fmt.Sprintf("%s%s", password, user.Salt)); err != nil {
		return err
	}
	return nil
}