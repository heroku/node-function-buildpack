package node

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

type SystemFunction struct {
	systemFunctionLayer layers.DependencyLayer
}

func NewSystemFunction(build build.Build) (SystemFunction, bool, error) {
	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return SystemFunction{}, false, err
	}

	dep, err := deps.Best("system-function", "0.4.3", build.Stack)
	if err != nil {
		return SystemFunction{}, false, err
	}

	return SystemFunction{
		systemFunctionLayer: build.Layers.DependencyLayer(dep),
	}, true, nil
}

func (f SystemFunction) Contribute(wg *sync.WaitGroup) error {
	defer wg.Done()

	defer func(startTime time.Time)() {
		endTime := time.Now()
		output := fmt.Sprintf("****** Total Execution Time System Function Contribute: %d (ms)", endTime.Sub(startTime)/time.Millisecond)
		fmt.Println(output)
	}(time.Now())

	if err := f.systemFunctionLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Expanding to %s", layer.Root)
		if e := helper.ExtractTarGz(artifact, layer.Root, 1); e != nil {
			return e
		}
		layer.Logger.Body("npm-installing the system function")
		cmd := exec.Command("npm", "install", "--production")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Dir = filepath.Join(layer.Root, "src")
		if e := cmd.Run(); e != nil {
			return e
		}

		return nil
	}, layers.Launch); err != nil {
		return err
	}

	systemFuncPath := filepath.Join(f.systemFunctionLayer.Root, "src", "index.js")
	fmt.Println(fmt.Sprintf("FUNCTION_URI = %s", systemFuncPath))
	if err := f.systemFunctionLayer.OverrideLaunchEnv("FUNCTION_URI", systemFuncPath); err != nil {
		return err
	}

	return nil
}
