package node

import (
    "log"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/heroku/libhkbuildpack/layers"
)

type MiddlewareFunctionB struct {
    Path  string `toml:"path"`
    Layer layers.Layer
}

func NewMiddlewareFunctionB(l layers.Layer, path string) MiddlewareFunctionB {
    return MiddlewareFunctionB{
        Path:  path,
        Layer: l,
    }
}

func (f MiddlewareFunctionB) Contribute() error {
    f.Layer.Touch()

    if err := f.Layer.WriteMetadata(f, layers.Launch); err != nil {
        return err
    }

    if err := os.MkdirAll(f.Layer.Root, 0755); err != nil {
        return err
    }

    cmd := exec.Command("git", "clone", "https://github.com/heroku/node-function-middleware.git")
    cmd.Dir = filepath.Join(f.Layer.Root)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    if err != nil {
        log.Fatalf("cmd.Run() failed with %s\n", err)
    }

    middlewarePath := filepath.Join(f.Layer.Root, "node-function-middleware")
    middlewareFunc := filepath.Join(middlewarePath, "index.js")
    if err := f.Layer.AppendPathLaunchEnv("MIDDLEWARE_FUNCTION_URI", middlewareFunc); err != nil {
        return err
    }

    cmd = exec.Command("npm", "install", "--production")
    cmd.Stdout = os.Stderr
    cmd.Stderr = os.Stderr
    cmd.Dir = middlewarePath
    if e := cmd.Run(); e != nil {
        return e
    }

    return nil
}
