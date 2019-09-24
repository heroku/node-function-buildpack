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

package node

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	nodeCNB "github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/heroku/libfnbuildpack/function"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestName(t *testing.T) {
	spec.Run(t, "Id", func(t *testing.T, _ spec.G, it spec.S) {

		g := NewGomegaWithT(t)

		it("has the right id", func() {
			b := NewBuildpack()

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
			b = NewBuildpack()
		})

		it.After(func() {
			path := filepath.Join(filepath.Join(f.Detect.Application.Root))
			files, _ := ioutil.ReadDir(path)
			for _, file := range files {
				deleteFile(t, filepath.Join(path, file.Name()))
			}
		})

		it("should fail if no files exits", func() {
			plan, err := b.Detect(f.Detect, m)

			g.Expect(plan).To(BeNil())
			g.Expect(err).ToNot(BeNil())
		})

		it("should fail if no package.json file exits", func() {
			plan, err := b.Detect(f.Detect, m)
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "square.js"), "module.exports = x => x**2")

			g.Expect(plan).To(BeNil())
			g.Expect(err).ToNot(BeNil())
		})

		it("should fail if package.json main field doesn't correspond to a file", func() {
			packageDotJson := `{
							    "name": "fixture",
							    "version": "1.0.0",
							    "main": "hello.js",
							    "dependencies": {
							    }
			                  }`

			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "package.json"), packageDotJson)
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "square.js"), "module.exports = x => x**2")

			plan, err := b.Detect(f.Detect, m)

			g.Expect(plan).To(BeNil())
			g.Expect(err).ToNot(BeNil())
		})

		it("should fail if package.json main field doesn't correspond to a Javascript file", func() {
			packageDotJson := `{
							    "name": "fixture",
							    "version": "1.0.0",
							    "main": "hello.txt",
							    "dependencies": {
							    }
			                  }`

			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "package.json"), packageDotJson)
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "hello.txt"), "Hello, World!")

			plan, err := b.Detect(f.Detect, m)

			g.Expect(plan).To(BeNil())
			g.Expect(err).ToNot(BeNil())
		})

		it("should enforce that package.json main field corresponds to an actual file", func() {
			packageDotJson := `{
							    "name": "fixture",
							    "version": "1.0.0",
							    "main": "square.js",
							    "dependencies": {
							    }
			                  }`

			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "package.json"), packageDotJson)
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "square.js"), "module.exports = x => x**2")

			plan, err := b.Detect(f.Detect, m)

			g.Expect(err).To(BeNil())
			g.Expect(plan).To(Equal(&buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{FunctionArtifact: ""},
				},
			}))
		})

		it("should enforce that typescript files work", func() {
			packageDotJson := `{
							    "name": "fixture",
							    "version": "1.0.0",
							    "main": "dist/square.js",
							    "dependencies": {
							    }
			                  }`

			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "package.json"), packageDotJson)
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "square.ts"), packageDotJson)
			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, "dist", "square.js"), "module.exports = x => x**2")

			plan, err := b.Detect(f.Detect, m)

			g.Expect(err).To(BeNil())
			g.Expect(plan).To(Equal(&buildplan.BuildPlan{
				nodeCNB.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"launch": true, "build": true},
				},
				Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{FunctionArtifact: ""},
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
			b = NewBuildpack()
		})

		it("won't build unless passed detection", func() {
			err := b.Build(f.Build)

			g.Expect(err).To(MatchError("buildpack passed detection but did not know how to actually build"))
		})

		it.Pend("will build if passed detection", func() {
			plan := buildplan.Plan{}
			plan.Requires = append(plan.Requires, buildplan.Required{Name: Dependency})

			f.AddDependency(Dependency, ".")

			err := b.Build(f.Build)

			g.Expect(err).To(BeNil())
		})
	}, spec.Report(report.Terminal{}))
}

func deleteFile(t *testing.T, filename string) {
	t.Helper()

	if err := os.RemoveAll(filename); err != nil {
		t.Fatal(err)
	}
}
