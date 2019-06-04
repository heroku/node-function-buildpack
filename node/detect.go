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
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/detect"

	"path/filepath"
)

func DetectNode(d detect.Detect) (bool, error) {
	jsFiles, err := filepath.Glob(filepath.Join(d.Application.Root, "*.js"))
	if err != nil {
		log.Fatal(err)
	}

	err = validateSourceFiles(jsFiles)
	if err != nil {
		return false, err
	}

	err = validatePackageJson(filepath.Join(d.Application.Root, "package.json"), jsFiles)
	if err != nil {
		return false, err
	}

	return true, nil
}

func validateSourceFiles(jsFiles []string) error {
	if len(jsFiles) == 0 {
		return errors.New("no .js source files were found")
	}

	return nil
}

func validatePackageJson(packageJsonFile string, jsFiles []string) error {
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

	if packageJson.Main == "" {
		return errors.New("missing \"main\" field in package.json")
	}

	for _, jsFile := range jsFiles {
		_, filename := filepath.Split(jsFile)
		if packageJson.Main == filename {
			return nil
		}
	}

	return errors.New(fmt.Sprintf("could not find \"%s\"", packageJson.Main))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
