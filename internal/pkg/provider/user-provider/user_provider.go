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
package user_provider

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/napptive/rdbms/pkg/rdbms"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	Schema 			     string = "catalog"
	UserTable            string = "users"
	UsernameColumn       string = "username"
	SaltColumn           string = "salt"
	SaltedPasswordColumn string = "salted_password"
)

// NewUserProvider generate a new User provider.
func NewUserProvider(conn rdbms.Conn, timeout time.Duration) UserProvider {
	return &userProvider{
		conn:    conn,
		timeout: timeout,
	}
}

type userProvider struct {
	conn    rdbms.Conn
	timeout time.Duration
}

// Add adds new user
func (u *userProvider) Add(user *entities.User) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	t := goqu.T(UserTable).Schema(Schema)

	sql, args, err := goqu.Insert(t).
		Cols(UsernameColumn, SaltColumn, SaltedPasswordColumn).Vals(
		goqu.Vals{user.Username, user.Salt, user.SaltedPassword},
	).ToSQL()

	if err != nil {
		log.Err(err).Msg("error inserting user")
		return nil, nerrors.NewInternalErrorFrom(err, "Error adding user [%s]", user.Username)
	}

	_, err = u.conn.Exec(ctx, sql, args...)
	if err != nil {
		log.Err(err).Msg("error executing insert user")
		return nil, nerrors.NewInternalErrorFrom(err, "Error adding user [%s]", user.Username)
	}
	return user, nil
}

// Remove deletes an existing user
func (u *userProvider) Remove(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	t := goqu.T(UserTable).Schema(Schema)

	sql, args, err := goqu.Delete(t).Where(goqu.Ex{
		UsernameColumn: username,
	}).ToSQL()
	if err != nil {
		log.Err(err).Msg("error removing user")
		return nerrors.NewInternalErrorFrom(err, "Error deleting user [%s]", username)
	}
	result, err := u.conn.Exec(ctx, sql, args...)
	if err != nil {
		log.Err(err).Msg("error executing remove user")
		return nerrors.NewInternalErrorFrom(err, "Error deleting user [%s]", username)
	}
	if result.RowsAffected() == 0 {
		return nerrors.NewNotFoundError("The user is not found [%s]", username)
	}

	return nil
}

// List retrieves a list of users
func (u *userProvider) List() ([]*entities.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	t := goqu.T(UserTable).Schema(Schema)

	sql, args, err := goqu.From(t).
		Select(UsernameColumn, SaltColumn, SaltedPasswordColumn).
		ToSQL()
	if err != nil {
		return nil, nerrors.NewInternalErrorFrom(err, "Error listing users")

	}
	rows, err := u.conn.Query(ctx, sql, args...)
	if err != nil {
		log.Err(err).Msg("error listing users")
		return nil, nerrors.NewInternalErrorFrom(err, "Error listing users")
	}

	userList := make([]*entities.User, 0)
	for rows.Next() {
		user := &entities.User{}
		err = rows.Scan(&user.Username, &user.Salt, &user.SaltedPassword)
		if err != nil {
			log.Err(err).Msg("error getting users when listing them")
			return nil, nerrors.NewInternalErrorFrom(err, "Error listing users")
		}
		userList = append(userList, user)
	}

	return userList, nil
}

// Get retrieves a user
func (u *userProvider) Get(username string) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.timeout)
	defer cancel()

	t := goqu.T(UserTable).Schema(Schema)

	sql, args, err := goqu.From(t).
		Select(UsernameColumn, SaltColumn, SaltedPasswordColumn).
		Where(goqu.Ex{
			UsernameColumn: username,
		}).ToSQL()
	if err != nil {
		log.Err(err).Msg("error getting user")
		return nil, nerrors.NewInternalErrorFrom(err, "Error getting user [%s]", username)
	}
	row := u.conn.QueryRow(ctx, sql, args...)
	user := &entities.User{}
	err = row.Scan(&user.Username, &user.Salt, &user.SaltedPassword)
	if err != nil {
		log.Err(err).Msg("error in scan when getting user")
		return nil, nerrors.NewInternalErrorFrom(err, "Error getting user [%s]", username)
	}
	return user, nil
}
