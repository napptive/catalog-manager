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

package storage

import (
	"fmt"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"strings"
)

// StorageManager is a struct to manage all the storage operations
type StorageManager struct {
	// basePath with the path where the repo storage is
	basePath string
}

func NewStorageManager(basePath string) *StorageManager {
	return &StorageManager{basePath: basePath}
}

func (s *StorageManager) removeDirectory(name string) error {
	if err := os.RemoveAll(name); err != nil {
		return nerrors.FromError(err)
	}
	return nil
}

func (s *StorageManager) createDirectory(name string) error {
	if err := os.MkdirAll(name, 0755); err != nil {
		return nerrors.FromError(err)
	}
	return nil
}

// CreateRepository creates a directory to storage a repository
func (s *StorageManager) CreateRepository(name string) error {

	if err := s.createDirectory(fmt.Sprintf("%s/%s", s.basePath, name)); err != nil {
		log.Err(err).Str("name", name).Msg("error creating repository")
		return err
	}
	return nil
}

// RepositoryExists checks if a repository exists
func (s *StorageManager) RepositoryExists(name string) (bool, error) {
	dir, err := os.Open(s.basePath)
	if err != nil {
		return false, nerrors.FromError(err)
	}
	defer dir.Close()

	repositories, err := dir.Readdirnames(0)
	if err != nil {
		return false, nerrors.FromError(err)
	}

	for _, repo := range repositories {
		log.Debug().Str("repo", repo).Msg("Repository Name")
		if repo == name {
			return true, nil
		}
	}

	return false, nil
}

// RemoveRepository removes the repository directory. Be careful using this function
// it will remove ALL the applications
func (s *StorageManager) RemoveRepository(name string) error {
	if err := s.removeDirectory(fmt.Sprintf("%s/%s", s.basePath, name)); err != nil {
		log.Err(err).Str("name", name).Msg("error removing repository")
		return err
	}
	return nil
}

// StoreApplication save all files in their corresponding path
func (s *StorageManager) StoreApplication(repo string, name string, version string, files []*entities.FileInfo) error {
	// 1.- Remove the old application
	// baseUrl/repo/application/tag
	dir := fmt.Sprintf("%s/%s/%s/%s", s.basePath, repo, name, version)
	if err := s.removeDirectory(dir); err != nil {
		log.Err(err).Str("application", dir).Msg("Error storing application, unable to delete old one")
		return err
	}

	// 2.- Create the application directory
	if err := s.createDirectory(dir); err != nil {
		log.Err(err).Str("application", dir).Msg("Error storing application, unable to create directory")
		return err
	}

	// 3.- Create the files and storage them
	for _, appFile := range files {
		// Check if the file is in a new directory
		splited := strings.Split(appFile.Path, "/")
		// create directory
		if len(splited) > 1 {
			pp := splited[:len(splited)-1]
			if err := s.createDirectory(fmt.Sprintf("%s/%s", dir, strings.Join(pp, "/"))); err != nil {
				return err
			}
		}
		file, err := os.Create(fmt.Sprintf("%s/%s", dir, appFile.Path))
		if err != nil {
			log.Err(err).Str("application", dir).Msg("Error storing application file, unable to create the file")
			// TODO: remove all stored
			return nerrors.FromError(err)
		}
		defer file.Close()

		if _, err = file.Write(appFile.Data); err != nil {
			log.Err(err).Str("application", dir).Msg("Error storing application file, unable to save the file")
			// TODO: when get/list applications are implemented,
			// test all cases where the add metadata fails and see if it should be rolled back
			return nerrors.FromError(err)
		}

		if err := file.Sync(); err != nil {
			return nerrors.FromError(err)
		}
	}

	return nil
}

// ApplicationExists checks if an application exists
func (s *StorageManager) ApplicationExists(name string) (bool, error) {
	return false, nerrors.NewUnimplementedError("not implemented yet!")
}

// RemoveApplication removes an application, returns an error if it does not exist
func (s *StorageManager) RemoveApplication(name string) (bool, error) {
	return false, nerrors.NewUnimplementedError("not implemented yet!")
}

func (s *StorageManager) GetApplication(repo string, name string, version string) ([]*entities.FileInfo, error) {

	// Find the application directory
	path := fmt.Sprintf("%s/%s/%s/%s", s.basePath, repo, name, version)
	return s.loadAppFile(path)
}

func (s *StorageManager) loadAppFile(path string) ([]*entities.FileInfo, error) {

	fileInfo, err := ioutil.ReadDir(path)
	// TODO: ver como puedo identificar un errorNotFound de otro
	if err != nil {
		return nil, nerrors.NewNotFoundErrorFrom(err, "Application not found")
	}

	var filesToReturn []*entities.FileInfo
	for _, file := range fileInfo {
		if file.IsDir() {
			files, err := s.loadAppFile(fmt.Sprintf("%s/%s", path, file.Name()))
			if err != nil {
				return nil, err
			}
			filesToReturn = append(filesToReturn, files...)
		} else {
			newPath := fmt.Sprintf("%s/%s", path, file.Name())
			data, err := ioutil.ReadFile(newPath)
			if err != nil {
				return nil, nerrors.NewInternalErrorFrom(err, "Error reading file")
			}
			filesToReturn = append(filesToReturn, &entities.FileInfo{
				Path: newPath,
				Data: data,
			})
		}

	}

	return filesToReturn, nil
}
