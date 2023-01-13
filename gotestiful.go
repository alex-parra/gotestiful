package main

import (
	"flag"

	gtf "github.com/alex-parra/gotestiful/internal"
)

const version = "v0.1.1"

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
