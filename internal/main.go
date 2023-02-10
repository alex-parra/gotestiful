package internal

import (
	"fmt"
	"os"
	"strconv"
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

	var coverProfile string
	if opts.FlagCoverReport {
		coverProfile = zvfb(opts.FlagCoverProfile, "./coverage.out")
	} else {
		newF, err := os.CreateTemp("", "coverage-*.out")
		if err != nil {
			return err
		}
		defer os.Remove(newF.Name())
		coverProfile = newF.Name()
	}

	// Compose and run 'go test ...'
	lineOut(sf("\nTesting %d packages in '%s'\n", len(testPkgs), opts.TestPath))
	testArgs := shArgs{"test"}
	testArgs = sliceAppendIf(opts.FlagVerbose, testArgs, "-v")
	testArgs = sliceAppendIf(!opts.FlagCache, testArgs, "-count=1")
	testArgs = sliceAppendIf(opts.FlagCover, testArgs, "-cover")
	testArgs = append(testArgs, "-coverprofile="+coverProfile)
	testArgs = append(testArgs, "-json")
	testArgs = append(testArgs, testPkgs...)
	err = shJSONPipe("go", testArgs, "", goTestOutput)
	wg.Wait()
	if err != nil {
		// print coverage even in the case of error (as test failure is error here); ignore its own error
		printCoverage(coverProfile, lineOut)
		return err
	}

	err = printCoverage(coverProfile, lineOut)
	if err != nil {
		return err
	}

	// Open html coverage report
	if opts.FlagCoverReport {
		shCmd("go", shArgs{"tool", "cover", "-html=" + coverProfile}, "")
	}
	return nil
}

func printCoverage(path string, lineOut func(...string)) error {
	var coverage float64

	// now read proper code coverage
	goCoverOutput := make(chan string)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for line := range goCoverOutput {
			if strings.HasPrefix(line, "total:") {
				// a bit of a hack but it works
				line = strings.Trim(line, "total: (statements) %\t\n")
				coverage, _ = strconv.ParseFloat(line, 64)
			}
		}
		wg.Done()
	}()

	err := shPipe("go", shArgs{"tool", "cover", "-func", path}, "", goCoverOutput)
	if err != nil {
		return err
	}

	chev := shColor("gray", "❯")
	cover := sf("%.2f", coverage) + "%"

	lineOut(sf("%s Coverage: %s", chev, shColor(coverageColor(coverage)+":bold", cover)))
	return nil
}
