package internal

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
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

	Err *error
}

func processOutput(params *processOutputParams) {
	regexNoTests := regexp.MustCompile(`^\?\s+(.+)\s+\[no test files\]$`)
	regexPackageSummary := regexp.MustCompile(`^(ok  \t|FAIL\t)`)
	regexCoverageAny := regexp.MustCompile(`^coverage: `)
	regexCoverageNonZero := regexp.MustCompile(`^coverage: (\d{1,3}\.\d{1,2}%) of statements\n$`)
	regexCoverageNoStatements := regexp.MustCompile(`^coverage: \[no statements\]\n$`)

	regexRunLine := regexp.MustCompile(`^=== (RUN|CONT)`)

	regexPassFailLine := regexp.MustCompile(`^(PASS|FAIL)$`)
	regexTestSummary := regexp.MustCompile(`^\s*--- (PASS|FAIL|SKIP): `)
	regexTestSkip := regexp.MustCompile(`^\s*--- SKIP: `)

	pkgsNoTests := []string{}
	pkgsFailed := []string{}
	coverages := []float64{}

	// the JSON output is printing all verbose; so we need to implement the FAIL output filtering ourselves
	linesPerTest := map[string][]string{}

	prevCoverages := map[string]string{}

	lineOutTrimmed := func(s string) {
		lines := strings.Split(s, "\n")
		for _, l := range lines {
			trimmed := strings.TrimRightFunc(l, unicode.IsSpace)
			if trimmed != "" {
				params.LineOut(trimmed)
			}
		}
	}

	maxPkgLen := 0
	for _, pkg := range params.ToTestPackages {
		pkglen := len(pkg)
		maxPkgLen = ifelse(maxPkgLen < pkglen, pkglen, maxPkgLen)
	}

	for event := range params.OutputChannel {

		if event.Action == "output" {
			if regexCoverageAny.MatchString(event.Output) {
				prevCoverages[event.Package] = event.Output
				continue
			}

			if regexPackageSummary.MatchString(event.Output) ||
				regexPassFailLine.MatchString(event.Output) ||
				regexRunLine.MatchString(event.Output) ||
				regexNoTests.MatchString(event.Output) {
				continue
			}

			testOutLine := event.Output
			testOutLine = strings.TrimRightFunc(testOutLine, unicode.IsSpace)

			testOutLine = strings.ReplaceAll(testOutLine, "    ", strings.Repeat(" ", params.IndentSpaces))
			if !regexTestSummary.MatchString(testOutLine) {
				// we need to save normal lines in case we want to print them later
				if testOutLine != "" {
					// All other lines
					testOutLine = strings.ReplaceAll(testOutLine, "\t", strings.Repeat(" ", params.IndentSpaces))
					testOutLine = shColor("whitesmoke", testOutLine)
					if event.Test != "" {
						linesPerTest[event.Test] = append(linesPerTest[event.Test], testOutLine)
					}
				}
			} else {
				isFail := strings.Contains(testOutLine, "--- FAIL")
				isSkipped := strings.Contains(testOutLine, "--- SKIP")

				// Parse "PASS/FAIL/SKIP" lines (test summary lines)
				testOutLine = strings.Replace(testOutLine, "(0.00s)", "", 1)

				testOutLine = strings.Replace(testOutLine, "--- PASS: ", shColor("whitesmoke", "✔ "), 1)
				testOutLine = strings.Replace(testOutLine, "--- FAIL: ", shColor("red", "✖ "), 1)

				if regexTestSkip.MatchString(testOutLine) {
					testOutLine = strings.Replace(testOutLine, "--- SKIP: ", shColor("gray", "≋ "), 1) + "    " + shColor("gray", "skipped")
				}

				// print first the FAIL/SKIP/PASS lines, then the actual output
				if isFail || isSkipped || params.FlagVerbose {
					lineOutTrimmed(testOutLine)
					for _, l := range linesPerTest[event.Test] {
						lineOutTrimmed(l)
					}
				}
			}
		}

		var outLine string

		if event.Test == "" && event.Action == "skip" {
			pkg := event.Package
			pkgsNoTests = append(pkgsNoTests, pkg)
			coverages = append(coverages, 0)

			if params.FlagSkipEmpty {
				continue
			}

			outLine = shColor("yellow:bold", "!") + " " + pkg
			outLine += strings.Repeat(" ", maxPkgLen-len(pkg)) + "   " + shColor("gray", sf("%6s", "0.0%"))
			outLine += "     " + shColor("yellow", "no tests")
		}

		if event.Test == "" && (event.Action == "pass" || event.Action == "fail") {
			if event.Action == "pass" {
				outLine = shColor("green", "✔ ") + shColor("reset:bold", event.Package)
			} else {
				outLine = shColor("red", "◼ ") + shColor("reset:bold", event.Package)
				pkgsFailed = append(pkgsFailed, event.Package)
			}
			outLine += strings.Repeat(" ", maxPkgLen-len(event.Package))

			pastCoverage := prevCoverages[event.Package]
			delete(prevCoverages, pastCoverage)
			if regexCoverageNoStatements.MatchString(pastCoverage) {
				outLine += "   "
				outLine += shColor("gray", sf("%6s", "-")+"     no statements")
			} else {
				pkgCoverage := regexCoverageNonZero.ReplaceAllString(pastCoverage, "$1")
				c := coverageParse(pkgCoverage)
				coverages = append(coverages, c)

				outLine += "   "
				outLine += shColor(coverageColor(c), sf("%6s", pkgCoverage))

				outLine += "     "

				if event.Elapsed == 0 {
					// slight heuristics - elipsed 0 == cached (we cannot tell otherwise from the JSON)
					outLine += shColor("gray", "cached")
				} else {
					outLine += fmt.Sprintf("%.3fs", event.Elapsed)
				}
			}

			if params.FlagVerbose {
				outLine += "\n" + shColor("gray", strings.Repeat("-", maxPkgLen+22))
			}
		}
		lineOutTrimmed(outLine)
	}

	params.LineOut()

	chev := shColor("gray", "❯")
	avgCoverage := sliceAvg(coverages)
	cover := sf("%.2f", avgCoverage) + "%"
	params.LineOut(sf("%s Coverage: %s", chev, shColor(coverageColor(avgCoverage)+":bold", cover)))
	pkgs := sf("tested: %d", len(params.ToTestPackages))
	pkgs += shColor("red", sf("    failed: %d", len(pkgsFailed)))
	pkgs += shColor("yellow", sf("    noTests: %d", len(pkgsNoTests)))
	pkgs += shColor("gray", sf("    excluded: %d", len(params.IgnoredPackages)))
	params.LineOut(sf("%s Pkgs: %s", chev, pkgs))

	params.LineOut()

	if params.FlagListEmpty {
		params.LineOut(shColor("yellow:bold", "Packages with no tests:"))
		for _, pkg := range pkgsNoTests {
			params.LineOut("- " + pkg)
		}
		if len(pkgsNoTests) != 0 {
			*params.Err = errors.New("there are packages without tests")
		}
	}

	if params.FlagListIgnored {
		params.LineOut(shColor("yellow:bold", "Packages ignored:"))
		for _, pkg := range params.IgnoredPackages {
			params.LineOut("- " + pkg)
		}
	}
}

type summaryLine struct {
	result   string
	pkg      string
	elapsed  string
	coverage string
}

func coverageParse(cov string) float64 {
	c, err := strconv.ParseFloat(strings.Trim(cov, "% \t\n"), 64)
	return ifelse(err != nil, 0, c)
}

func coverageColor(cov float64) string {
	return ifelse(cov < 50, "red", ifelse(cov < 75, "yellow", "green"))
}
