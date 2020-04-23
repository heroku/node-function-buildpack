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

type MiddlewareFunction struct {
	middlewareFunctionLayer layers.DependencyLayer
}

func NewMiddlewareFunction(build build.Build) (MiddlewareFunction, bool, error) {
	deps, err := build.Buildpack.Dependencies()
	if err != nil {
		return MiddlewareFunction{}, false, err
	}

	dep, err := deps.Best("middleware-function", "1.1.2", build.Stack)
	if err != nil {
		return MiddlewareFunction{}, false, err
	}

	return MiddlewareFunction{
		middlewareFunctionLayer: build.Layers.DependencyLayer(dep),
	}, true, nil
}

func (f MiddlewareFunction) Contribute(wg *sync.WaitGroup) error {
	defer wg.Done()

	defer func(startTime time.Time)() {
		endTime := time.Now()
		output := fmt.Sprintf("****** Total Execution Time Middleware Function Contribute: %d (ms)", endTime.Sub(startTime)/time.Millisecond)
		fmt.Println(output)
	}(time.Now())

	if err := f.middlewareFunctionLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Expanding to %s", layer.Root)
		if e := helper.ExtractTarGz(artifact, layer.Root, 1); e != nil {
			return e
		}
		layer.Logger.Body("npm-installing the middleware function")

		cmd := exec.Command("npm", "install")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Dir = filepath.Join(layer.Root, "middleware")
		if e := cmd.Run(); e != nil {
			return e
		}

		cmd = exec.Command("npm", "run", "build")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Dir = filepath.Join(layer.Root, "middleware")
		if e := cmd.Run(); e != nil {
			return e
		}

		return nil
	}, layers.Launch); err != nil {
		return err
	}

	middlewareFuncPath := filepath.Join(f.middlewareFunctionLayer.Root, "middleware", "dist", "index.js")
	fmt.Println(fmt.Sprintf("MIDDLEWARE_FUNCTION_URI = %s", middlewareFuncPath))
	if err := f.middlewareFunctionLayer.OverrideLaunchEnv("MIDDLEWARE_FUNCTION_URI", middlewareFuncPath); err != nil {
		return err
	}

	return nil
}
