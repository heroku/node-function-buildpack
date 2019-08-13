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

package node

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/npm-cnb/modules"
	"github.com/heroku/libfnbuildpack/function"
	"github.com/heroku/libhkbuildpack/build"
	"github.com/heroku/libhkbuildpack/detect"
)

type NodeBuildpack struct {
	id string
}

func (bp *NodeBuildpack) Id() string {
	return bp.id
}

func (bp *NodeBuildpack) Detect(d detect.Detect, m function.Metadata) (*buildplan.BuildPlan, error) {
	if detected, err := bp.detect(d); err != nil {
		return nil, err
	} else if detected {
		plan := BuildPlanContribution(d, m)
		return &plan, nil
	}
	// didn't detect
	return nil, nil
}

func (*NodeBuildpack) detect(d detect.Detect) (bool, error) {
	// Try npm
	if _, ok := d.BuildPlan[modules.Dependency]; ok {
		return true, nil
	}
	// Try node
	return DetectNode(d)
}

func (*NodeBuildpack) Build(b build.Build) error {
	invoker, ok, err := NewNodeInvoker(b)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("buildpack passed detection but did not know how to actually build")
	}
	if err := invoker.Contribute(); err != nil {
		return err
	}

	middlewareLayerA := b.Layers.Layer("middlewareA")
	middlewareLayerB := b.Layers.Layer("middlewareB")

	systemLayer := b.Layers.Layer("system")

	bpBinDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	bpDir := filepath.Join(bpBinDir, "../")

	middlewareFuncA := NewMiddlewareFunction(middlewareLayerA, bpDir)
	if err := middlewareFuncA.Contribute(); err != nil {
		return err
	}

	middlewareFuncB := NewMiddlewareFunction(middlewareLayerB, bpDir)
	if err := middlewareFuncB.Contribute(); err != nil {
		return err
	}

	sysFunc := NewSystemFunction(systemLayer, bpDir)
	if err := sysFunc.Contribute(); err != nil {
		return err
	}

	return nil
}

func NewBuildpack() function.Buildpack {
	return &NodeBuildpack{
		id: "node",
	}
}
