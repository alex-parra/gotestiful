package internal

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
)

type RunTestsOpts struct {
	TestPath         string
	FlagColor        bool
	FlagCache        bool
	FlagCover        bool
	FlagCoverReport  bool
	FlagCoverProfile string
	FlagVerbose      bool
	FlagListIgnored  bool
	FlagSkipEmpty    bool
	FlagListEmpty    bool
	Excludes         []string
}

func RunTests(opts RunTestsOpts) {
	color.NoColor = !opts.FlagColor
	coverProfile := zvfb(opts.FlagCoverProfile, "./coverage.out")

	// Get list of all packages in the test path
	allPkgsStr := shCmd("go", shArgs{"list", opts.TestPath}, "")
	allPkgs := splitLines(allPkgsStr)

	// Exclude by package prefix
	testPkgs := allPkgs
	if len(opts.Excludes) > 0 {
		testPkgsStr := shCmd("grep", shArgs{"-Ev", getExcludePattern(opts.Excludes)}, allPkgsStr)
		testPkgs = splitLines(testPkgsStr)
	}

	ignoredPkgs := sliceExclude(allPkgs, testPkgs)

	// function to inject that actually "prints" each line
	lineOut := func(str ...string) { fmt.Println(strings.Join(str, " ")) }

	// channel to receive each 'go test' stdout line
	goTestOutput := make(chan string)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		processOutput(&processOutputParams{
			OutputChannel:   goTestOutput,
			LineOut:         lineOut,
			ToTestPackages:  testPkgs,
			IgnoredPackages: ignoredPkgs,
			FlagVerbose:     opts.FlagVerbose,
			FlagSkipEmpty:   opts.FlagSkipEmpty,
			FlagListEmpty:   opts.FlagListEmpty,
			FlagListIgnored: opts.FlagListIgnored,
			IndentSpaces:    2,
		})
		wg.Done()
	}()

	// Compose and run 'go test ...'
	lineOut(sf("\nTesting %d packages in '%s'\n", len(testPkgs), opts.TestPath))
	testArgs := shArgs{"test"}
	testArgs = sliceAppendIf(opts.FlagVerbose, testArgs, "-v")
	testArgs = sliceAppendIf(!opts.FlagCache, testArgs, "-count=1")
	testArgs = sliceAppendIf(opts.FlagCover, testArgs, "-cover")
	testArgs = sliceAppendIf(opts.FlagCoverProfile != "" || opts.FlagCoverReport, testArgs, "-coverprofile="+coverProfile)
	testArgs = append(testArgs, testPkgs...)
	shPipe("go", testArgs, "", goTestOutput)
	wg.Wait()

	// Open html coverage report
	if opts.FlagCoverReport {
		shCmd("go", shArgs{"tool", "cover", "-html=" + coverProfile}, "")
	}
}

func getExcludePattern(pkgs []string) string {
	pkgExcludePattern := ""
	for _, path := range pkgs {
		if pkgExcludePattern != "" {
			pkgExcludePattern = pkgExcludePattern + "|"
		}
		pkgExcludePattern += "^" + path
	}
	return pkgExcludePattern
}
