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
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"io/ioutil"
	"log"

	"path/filepath"
)

// DetectNode answers true if there is only one .js file in the root path
func DetectNode(d detect.Detect) (bool, error) {
	path := filepath.Join(d.Application.Root)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	numberOfJsFiles := 0
	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".js" {
				numberOfJsFiles++
				if numberOfJsFiles > 1 {
					return false, errors.New("found more than one .js file")
				}
			}
		}
	}
	if numberOfJsFiles == 0 {
		return false, errors.New("missing .js file")
	}

	return true, nil
}
