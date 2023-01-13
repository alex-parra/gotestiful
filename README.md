# gotestiful

> gotest + beatiful

`gotestiful` is a go terminal program that wraps 'go test' to streamline tests running and config and also improves the coverage presentation output.

## Quick start

1. run `go install github.com/alex-parra/gotestiful@latest`
2. run `gotestiful` from the root of any go project (where go.mod is)

## Examples

- `gotestiful` runs tests for the current folder eg. `go test ./...`
- `gotestiful -help` shows examples and flags infos
- `gotesttiful some/pkg` runs only that package eg. `go test some/pkg`
- `gotestiful -cache=false` runs tests without cache eg. `go test -count=1 ...`
- `gotestiful init` creates a base configuration in the current folder  
  (the config file is optional. you may opt to use flags only)
- ... see `gotestiful -help` for all flags

## Features:

- **config file per project**  
  run `gotestiful init` from your project root to create a `.gotestiful` config file and then adjust the settings.  
  afterwards you only need to run `gotestiful` and the config is read

- **exclusion list**  
  add packages (or just prefixes) to the config `exclude` array to not test those packages.  
  example: exclude generated code such as protobuf packages

- **global coverage summary**  
  shows the overall code coverage calculated from the coverage score of each tested package.

- **open html coverage detail report**  
  set the `-report` flag and the coverage html detail will open (eg. `go tool cover -html`)

## Contributors

<a href="https://github.com/Tanu-N-Prabhu/Python/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=alex-parra/gotestiful"/>
</a>

<small>Made with [contributors-img](https://contrib.rocks).</small>
