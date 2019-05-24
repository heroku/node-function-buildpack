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
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/detect"

	"path/filepath"
)

func DetectNode(d detect.Detect) (bool, error) {
	path := filepath.Join(d.Application.Root)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	var jsFunctionFileName string
	var jsFileCount int

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".js" {
				jsFunctionFileName = file.Name()
				jsFileCount++
				if jsFileCount > 1 {
					return false, errors.New("found more than one .js file")
				}
			}
		}
	}

	if jsFileCount == 0 {
		return false, errors.New("missing .js file")
	}

	err = validatePackageJson(filepath.Join(path, "package.json"), jsFunctionFileName)
	if err != nil {
		return false, err
	}

	return true, nil
}

func validatePackageJson(packageJsonFile, mainJsFunctionFile string) error {
	if !fileExists(packageJsonFile) {
		return errors.New("missing package.json file")
	}

	var data []byte
	data, err := ioutil.ReadFile(packageJsonFile)
	if err != nil {
		return err
	}

	packageJson := struct {
		Main string `json:"main"`
	}{}
	if err := json.Unmarshal(data, &packageJson); err != nil {
		return err
	}

	if packageJson.Main == "" || packageJson.Main != mainJsFunctionFile {
		return errors.New("invalid or missing \"main\" field in package.json")
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
