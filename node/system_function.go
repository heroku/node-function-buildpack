package node

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/buildpack/libbuildpack/layers"
)

type SystemFunction struct {
	Path  string `toml:"path"`
	Layer layers.Layer
}

func NewSystemFunction(l layers.Layer) (SystemFunction, error) {
	// TODO push this up into buildpack.go
	buildpackDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return SystemFunction{}, err
	}

	return SystemFunction{
		Path:  filepath.Join(buildpackDir, "../system"),
		Layer: l,
	}, nil
}

func (f SystemFunction) Contribute() error {
	if err := f.Layer.WriteMetadata(f, layers.Launch); err != nil {
		return err
	}

	filenames := []string{"index.js", "package.json"}
	for _, filename := range filenames {
		file, err := ioutil.ReadFile(filepath.Join(filename))
		if err != nil {
			return err
		}

		destFile := filepath.Join(f.Layer.Root, filename)
		err = ioutil.WriteFile(destFile, file, 644)
		if err != nil {
			return err
		}
	}

	systemFunc := filepath.Join(f.Layer.Root, "index.js")
	if err := f.Layer.OverrideLaunchEnv("FUNCTION_URI", systemFunc); err != nil {
		return err
	}

	cmd := exec.Command("npm", "install", "--production")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Dir = f.Layer.Root
	if e := cmd.Run(); e != nil {
		return e
	}

	return nil
}
