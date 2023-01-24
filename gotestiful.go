/*
`gotestiful` wraps 'go test' to streamline running tests while also giving you a pleasant output presentation.

# Install

	go install github.com/alex-parra/gotestiful@latest

# Examples

	`gotestiful init`
	- creates a base configuration in the current folder (the config file is optional. you may opt to use flags only)

	`gotestiful help`
	- shows examples and flags infos

	`gotestiful`
	- runs tests for the current folder eg. `go test ./...`

	`gotesttiful some/pkg`
	- runs only that package eg. `go test some/pkg`

	`gotestiful -cache=false`
	- runs tests without cache eg. `go test -count=1 ...`

	... see `gotestiful help` for all flags

# Features:

	config file per project [optional]
	- run `gotestiful init` from your project root to create a `.gotestiful` config file and then adjust the settings. afterwards you only need to run `gotestiful` and the config is read

	exclusion list
	- add packages (or just prefixes) to the config `exclude` array to not test those packages eg. exclude generated code such as protobuf packages

	global coverage summary
	- shows the overall code coverage calculated from the coverage score of each tested package.

	open html coverage detail report
	- set the `-report` flag and the coverage html detail will open (eg. `go tool cover -html`)
*/
package main

import (
	"flag"

	gtf "github.com/alex-parra/gotestiful/internal"
)

const version = "v0.1.4"

func main() {
	conf := gtf.GetConfig()

	flagVersion := flag.Bool("version", false, "Gotestiful version: print version information")
	flagColor := flag.Bool("color", conf.Color, "Colorize output: turn colorized output on/off")
	flagCache := flag.Bool("cache", conf.Cache, "Test caching: tests cache on/off eg. 'go test -count=1' if false")
	flagCover := flag.Bool("cover", conf.Cover, "Coverage: turn coverage reporting on/off eg. 'go test -cover'")
	flagCoverReport := flag.Bool("report", conf.Report, "Coverage details: open html coverage report eg. 'go tool cover -html'")
	flagCoverProfile := flag.String("coverprofile", conf.CoverProfile, "Coverage profile: coverage report output file path (default ./coverage.out)")
	flagVerbose := flag.Bool("v", conf.Verbose, "Verbose output: run tests with verbose output eg. 'go test -v'")
	flagListIgnored := flag.Bool("listignored", conf.ListIgnored, "Excluded packages: list ignored packages (at the end)")
	flagSkipEmpty := flag.Bool("skipempty", conf.SkipEmpty, "No tests omit: do not show packages with no tests in the output (affects coverage)")
	flagListEmpty := flag.Bool("listempty", conf.ListEmpty, "No tests list: list packages with no tests (at the end)")
	flag.Usage = gtf.PrintHelp
	flag.Parse()

	testPath := flag.Arg(0)
	if testPath == "" {
		testPath = "./..."
	}

	switch {
	case *flagVersion:
		gtf.PrintVersion(version)

	case testPath == "init":
		gtf.InitConfig()

	default:
		gtf.RunTests(gtf.RunTestsOpts{
			TestPath:         testPath,
			FlagColor:        *flagColor,
			FlagCache:        *flagCache,
			FlagCover:        *flagCover,
			FlagCoverReport:  *flagCoverReport,
			FlagCoverProfile: *flagCoverProfile,
			FlagVerbose:      *flagVerbose,
			FlagListIgnored:  *flagListIgnored,
			FlagSkipEmpty:    *flagSkipEmpty,
			FlagListEmpty:    *flagListEmpty,
			Excludes:         conf.Exclude,
		})
	}

}
