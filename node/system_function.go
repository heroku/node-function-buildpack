package node

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/layers"
)

type SystemFunction struct {
	Path string `toml:"path"`
	Layer layers.Layer
}

func NewSystemFunction(l layers.Layer) (SystemFunction, error) {
	// TODO push this up into buildpack.go
	buildpackDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return SystemFunction{}, err
	}

	return SystemFunction{
		Path: filepath.Join(buildpackDir, "../lib/system.js"),
		Layer: l,
	}, nil
}

func (f SystemFunction) Contribute() error {
	if err := f.Layer.WriteMetadata(f, layers.Launch); err != nil {
		return err
	}

	jsFile, err := ioutil.ReadFile(filepath.Join(f.Path))
	if err != nil {
		return err
	}

	destFile := filepath.Join(f.Layer.Root, "system.js")
	if err := f.Layer.OverrideLaunchEnv("FUNCTION_URI", destFile); err != nil {
		return err
	}

	if err = ioutil.WriteFile(destFile, jsFile, 0644); err != nil {
		return err
	}

	return nil
}