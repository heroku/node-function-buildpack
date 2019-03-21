# `node-function-buildpack`

The Node Function Buildpack is a Cloud Native Buildpack V3 that provides the riff [Node Function Invoker](https://github.com/projectriff/node-function-invoker) to functions.

This buildpack is designed to work in collaboration with other buildpacks, which are tailored to
support (and know how to build / run) languages supported by riff.

## In Plain English

In a nutshell, when combined with the other buildpacks present in the [riff builder](https://github.com/projectriff/riff-buildpack-group) what this means (and especially when dealing with the riff CLI which takes care of the creation of the `riff.toml` file for you):

- The presence of a `package.json` file and/or the fact that the `--artifact` flag points to a `.js` file will result in
  1. the `npm installation` of the function if applicable
  2. the execution as a node function thanks to the [node invoker](https://github.com/projectriff/node-function-invoker)
- Ambiguity in the detection process will result in a build failure
- The presence of the `--invoker` flag will entirely bypass the detection mechanism and force a given language/invoker

## Detailed Buildpack Behavior

### Detection Phase

Detection passes if

- a `$APPLICATION_ROOT/riff.toml` exists and
- the build plan already contains a `npm` key (typically because an NPM based application was detected by the [npm buildpack](https://github.com/cloudfoundry/npm-cnb))
  1. alternatively, if the file pointed to by the `artifact` value in `riff.toml` exists and has a `.js` extension

If detection passes, the buildpack will add a `riff-invoker-node` key and `fn` metadata extracted from the riff metadata.

If several languages are detected simultaneously, the detect phase errors out.
The `override` key in `riff.toml` can be used to bypass detection and force the use of a particular invoker.

### Build Phase

If a node function has been detected

- Contributes the riff Node Invoker to a launch layer, set as the main `node` entry point with `FUNCTION_URI = <artifact>` set as an environment variable.
  Note that `artifact` may actually be empty, in which case the invoker will `require()` the current directory (the function), which in turn expects that it contains a valid `package.json` file with its `main` entry point set.

The function behavior is exposed _via_ standard buildpack [process types](https://github.com/buildpack/spec/blob/master/buildpack.md#launch):

- Contributes `web` process
- Contributes `function` process

## How to Build

You can build the buildpack by running

```bash
make
```

This will package (with pre-downloaded cache layers) the buildpack in the
`artifactory/io/projectriff/node/io.projectriff.node/latest` directory. That can be used as a `uri` in a `builder.toml`
file of a builder (see https://github.com/projectriff/riff-buildpack-group)

## License

This buildpack is released under version 2.0 of the [Apache License](https://www.apache.org/licenses/LICENSE-2.0).
