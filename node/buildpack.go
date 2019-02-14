/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/npm-cnb/modules"
	"github.com/projectriff/riff-buildpack/invoker"
	"github.com/projectriff/riff-buildpack/metadata"
)

type NodeBuildpack struct {
	name string
}

func (b *NodeBuildpack) Name() string {
	return b.name
}

func (b *NodeBuildpack) Detect(detect detect.Detect, metadata metadata.Metadata) (bool, error) {
	// Try npm
	if _, ok := detect.BuildPlan[modules.Dependency]; ok {
		return true, nil
	}
	// Try node
	return DetectNode(detect, metadata)
}

func (b *NodeBuildpack) BuildPlan(detect detect.Detect, metadata metadata.Metadata) buildplan.BuildPlan {
	return BuildPlanContribution(detect, metadata)
}

func (b *NodeBuildpack) Invoker(build build.Build) (invoker.Invoker, bool, error) {
	return NewNodeInvoker(build)
}

func NewBuildpack() invoker.Buildpack {
	return &NodeBuildpack{
		name: "node",
	}
}
