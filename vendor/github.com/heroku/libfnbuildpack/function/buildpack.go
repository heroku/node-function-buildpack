/*
 * Copyright 2019 The original author or authors
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

package function

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

const (
	ErrorInitialize          = 101
	ErrorDetectReadMetadata  = 102
	ErrorDetectInternalError = 103
	ErrorBuildInternalError  = 104
)

type Buildpack interface {
	Id() string
	Detect(detect detect.Detect, metadata Metadata) (*buildplan.Plan, error)
	Build(build build.Build) error
}

func Detect(bp Buildpack) {
	d, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Detect: %s\n", err)
		os.Exit(ErrorInitialize)
	}

	if code, err := doDetect(bp, d); err != nil {
		d.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func doDetect(bp Buildpack, d detect.Detect) (int, error) {
	m, ok, err := NewMetadata(d.Application, d.Logger)
	if err != nil {
		return d.Error(ErrorDetectReadMetadata), fmt.Errorf("unable to read riff metadata: %s", err.Error())
	}

	if !ok {
		return d.Fail(), nil
	}

	if m.Override != "" && m.Override != bp.Id() {
		// targeting a different language
		return d.Fail(), nil
	}

	plan, err := bp.Detect(d, m)
	if err != nil {
		d.Logger.Info("Error trying to use %s invoker: %s", bp.Id(), err.Error())
		return d.Error(ErrorDetectInternalError), nil
	}
	if plan == nil {
		if m.Override == "" {
			// didn't detect, normal
			return d.Fail(), nil
		}
		// expected to detect, but didn't
		d.Logger.Info("Unable to detect invoker: %s", bp.Id())
		return d.Error(ErrorDetectInternalError), nil
	}

	d.Logger.Debug("Detected language: %q.", bp.Id())
	return d.Pass(*plan)
}

func Build(bp Buildpack) {
	b, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build: %s\n", err)
		os.Exit(ErrorInitialize)
	}

	if code, err := doBuild(bp, b); err != nil {
		b.Logger.Info(err.Error())
		os.Exit(code)
	} else {
		os.Exit(code)
	}
}

func doBuild(bp Buildpack, b build.Build) (int, error) {
	b.Logger.Title(b.Buildpack)

	if err := bp.Build(b); err != nil {
		return b.Failure(ErrorBuildInternalError), fmt.Errorf("unable to build invoker %q: %s", bp.Id(), err)
	}
	return b.Success(buildpackplan.Plan{})
}
