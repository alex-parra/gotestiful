package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/exp/slices"
)

type processOutputParams struct {
	OutputChannel   <-chan TestEvent
	LineOut         func(str ...string)
	ToTestPackages  []string
	IgnoredPackages []string
	FlagVerbose     bool
	FlagSkipEmpty   bool
	FlagListEmpty   bool
	FlagListIgnored bool
	IndentSpaces    int
	DummyPackages   []Package
	AverageCoverage bool

	FailedTests *[]string
}

var regexNoTests = regexp.MustCompile(`^\?\s+(.+)\s+\[no test files\]$`)
var regexPackageSummary = regexp.MustCompile(`^(ok  \t|FAIL\t)`)
var regexCoverageAny = regexp.MustCompile(`^coverage: `)
var regexCoverageNonZero = regexp.MustCompile(`^coverage: (\d{1,3}\.\d{1,2}%) of statements\n$`)
var regexCoverageNoStatements = regexp.MustCompile(`^coverage: \[no statements\]\n$`)
var regexRunLine = regexp.MustCompile(`^=== (RUN|CONT|PAUSE)`)
var regexPassFailLine = regexp.MustCompile(`^(PASS|FAIL)$`)
var regexTestSummary = regexp.MustCompile(`^\s*--- (PASS|FAIL|SKIP): `)
var regexTestSkip = regexp.MustCompile(`^\s*--- SKIP: `)

