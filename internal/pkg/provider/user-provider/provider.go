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

package user_provider

import "github.com/napptive/catalog-manager/internal/pkg/entities"

// UserProvider is an interface with user operations
type UserProvider interface {
	// Add adds new user
	Add(user *entities.User) (*entities.User, error)
	// Remove deletes an existing user
	Remove(username string) error
	// List retrieves a list of users
	List() ([]*entities.User, error)
	// Get retrieves a user
	Get(username string) (*entities.User, error)
}
