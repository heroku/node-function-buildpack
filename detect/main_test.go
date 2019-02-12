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
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/buildpack/libbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	nodeCNB "github.com/cloudfoundry/nodejs-cnb/node"
	"github.com/cloudfoundry/npm-cnb/modules"
	. "github.com/onsi/gomega"
	"github.com/projectriff/node-function-buildpack/node"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestDetect(t *testing.T) {
	spec.Run(t, "Detect", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		var f *test.DetectFactory

		it.Before(func() {
			f = test.NewDetectFactory(t)
		})

		it("fails without metadata", func() {
			g.Expect(d(f.Detect)).To(Equal(detect.FailStatusCode))
		})

		it("passes and opts in for the node-invoker if the NPM app BP applied", func() {
			f.AddBuildPlan(modules.Dependency, buildplan.Dependency{})
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `artifact = "my.js"`)

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				node.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{node.FunctionArtifact: "my.js"},
				},
			}))
		})

		it("passes and opts in for the node-invoker if the NPM app BP did not apply, but artifact is .js", func() {
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "my.js"), "module.exports = x => x**2")
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `artifact = "my.js"`)

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				node.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{node.FunctionArtifact: "my.js"},
				},
			}))
		})

		it("passes and opts in for the node-invoker if the override matches", func() {
			f.AddBuildPlan("jvm-application", buildplan.Dependency{})
			f.AddBuildPlan(modules.Dependency, buildplan.Dependency{})
			test.WriteFileWithPerm(t, filepath.Join(f.Detect.Application.Root, "fn.sh"), 0755 /*<-executable*/, "some bash")
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `artifact = "fn.sh"
override = "node"`)

			g.Expect(d(f.Detect)).To(Equal(detect.PassStatusCode))
			g.Expect(f.Output).To(Equal(buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				node.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{node.FunctionArtifact: "fn.sh"},
				},
			}))
		})

		it.Focus("fails if override is missmatched", func() {
			f.AddBuildPlan("jvm-application", buildplan.Dependency{})
			f.AddBuildPlan(modules.Dependency, buildplan.Dependency{})
			test.WriteFileWithPerm(t, filepath.Join(f.Detect.Application.Root, "fn.sh"), 0755 /*<-executable*/, "some bash")
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "riff.toml"), `artifact = "fn.js"
override = "java"`)

			g.Expect(d(f.Detect)).To(Equal(detect.FailStatusCode))
		})
	}, spec.Report(report.Terminal{}))
}
