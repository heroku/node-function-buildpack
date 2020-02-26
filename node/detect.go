/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package node

import (
	"errors"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/detect"

	"path/filepath"
)

func DetectNode(d detect.Detect) (bool, error) {
	err := validatePackageJson(d.Application.Root)
	if err != nil {
		return false, err
	}

	return true, nil
}

func validatePackageJson(applicationRoot string) error {
	var packageJsonFile = filepath.Join(applicationRoot, "package.json")
	if !fileExists(packageJsonFile) {
		return errors.New("could not find a package.json file")
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
