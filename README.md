# gotestiful

> gotest + beatiful

`gotestiful` is a go terminal program that wraps 'go test' to streamline tests running and config and also improves the coverage presentation output.

## Install

```
go install github.com/alex-parra/gotestiful@latest
```

## Examples

- `gotestiful init` creates a base configuration in the current folder  
  (the config file is optional. you may opt to use flags only)
- `gotestiful help` shows examples and flags infos
- `gotestiful` runs tests for the current folder eg. `go test ./...`
- `gotesttiful some/pkg` runs only that package eg. `go test some/pkg`
- `gotestiful -cache=false` runs tests without cache eg. `go test -count=1 ...`
- ... see `gotestiful help` for all flags

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
