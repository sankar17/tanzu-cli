// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-cli/pkg/constants"
)

// SaveFile saves the file to the provided path
// Also creates missing directories if any
func SaveFile(filePath string, data []byte) error {
	dirName := filepath.Dir(filePath)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			return merr
		}
	}

	err := os.WriteFile(filePath, data, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "unable to save file '%s'", filePath)
	}

	return nil
}

// CopyFile copies source file to dest file
func CopyFile(sourceFile, destFile string) error {
	input, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	err = os.WriteFile(destFile, input, constants.ConfigFilePermissions)
	return err
}

// PathExists returns true if file/directory exists otherwise returns false
func PathExists(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
