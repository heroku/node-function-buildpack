/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package node_test

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	nodeCNB "github.com/cloudfoundry/nodejs-cnb/node"
	"github.com/cloudfoundry/npm-cnb/modules"
	. "github.com/onsi/gomega"
	"github.com/projectriff/node-function-buildpack/node"
	"github.com/projectriff/riff-buildpack/function"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestName(t *testing.T) {
	spec.Run(t, "Id", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		it("has the right id", func() {
			b := node.NewBuildpack()

			g.Expect(b.Id()).To(Equal("node"))
		})
	}, spec.Report(report.Terminal{}))
}

func TestDetect(t *testing.T) {
	spec.Run(t, "Detect", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.DetectFactory
		var m function.Metadata
		var b function.Buildpack

		it.Before(func() {
			f = test.NewDetectFactory(t)
			m = function.Metadata{}
			b = node.NewBuildpack()
		})

		it("fails by default", func() {
			plan, err := b.Detect(f.Detect, m)

			g.Expect(err).To(BeNil())
			g.Expect(plan).To(BeNil())
		})

		it("passes if the NPM app BP applied", func() {
			f.AddBuildPlan(modules.Dependency, buildplan.Dependency{})

			plan, err := b.Detect(f.Detect, m)

			g.Expect(err).To(BeNil())
			g.Expect(plan).To(Equal(&buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				node.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{node.FunctionArtifact: ""},
				},
			}))
		})

		it("passes if the NPM app BP did not apply, but artifact is .js", func() {
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "my.js"), "module.exports = x => x**2")
			m.Artifact = "my.js"

			plan, err := b.Detect(f.Detect, m)

			g.Expect(err).To(BeNil())
			g.Expect(plan).To(Equal(&buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				node.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{node.FunctionArtifact: "my.js"},
				},
			}))
		})
	}, spec.Report(report.Terminal{}))
}

func TestBuild(t *testing.T) {
	spec.Run(t, "Build", func(t *testing.T, _ spec.G, it spec.S) {
		g := NewGomegaWithT(t)

		var f *test.BuildFactory
		var b function.Buildpack

		it.Before(func() {
			f = test.NewBuildFactory(t)
			b = node.NewBuildpack()
		})

		it("won't build unless passed detection", func() {
			err := b.Build(f.Build)

			g.Expect(err).To(MatchError("buildpack passed detection but did not know how to actually build"))
		})

		it.Pend("will build if passed detection", func() {
			f.AddBuildPlan(node.Dependency, buildplan.Dependency{})
			f.AddDependency(node.Dependency, ".")

			err := b.Build(f.Build)

			g.Expect(err).To(BeNil())
		})
	}, spec.Report(report.Terminal{}))
}
