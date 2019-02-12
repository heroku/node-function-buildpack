/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/npm-cnb/modules"
	"github.com/projectriff/node-function-buildpack/metadata"
	"github.com/projectriff/node-function-buildpack/node"
)

const (
	Error_Initialize          = 101
	Error_ReadMetadata        = 102
	Error_DetectedNone        = 103
	Error_DetectAmbiguity     = 104
	Error_UnsupportedLanguage = 105
	Error_DetectInternalError = 106
)

func main() {
	detect, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err)
		os.Exit(Error_Initialize)
	}

	if err := detect.BuildPlan.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
		os.Exit(Error_Initialize)
	}

	if code, err := d(detect); err != nil {
		detect.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func d(detect detect.Detect) (int, error) {
	metadata, ok, err := metadata.NewMetadata(detect.Application, detect.Logger)
	if err != nil {
		return detect.Error(Error_ReadMetadata), fmt.Errorf("unable to read riff metadata: %s", err.Error())
	}

	if !ok {
		return detect.Fail(), nil
	}

	detected := false

	if metadata.Override != "" {
		if metadata.Override == node.Name {
			detected = true
			detect.Logger.Debug("Override language: %q.", node.Name)
		}
	} else {
		// Try npm
		if _, ok := detect.BuildPlan[modules.Dependency]; ok {
			detected = true
		} else {
			// Try node
			if ok, err := node.DetectNode(detect, metadata); err != nil {
				detect.Logger.Info("Error trying to use node invoker: %s", err.Error())
				return detect.Error(Error_DetectInternalError), nil
			} else if ok {
				detected = true
			}
		}

		if detected {
			detect.Logger.Debug("Detected language: %q.", node.Name)
		}
	}

	if detected {
		return detect.Pass(node.BuildPlanContribution(detect, metadata))
	}

	return detect.Fail(), nil
}
