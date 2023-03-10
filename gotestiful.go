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
	"errors"
	"flag"
	"log"
	"os"

	gtf "github.com/alex-parra/gotestiful/internal"
)

const version = "v1.1.2"

func main() {
	conf, err := gtf.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	flagVersion := flag.Bool("version", false, "Gotestiful version: print version information")
	flagColor := flag.Bool("color", conf.Color, "Colorize output: turn colorized output on/off")
	flagCache := flag.Bool("cache", conf.Cache, "Test caching: tests cache on/off eg. 'go test -count=1' if false")
	flagCover := flag.Bool("cover", conf.Cover, "Coverage: turn coverage reporting on/off eg. 'go test -cover'")
	flagCoverReport := flag.Bool("report", conf.Report, "Coverage details: open html coverage report eg. 'go tool cover -html'")
	flagCoverProfile := flag.String("coverprofile", conf.CoverProfile, "Coverage profile: coverage report output file path (default ./coverage.out). Takes longer (disables caching).")
	flagVerbose := flag.Bool("v", conf.Verbose, "Verbose output: run tests with verbose output eg. 'go test -v'")
	flagListIgnored := flag.Bool("listignored", conf.ListIgnored, "Excluded packages: list ignored packages (at the end)")
	flagSkipEmpty := flag.Bool("skipempty", conf.SkipEmpty, "No tests omit: do not show packages with no tests in the output (affects coverage)")
	flagListEmpty := flag.Bool("listempty", conf.ListEmpty, "No tests list: list packages with no tests (at the end)")
	flagFullCoverage := flag.Bool("fullCoverage", conf.FullCoverage, "Count overall coverage including packages without tests. Takes longer (disables caching).")
	flagTestOutput := flag.String("testoutput", conf.TestOutput, "Print JSON output of go test to the given file. Output format is same as go test with -json flag")

	flagAzureDevopsURL := flag.String("azureDevopsURL", "", "Add azure devops URL to send a request with comment to")
	flagAzureDevopsAuthToken := flag.String("azureDevopsAuthToken", "", "Add azure devops auth token to send a request with comment to")

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
		err := gtf.InitConfig()
		if err != nil {
			log.Fatal(err)
		}

	default:
		err := gtf.RunTests(gtf.RunTestsOpts{
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
			FlagFullCoverage: *flagFullCoverage,
			Excludes:         conf.Exclude,
			FlagTestOutput:   *flagTestOutput,

			Azure: gtf.AzureConf{
				URL:  *flagAzureDevopsURL,
				Auth: *flagAzureDevopsAuthToken,
			},
		})

		switch {
		case errors.Is(err, gtf.ErrTestRunIgnore):
			os.Exit(1) // Known error due to tests failing. No need to log.
		case err != nil:
			log.Fatal(err)
		}
	}

}
