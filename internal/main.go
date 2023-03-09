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

	// Get packages to test
	testPkgsMap, testPkgs, ignoredPkgs, err := getPackages(opts.TestPath, opts.Excludes)
	if err != nil {
		return err
	}

	// Create blank test files in no-tests packages (needed for fullCoverage)
	var newFiles []string
	var newPackages []Package
	if opts.FlagFullCoverage {
		lineOut(sf("\nGenerating empty tests for full coverage in '%s'", opts.TestPath))

		var err error
		newFiles, newPackages, err = fixPkgsWithNoTests(testPkgsMap, testPkgs)
		if err != nil {
			return err
		}

		defer deleteFiles(&newFiles)
	}

	// Determine cover-profile file name
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
			NoTestsPackages: newPackages,
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

// Helpers --------------

func getPackages(testPath string, excludes []string) (map[string]Package, []string, []string, error) {
	allPkgs := []string{}
	allPkgsMap := map[string]Package{}

	// Load all packages
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
		return nil, nil, nil, err
	}

	// Exclude packages to ignore
	pkgsToTest, pkgsIgnored, err := excludePackages(allPkgs, excludes)
	if err != nil {
		return nil, nil, nil, err
	}

	// Build packages to test map
	pkgsToTestMap := map[string]Package{}
	for _, pkg := range pkgsToTest {
		pkgsToTestMap[pkg] = allPkgsMap[pkg]
	}

	return pkgsToTestMap, pkgsToTest, pkgsIgnored, nil
}

// "Eliminate" no-tests pakages by creating blank test file in them
func fixPkgsWithNoTests(pkgsMap map[string]Package, pkgs []string) (newFiles []string, packages []Package, err error) {
	noTestsPkgs := []string{}
	goListOutput := make(chan TestEvent)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for o := range goListOutput {
			if o.Action == "skip" && o.Test == "" {
				noTestsPkgs = append(noTestsPkgs, o.Package)
			}
		}

		wg.Done()
	}()

	testArgs := shArgs{"test"}
	testArgs = append(testArgs, "-list", ".")
	testArgs = append(testArgs, "-json")
	testArgs = append(testArgs, pkgs...)
	err = shJSONPipe("go", testArgs, "", goListOutput, io.Discard)
	wg.Wait()
	if err != nil {
		return nil, nil, err
	}

	// Create blank test file in each no-test package
	for _, pkg := range noTestsPkgs {
		p := pkgsMap[pkg]

		file, err := os.CreateTemp(p.Dir, "dummy_*_test.go")
		if err != nil {
			return nil, nil, err
		}

		textToWrite := fmt.Sprintf("package %s\n", p.Name) // all that's need to be a valid test
		err = os.WriteFile(file.Name(), []byte(textToWrite), 0o666)
		if err != nil {
			return nil, nil, err
		}

		newFiles = append(newFiles, file.Name())
		packages = append(packages, p)
	}

	return newFiles, packages, nil
}