func processOutput(params *processOutputParams) {
	pkgsNoTests := []string{}
	pkgsFailed := []string{}
	failedTests := []string{}
	coverages := []float64{}
	testOutputLines := map[string][]string{}
	prevCoverages := map[string]string{}

	maxPkgLen := 0
	for _, pkg := range params.ToTestPackages {
		pkglen := len(pkg)
		maxPkgLen = ifelse(maxPkgLen < pkglen, pkglen, maxPkgLen)
	}

	dummyPackages := map[string]bool{}
	for _, p := range params.DummyPackages {
		dummyPackages[p.ImportPath] = true
	}

	lineOutTrimmed := func(s string) {
		lines := strings.Split(s, "\n")
		for _, l := range lines {
			trimmed := strings.TrimRightFunc(l, unicode.IsSpace)
			if trimmed != "" {
				params.LineOut(trimmed)
			}
		}
	}

	printSkipped := func(pkg string) {
		pkgsNoTests = append(pkgsNoTests, pkg)

		if !params.FlagSkipEmpty {
			coverages = append(coverages, 0)

			outLine := shColor("yellow:bold", "!") + " " + pkg
			outLine += strings.Repeat(" ", maxPkgLen-len(pkg)) + "   " + shColor("gray", sf("%6s", "0.0%"))
			outLine += "     " + shColor("yellow", "no tests")
			lineOutTrimmed(outLine)
		}
	}

	// Parse each line output
	for event := range params.OutputChannel {

		if event.Action == "output" {
			if regexPackageSummary.MatchString(event.Output) ||
				regexPassFailLine.MatchString(event.Output) ||
				regexRunLine.MatchString(event.Output) ||
				regexNoTests.MatchString(event.Output) {
				continue
			}

			// Save coverage for later
			if regexCoverageAny.MatchString(event.Output) {
				prevCoverages[event.Package] = event.Output
				continue
			}

			testOutLine := event.Output
			testOutLine = strings.TrimRightFunc(testOutLine, unicode.IsSpace)
			testOutLine = strings.ReplaceAll(testOutLine, "    ", strings.Repeat(" ", params.IndentSpaces))

			if regexTestSummary.MatchString(testOutLine) {
				isFail := strings.Contains(testOutLine, "--- FAIL")
				if isFail && event.Test != "" && !slices.Contains(failedTests, event.Test) {
					failedTests = append(failedTests, event.Test)
				}

				testOutLine = strings.Replace(testOutLine, "(0.00s)", "", 1)
				testOutLine = strings.Replace(testOutLine, "--- PASS: ", shColor("whitesmoke", "✔ "), 1)
				testOutLine = strings.Replace(testOutLine, "--- FAIL: ", shColor("red", "✖ "), 1)

				if regexTestSkip.MatchString(testOutLine) {
					testOutLine = strings.Replace(testOutLine, "--- SKIP: ", shColor("gray", "≋ "), 1) + "    " + shColor("gray", "skipped")
				}

				// Print non-package lines if verbose or the test failed
				if params.FlagVerbose || slices.Contains(failedTests, event.Test) {
					lineOutTrimmed(testOutLine)
					for _, l := range testOutputLines[event.Test] {
						lineOutTrimmed(l)
					}
					// clear already printed lines
					testOutputLines[event.Test] = []string{}
				}

			} else if testOutLine != "" {
				testOutLine = strings.ReplaceAll(testOutLine, "\t", strings.Repeat(" ", params.IndentSpaces))
				testOutLine = shColor("whitesmoke", testOutLine)

				if event.Test != "" {
					// we need to save some lines so we can print them later
					testOutputLines[event.Test] = append(testOutputLines[event.Test], testOutLine)
				}
			}
		}

		// Print no test packages
		if event.Test == "" && (event.Action == "skip" || (event.Action == "pass" && dummyPackages[event.Package])) {
			printSkipped(event.Package)
			continue
		}

		// Print Package PASS / FAIL lines
		var outLine string
		if event.Test == "" && (event.Action == "pass" || event.Action == "fail") {

			if event.Action == "pass" {
				outLine = shColor("green", "✔ ") + shColor("reset:bold", event.Package)
			}

			if event.Action == "fail" {
				pkgsFailed = append(pkgsFailed, event.Package)
				outLine = shColor("red", "◼ ") + shColor("reset:bold", event.Package)
			}

			outLine += strings.Repeat(" ", maxPkgLen-len(event.Package))

			// Build package coverage + elapsed
			prevCoverage := prevCoverages[event.Package]
			if regexCoverageNoStatements.MatchString(prevCoverage) {
				outLine += "   " + shColor("gray", sf("%6s", "-")+"     no statements")

			} else {
				pkgCoverage := regexCoverageNonZero.ReplaceAllString(prevCoverage, "$1")
				c := coverageParse(pkgCoverage)
				coverages = append(coverages, c)
				outLine += "   " + shColor(coverageColor(c), sf("%6s", pkgCoverage)) + "     "

				if event.Elapsed == 0 {
					outLine += shColor("gray", "cached") // elapsed 0 == cached (we cannot tell otherwise from the JSON)
				} else {
					outLine += fmt.Sprintf("%.3fs", event.Elapsed)
				}
			}

			if params.FlagVerbose {
				// print a separator between packages
				outLine += "\n" + shColor("gray", strings.Repeat("-", maxPkgLen+22))
			}
		}

		lineOutTrimmed(outLine)
	}

	params.LineOut()

	// Print summary
	chev := shColor("gray", "❯")
	pkgs := sf("tested: %d", len(params.ToTestPackages))
	pkgs += shColor("red", sf("    failed: %d", len(pkgsFailed)))
	pkgs += shColor("yellow", sf("    noTests: %d", len(pkgsNoTests)))
	pkgs += shColor("gray", sf("    excluded: %d", len(params.IgnoredPackages)))
	params.LineOut(sf("%s Pkgs: %s", chev, pkgs))

	// Print coverage
	avgCoverage := sliceAvg(coverages)
	covFormatted := sf("%.2f", avgCoverage) + "%"
	covColor := coverageColor(avgCoverage) + ":bold"
	if params.AverageCoverage {
		params.LineOut(sf("%s Coverage: %s   [average]   %s", chev, shColor(covColor, covFormatted), shColor("gray", "(set flag 'fullCoverage' for accurate calculation)")))
	}

	if params.FlagListEmpty {
		params.LineOut()
		params.LineOut(shColor("yellow:bold", "Packages with no tests:"))
		for _, pkg := range pkgsNoTests {
			params.LineOut("- " + pkg)
		}
	}

	if params.FlagListIgnored {
		params.LineOut()
		params.LineOut(shColor("yellow:bold", "Packages ignored:"))
		for _, pkg := range params.IgnoredPackages {
			params.LineOut("- " + pkg)
		}
	}

	if params.FailedTests != nil {
		*params.FailedTests = failedTests
	}
}

func coverageParse(cov string) float64 {
	c, err := strconv.ParseFloat(strings.Trim(cov, "% \t\n"), 64)
	return ifelse(err != nil, 0, c)
}

func coverageColor(cov float64) string {
	return ifelse(cov < 50, "red", ifelse(cov < 75, "yellow", "green"))
}
