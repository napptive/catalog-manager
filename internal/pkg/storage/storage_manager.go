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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type StorageManager interface {
	// StoreApplication save all files in their corresponding path
	StoreApplication(repo string, name string, version string, files []*entities.FileInfo) error
	// GetApplication returns the application files
	GetApplication(repo string, name string, version string, compressed bool) ([]*entities.FileInfo, error)
	// RemoveApplication removes an application, returns an error if it does not exist
	RemoveApplication(repo string, name string, version string) error
	// ApplicationExists checks if an application exists
	ApplicationExists(repo string, name string, version string) (bool, error)
	// CreateRepository creates a directory to storage a repository
	CreateRepository(name string) error
	// RepositoryExists checks if a repository exists
	RepositoryExists(name string) (bool, error)
	// RemoveRepository removes the repository directory. Be careful using this function
	RemoveRepository(name string) error
}

// StorageManager is a struct to manage all the storage operations
type storageManager struct {
	// basePath with the path where the repo storage is
	basePath string
}

func NewStorageManager(basePath string) StorageManager {
	return &storageManager{basePath: basePath}
}

// getAppDirectory compose the application directory
func (s *storageManager) getAppDirectory(repo string, name string, version string) string {
	return fmt.Sprintf("%s/%s/%s/%s", s.basePath, repo, name, version)
}

func (s *storageManager) removeDirectory(name string) error {
	if err := os.RemoveAll(name); err != nil {
		return nerrors.FromError(err)
	}
	return nil
}

func (s *storageManager) createDirectory(name string) error {
	if err := os.MkdirAll(name, 0755); err != nil {
		return nerrors.FromError(err)
	}
	return nil
}

// CreateRepository creates a directory to storage a repository
func (s *storageManager) CreateRepository(name string) error {

	if err := s.createDirectory(fmt.Sprintf("%s/%s", s.basePath, name)); err != nil {
		log.Err(err).Str("name", name).Msg("error creating repository")
		return err
	}
	return nil
}

// RepositoryExists checks if a repository exists
func (s *storageManager) RepositoryExists(name string) (bool, error) {
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
func (s *storageManager) RemoveRepository(name string) error {
	if err := s.removeDirectory(fmt.Sprintf("%s/%s", s.basePath, name)); err != nil {
		log.Err(err).Str("name", name).Msg("error removing repository")
		return err
	}
	return nil
}

// StoreApplication save all files in their corresponding path
func (s *storageManager) StoreApplication(repo string, name string, version string, files []*entities.FileInfo) error {
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
func (s *storageManager) ApplicationExists(repo string, name string, version string) (bool, error) {
	appDir := s.getAppDirectory(repo, name, version)
	_, err := os.Stat(appDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, nerrors.NewInternalErrorFrom(err, "unable to check if the application exists")
		}
	}
	return true, nil
}

// RemoveApplication removes an application, returns an error if it does not exist
func (s *storageManager) RemoveApplication(repo string, name string, version string) error {
	appName := s.getAppDirectory(repo, name, version)

	exists, err := s.ApplicationExists(repo, name, version)
	if err != nil {
		log.Err(err).Str("appName", appName).Msg("error deleting application")
		return nerrors.NewInternalErrorFrom(err, "unable to delete application")
	}
	if !exists {
		return nerrors.NewNotFoundError("unable to delete application")
	}

	if err := s.removeDirectory(appName); err != nil {
		log.Err(err).Str("appName", appName).Msg("error deleting application")
		return nerrors.NewInternalErrorFrom(err, "unable to delete application")
	}
	// clean directories
	if err := s.cleanApplicationDirectory(repo, name); err != nil{
		log.Err(err).Str("appName", appName).Msg("error cleaning application directory")
	}
	return nil
}

// GetApplication returns the application files
func (s *storageManager) GetApplication(repo string, name string, version string, compressed bool) ([]*entities.FileInfo, error) {

	// Find the application directory
	path := fmt.Sprintf("%s%s/%s/%s", s.basePath, repo, name, version)
	if compressed {
		return s.loadAppFileTgz(name, path)
	}
	return s.loadAppFile(path, fmt.Sprintf("./%s", name))
}

func (s *storageManager) loadAppFileTgz(name string, path string)([]*entities.FileInfo, error) {
	var buf bytes.Buffer

	zr := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zr)

	// walk through every file in the folder
	if wErr := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		// generate tar header
		header, nErr := tar.FileInfoHeader(fi, file)
		if nErr != nil {
			return nErr
		}

		// remove repo path
		header.Name = filepath.ToSlash(fmt.Sprintf("./%s%s", name, strings.Replace(file, path, "", -1)))

		// write header
		if nErr := tw.WriteHeader(header); nErr != nil {
			return nErr
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, nErr := os.Open(file)
			if nErr != nil {
				return err
			}
			if _, nErr := io.Copy(tw, data); nErr != nil {
				return nErr
			}
		}
		return nil
	}); wErr != nil {
		return nil, nerrors.NewInternalErrorFrom(wErr, "Error getting application")

	}

	// produce tar
	if nErr := tw.Close(); nErr != nil {
		return nil, nerrors.NewInternalErrorFrom(nErr, "Error closing tar")
	}
	// produce gzip
	if nErr := zr.Close(); nErr != nil {
		return nil, nerrors.NewInternalErrorFrom(nErr, "Error closing tgz")
	}

	return []*entities.FileInfo{&entities.FileInfo{
		Path: fmt.Sprintf("./%s.tgz", name),
		Data: buf.Bytes(),
	}}, nil
}

// loadAppFile gets the content of application files
func (s *storageManager) loadAppFile(path string, filePath string) ([]*entities.FileInfo, error) {

	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, nerrors.NewNotFoundErrorFrom(err, "Application not found")
	}

	var filesToReturn []*entities.FileInfo
	for _, file := range fileInfo {
		if file.IsDir() {
			files, err := s.loadAppFile(fmt.Sprintf("%s/%s", path, file.Name()), fmt.Sprintf("%s/%s", filePath, file.Name()))
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
				Path: fmt.Sprintf("%s/%s", filePath, file.Name()),
				Data: data,
			})
		}

	}

	return filesToReturn, nil
}

// checkEmptyDirs check if a directory is empty
func (s *storageManager) checkEmptyDirs(path string) (bool, error){

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

// cleanApplicationDirectory checks if application directory is empty and removes it and
// check if repository directory is empty and removes it
func (s *storageManager) cleanApplicationDirectory(repo string, name string) error{

	// - check if the application removed was the unique version for this application
	// basePath/repository/app
	appPath := fmt.Sprintf("%s/%s/%s", s.basePath, repo, name)
	last, err := s.checkEmptyDirs(appPath)
	if err != nil {
		return nerrors.FromError(err)
	}

	// remove the app directory
	if last {
		log.Debug().Str("path", appPath).Msg("removing directory")
		if err = s.removeDirectory(appPath); err != nil {
			return nerrors.FromError(err)
		}
	}
	// basePath/repository
	appPath = fmt.Sprintf("%s/%s", s.basePath, repo)
	last, err = s.checkEmptyDirs(appPath)
	if err != nil {
		return nerrors.FromError(err)
	}

	// remove the app directory
	if last {
		log.Debug().Str("path", appPath).Msg("removing directory")
		if err = s.removeDirectory(appPath); err != nil {
			return nerrors.FromError(err)
		}
	}
	return nil
}