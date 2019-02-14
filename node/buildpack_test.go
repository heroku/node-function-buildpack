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
	"github.com/projectriff/riff-buildpack/invoker"
	"github.com/projectriff/riff-buildpack/metadata"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestName(t *testing.T) {
	spec.Run(t, "Name", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		it("has the right name", func() {
			b := node.NewBuildpack()

			g.Expect(b.Name()).To(Equal("node"))
		})
	}, spec.Report(report.Terminal{}))
}

func TestDetect(t *testing.T) {
	spec.Run(t, "Detect", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.DetectFactory
		var m metadata.Metadata
		var b invoker.Buildpack

		it.Before(func() {
			f = test.NewDetectFactory(t)
			m = metadata.Metadata{}
			b = node.NewBuildpack()
		})

		it("fails by default", func() {
			detected, err := b.Detect(f.Detect, m)

			g.Expect(detected).To(BeFalse())
			g.Expect(err).To(BeNil())
		})

		it("passes if the NPM app BP applied", func() {
			f.AddBuildPlan(modules.Dependency, buildplan.Dependency{})

			detected, err := b.Detect(f.Detect, m)

			g.Expect(detected).To(BeTrue())
			g.Expect(err).To(BeNil())
		})

		it("passes if the NPM app BP did not apply, but artifact is .js", func() {
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "my.js"), "module.exports = x => x**2")
			m.Artifact = "my.js"

			detected, err := b.Detect(f.Detect, m)

			g.Expect(detected).To(BeTrue())
			g.Expect(err).To(BeNil())
		})
	}, spec.Report(report.Terminal{}))
}

func TestBuildPlan(t *testing.T) {
	spec.Run(t, "BuildPlan", func(t *testing.T, _ spec.G, it spec.S) {
		g := NewGomegaWithT(t)

		var f *test.DetectFactory
		var m metadata.Metadata
		var b invoker.Buildpack

		it.Before(func() {
			f = test.NewDetectFactory(t)
			m = metadata.Metadata{}
			b = node.NewBuildpack()
		})

		it("creates a buildplan for a package.json detection", func() {
			f.AddBuildPlan(modules.Dependency, buildplan.Dependency{})

			plan := b.BuildPlan(f.Detect, m)

			g.Expect(plan).To(Equal(buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				node.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{node.FunctionArtifact: ""},
				},
			}))
		})

		it("creates a buildplan for .js artifact", func() {
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "my.js"), "module.exports = x => x**2")
			m.Artifact = "my.js"

			plan := b.BuildPlan(f.Detect, m)

			g.Expect(plan).To(Equal(buildplan.BuildPlan{
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

func TestInvoker(t *testing.T) {
	spec.Run(t, "Invoker", func(t *testing.T, _ spec.G, it spec.S) {
		g := NewGomegaWithT(t)

		var f *test.BuildFactory
		var b invoker.Buildpack

		it.Before(func() {
			f = test.NewBuildFactory(t)
			b = node.NewBuildpack()
		})

		it("won't build unless passed detection", func() {
			_, ok, err := b.Invoker(f.Build)

			g.Expect(ok).To(BeFalse())
			g.Expect(err).To(BeNil())
		})

		it.Pend("will build if passed detection", func() {
			// TODO configure state from the buildplan
			f.AddBuildPlan(node.Dependency, buildplan.Dependency{})
			i, ok, err := b.Invoker(f.Build)

			g.Expect(ok).To(BeTrue())
			g.Expect(err).To(BeNil())
			g.Expect(i).To(Equal(node.RiffNodeInvoker{
				// TODO set actual values
			}))
		})
	}, spec.Report(report.Terminal{}))
}
