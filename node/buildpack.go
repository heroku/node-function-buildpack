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
	"sync"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/heroku/libfnbuildpack/function"
)

type NodeBuildpack struct {
	id string
}

func (bp *NodeBuildpack) Id() string {
	return bp.id
}

func (bp *NodeBuildpack) Detect(d detect.Detect, m function.Metadata) (*buildplan.Plan, error) {
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
	return DetectNode(d)
}

func (*NodeBuildpack) Build(b build.Build) error {
	var wg sync.WaitGroup
	wg.Add(3)

	invoker, ok, err := NewNodeInvoker(b)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("buildpack passed detection but did not know how to actually build")
	}
	go invoker.Contribute(&wg)
	//if err := invoker.Contribute(); err != nil {
	//	return err
	//}

	sysFunc, ok, err := NewSystemFunction(b)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("buildpack passed detection but did not know how to actually build")
	}
	go sysFunc.Contribute(&wg)
	//if err := sysFunc.Contribute(); err != nil {
	//	return err
	//}

	middlewareFunc, ok, err := NewMiddlewareFunction(b)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("buildpack passed detection but did not know how to actually build")
	}
	go middlewareFunc.Contribute(&wg)
	//if err := middlewareFunc.Contribute(); err != nil {
	//	return err
	//}

	wg.Wait()

	return nil
}

func NewBuildpack() function.Buildpack {
	return &NodeBuildpack{
		id: "node",
	}
}
