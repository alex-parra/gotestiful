package internal

import (
	"fmt"
	"strings"
	"sync"
	"time"

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

type TestEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

func RunTests(opts RunTestsOpts) error {
	color.NoColor = !opts.FlagColor
	coverProfile := zvfb(opts.FlagCoverProfile, "./coverage.out")

	// Get list of all packages in the test path
	allPkgsStr, err := shCmd("go", shArgs{"list", opts.TestPath}, "")
	if err != nil {
		return err
	}
	allPkgs := splitLines(allPkgsStr)

	testPkgs, ignoredPkgs, err := excludePackages(allPkgs, opts.Excludes)
	if err != nil {
		return err
	}

	// function to inject that actually "prints" each line
	lineOut := func(str ...string) { fmt.Println(strings.Join(str, " ")) }

	// channel to receive each 'go test' stdout line
	goTestOutput := make(chan TestEvent)

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
	testArgs = append(testArgs, "-json")
	testArgs = append(testArgs, testPkgs...)
	err = shJSONPipe("go", testArgs, "", goTestOutput)
	wg.Wait()
	if err != nil {
		return err
	}

	// Open html coverage report
	if opts.FlagCoverReport {
		shCmd("go", shArgs{"tool", "cover", "-html=" + coverProfile}, "")
	}
	return nil
}
