# function-usages
[![CI](https://github.com/erikgb/function-usages/actions/workflows/ci.yml/badge.svg)](https://github.com/erikgb/function-usages/actions/workflows/ci.yml)

This is an **EXPERIMENTAL** Crossplane composition function to simplify managing Crossplane [usages to control deletion ordering][usages] in Crossplane compositions.

## Development

To learn how to work with Crossplane functions:

* [Follow the guide to writing a composition function in Go][function guide]
* [Learn about how composition functions work][functions]
* [Read the function-sdk-go package documentation][package docs]

This function uses [Go][go], [Docker][docker], and the [Crossplane CLI][cli].

```shell
# Run code generation - see input/generate.go
$ go generate ./...

# Run tests - see fn_test.go
$ go test ./...

# Build the function's runtime image - see Dockerfile
$ docker build . --tag=runtime

# Build a function package - see package/crossplane.yaml
$ crossplane xpkg build -f package --embed-runtime-image=runtime
```

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
[go]: https://go.dev
[function guide]: https://docs.crossplane.io/knowledge-base/guides/write-a-composition-function-in-go
[package docs]: https://pkg.go.dev/github.com/crossplane/function-sdk-go
[docker]: https://www.docker.com
[cli]: https://docs.crossplane.io/latest/cli
[usages]: https://docs.crossplane.io/latest/managed-resources/usages/#usage-for-deletion-ordering
