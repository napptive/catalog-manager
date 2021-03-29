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
package catalog_manager

//go:generate  mockgen -destination provider_mock.go -package=catalog_manager github.com/napptive/catalog-manager/internal/pkg/provider MetadataProvider
//go:generate  mockgen -destination storage_mock.go -package=catalog_manager github.com/napptive/catalog-manager/internal/pkg/storage StorageManager
////go:generate  mockgen -destination manager_mock.go -self_package  github.com/napptive/catalog-manager/internal/server/catalog-manager -package=catalog_manager . Manager


//Mock is a place holder to unify all mock generators.
func Mock() {}

