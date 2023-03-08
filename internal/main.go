package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type RunInitEmptyOpts struct {
	TestPath string
	Excludes []string
}

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
	FlagFullCoverage bool
	Excludes         []string
	FlagTestOutput   string

	Azure AzureConf
}

type TestEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

type Package struct {
	Dir        string
	ImportPath string
	Name       string
}

var ErrTestRunIgnore = errors.New("test run error")

func RunTests(opts RunTestsOpts) error {
	color.NoColor = !opts.FlagColor

	// function to inject that actually "prints" each line
	lineOut := func(str ...string) { fmt.Println(strings.Join(str, " ")) }

	var newFiles []string
	var newPackages []Package
	if opts.FlagFullCoverage {
		lineOut(sf("\nGenerating empty tests for full coverage in '%s'", opts.TestPath))

		var err error
		newFiles, newPackages, err = initEmpty(opts.TestPath, opts.Excludes)
		if err != nil {
			return err
		}
		defer func() {
			for _, f := range newFiles {
				os.Remove(f)
			}
		}()
	}

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

	var coverProfile string
	if opts.FlagCoverReport || opts.FlagFullCoverage {
		coverProfile = opts.FlagCoverProfile
		// If empty use throw-away coverage profile
		if coverProfile == "" {
			tempCoverProfile, err := os.CreateTemp("", "coverage-*.out")
			if err != nil {
				return err
			}
			defer os.Remove(tempCoverProfile.Name())
			coverProfile = tempCoverProfile.Name()
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	goTestOutput := make(chan TestEvent) // channel to receive each 'go test' stdout line
	var failedTests []string
	var totalCoverage float64

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
			DummyPackages:   newPackages,
			CoverProfile:    coverProfile,
			FailedTests:     &failedTests,
			TotalCoverage:   &totalCoverage,
		})
		wg.Done()
	}()

	testOut := io.Discard
	if opts.FlagTestOutput != "" {
		file, err := os.OpenFile(opts.FlagTestOutput, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		if err != nil {
			return err
		}
		testOut = file
	}

	// Compose and run 'go test ...'
	lineOut(sf("\nTesting %d packages in '%s'\n", len(testPkgs), opts.TestPath))
	testArgs := shArgs{"test"}
	testArgs = sliceAppendIf(opts.FlagVerbose, testArgs, "-v")
	testArgs = sliceAppendIf(!opts.FlagCache, testArgs, "-count=1")
	testArgs = sliceAppendIf(opts.FlagCover, testArgs, "-cover")
	testArgs = sliceAppendIf(coverProfile != "", testArgs, "-coverprofile="+coverProfile)
	testArgs = append(testArgs, "-json")
	testArgs = append(testArgs, testPkgs...)
	testErr := shJSONPipe("go", testArgs, "", goTestOutput, testOut)
	wg.Wait()

	// Publish Azure Coverage PR comment
	opts.Azure.sendAzureComment(totalCoverage, failedTests)

	if testErr != nil {
		return ErrTestRunIgnore
	}

	// Open html coverage report
	if opts.FlagCoverReport {
		shCmd("go", shArgs{"tool", "cover", "-html=" + coverProfile}, "")
	}

	return nil
}

// Add dummy test files to packages with no tests
func initEmpty(testPath string, excludes []string) (newFiles []string, packages []Package, err error) {
	allPkgs := []string{}
	allPkgsMap := map[string]Package{}

	{
		pkgChan := make(chan Package)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			for p := range pkgChan {
				allPkgs = append(allPkgs, p.ImportPath)
				allPkgsMap[p.ImportPath] = p
			}
			wg.Done()
		}()
		err := shJSONPipe("go", shArgs{"list", "-json", testPath}, "", pkgChan, io.Discard)
		wg.Wait()
		if err != nil {
			return nil, nil, err
		}
	}

	skippedPkgs := []string{}

	{
		testPkgs, _, err := excludePackages(allPkgs, excludes)

		if err != nil {
			return nil, nil, err
		}

		goListOutput := make(chan TestEvent)

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			for o := range goListOutput {
				if o.Action == "skip" && o.Test == "" {
					skippedPkgs = append(skippedPkgs, o.Package)
				}
			}

			wg.Done()
		}()

		testArgs := shArgs{"test"}
		testArgs = append(testArgs, "-list", ".")
		testArgs = append(testArgs, "-json")
		testArgs = append(testArgs, testPkgs...)
		err = shJSONPipe("go", testArgs, "", goListOutput, io.Discard)
		wg.Wait()
		if err != nil {
			return nil, nil, err
		}
	}

	defer func() {
		if err != nil {
			for _, f := range newFiles {
				os.Remove(f)
			}
		}
	}()

	for _, importPath := range skippedPkgs {
		p, ok := allPkgsMap[importPath]
		if !ok {
			return nil, nil, fmt.Errorf("package %q not found in map", importPath)
		}

		// this is all that's need to be printed to be a valid test
		textToWrite := "package " + p.Name + "\n"

		file, err := os.CreateTemp(p.Dir, "dummy_*_test.go")
		if err != nil {
			return nil, nil, err
		}

		// we know we can write to dummy_test, because there is no test in the package
		err = os.WriteFile(file.Name(), []byte(textToWrite), 0o666)
		if err != nil {
			return nil, nil, err
		}

		newFiles = append(newFiles, file.Name())
		packages = append(packages, p)
	}
	return newFiles, packages, nil
}
