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
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "sync"
    "time"

    "github.com/buildpack/libbuildpack/application"
    "github.com/buildpack/libbuildpack/buildplan"
    "github.com/cloudfoundry/libcfbuildpack/build"
    "github.com/cloudfoundry/libcfbuildpack/detect"
    "github.com/cloudfoundry/libcfbuildpack/helper"
    "github.com/cloudfoundry/libcfbuildpack/layers"
    "github.com/heroku/libfnbuildpack/function"
)

const (
    // Name is a short, human friendly name for the node invoker
    Name = "node"

    // Dependency is a key identifying the node invoker dependency in the build plan.
    Dependency = "riff-invoker-node"
)

// RiffNodeInvoker represents the Node invoker contributed by the buildpack.
type RiffNodeInvoker struct {
    // A reference to the user function source tree.
    application application.Application

    // The file in the function tree that is the entrypoint.
    // May be empty, in which case the function is require()d as a node module.
    functionJS string

    // Provides access to the launch layers, used to craft the process commands.
    layers layers.Layers

    // A dedicated layer for the node invoker itself. Cacheable once npm-installed
    invokerLayer layers.DependencyLayer

    // A dedicated layer for the function location. Not cacheable, as it changes with the value of functionJS.
    functionLayer layers.Layer
}

func BuildPlanContribution(d detect.Detect, m function.Metadata) buildplan.Plan {
    return buildplan.Plan{}
}

// Contribute expands the node invoker tgz and creates launch configurations that run "node server.js"
func (r RiffNodeInvoker) Contribute(wg *sync.WaitGroup) error {
    defer wg.Done()

    defer func(startTime time.Time)() {
        endTime := time.Now()
        output := fmt.Sprintf("****** Total Execution Time Riff Invoker Contribute: %d (ms)", endTime.Sub(startTime)/time.Millisecond)
        fmt.Println(output)
    }(time.Now())

    if err := r.invokerLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
        layer.Logger.Body("Expanding to %s", layer.Root)
        if e := helper.ExtractTarGz(artifact, layer.Root, 1); e != nil {
            return e
        }
        layer.Logger.Body("npm-installing the node invoker")
        cmd := exec.Command("npm", "install", "--production")
        cmd.Stdout = os.Stderr
        cmd.Stderr = os.Stderr
        cmd.Dir = layer.Root
        if e := cmd.Run(); e != nil {
            return e
        }

        return nil
    }, layers.Launch); err != nil {
        return err
    }

    // Only need to set the ENV var to point to the user function location
    if err := r.functionLayer.Contribute(marker{"NodeJS", r.functionJS}, func(layer layers.Layer) error {
        return layer.OverrideLaunchEnv("USER_FUNCTION_URI", filepath.Join(r.application.Root, r.functionJS))
    }, layers.Launch); err != nil {
        return err
    }

    command := fmt.Sprintf(`node %s/server.js`, r.invokerLayer.Root)

    return r.layers.WriteApplicationMetadata(layers.Metadata{
        Processes: layers.Processes{
            layers.Process{Type: "web", Command: command},
            layers.Process{Type: "function", Command: command},
        },
    })
}

func NewNodeInvoker(build build.Build) (RiffNodeInvoker, bool, error) {
    deps, err := build.Buildpack.Dependencies()
    if err != nil {
        return RiffNodeInvoker{}, false, err
    }

    dep, err := deps.Best(Dependency, "0.1.3", build.Stack)
    if err != nil {
        return RiffNodeInvoker{}, false, err
    }

    return RiffNodeInvoker{
        application:   build.Application,
        functionJS:    "",
        layers:        build.Layers,
        invokerLayer:  build.Layers.DependencyLayer(dep),
        functionLayer: build.Layers.Layer("function"),
    }, true, nil
}

type marker struct {
    Language string `toml:"language"`
    Function string `toml:"function"`
}

func (m marker) Identity() (string, string) {
    return m.Language, m.Function
}
