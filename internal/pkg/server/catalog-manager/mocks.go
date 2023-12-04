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
package catalog_manager

//go:generate  mockgen -destination metadata_provider_mock.go -package=catalog_manager github.com/napptive/catalog-manager/internal/pkg/provider/metadata MetadataProvider
//go:generate  mockgen -destination storage_mock.go -package=catalog_manager github.com/napptive/catalog-manager/internal/pkg/storage StorageManager
//go:generate  mockgen -destination catalog_add_server_mock.go -package=catalog_manager github.com/napptive/grpc-catalog-go Catalog_AddServer
//go:generate  mockgen -destination manager_mock.go  -package=catalog_manager  github.com/napptive/catalog-manager/internal/pkg/server/catalog-manager Manager

// Mock is a place holder to unify all mock generators.
func Mock() {}
