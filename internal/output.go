package internal

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type processOutputParams struct {
	OutputChannel   chan string
	LineOut         func(str ...string)
	ToTestPackages  []string
	IgnoredPackages []string
	FlagVerbose     bool
	FlagSkipEmpty   bool
	FlagListEmpty   bool
	FlagListIgnored bool
	IndentSpaces    int
}

func processOutput(params *processOutputParams) {
	regexNoTests := regexp.MustCompile(`^\?\s+(.+)\s+\[no test files\]$`)
	regexPackageSummary := regexp.MustCompile(`^(ok  \t|FAIL\t)`)
	regexFailCoverage := regexp.MustCompile(`^coverage: `)
	regexCoverageString := regexp.MustCompile(`coverage: (\d{1,3}\.\d{1,2}%) of statements$`)
	regexCoverageNoStatements := regexp.MustCompile(`^coverage: \[no statements\]$`)
	regexRunLine := regexp.MustCompile(`^=== RUN`)
	regexPassFailLine := regexp.MustCompile(`^(PASS|FAIL)$`)
	regexTestSummary := regexp.MustCompile(`^\s*--- (PASS|FAIL): `)

	pkgsNoTests := []string{}
	pkgsFailed := []string{}
	coverages := []float64{}
	prevCoverage := ""

	maxPkgLen := 0
	for _, pkg := range params.ToTestPackages {
		pkglen := len(pkg)
		maxPkgLen = ifelse(maxPkgLen < pkglen, pkglen, maxPkgLen)
	}

	for line := range params.OutputChannel {
		// When a package fails coverage is printed alone, before that package's FAIL line.
		// So store it and append to the proper package summary line
		if regexFailCoverage.MatchString(line) {
			prevCoverage = line
			continue
		}

		// Omit "irrelevant" lines
		if regexPassFailLine.MatchString(line) || regexRunLine.MatchString(line) {
			continue
		}

		// Reduce indentation
		line = strings.ReplaceAll(line, "    ", strings.Repeat(" ", params.IndentSpaces))

		if regexNoTests.MatchString(line) {
			// Handle packages with no tests
			pkg := regexNoTests.ReplaceAllString(line, "$1")
			pkgsNoTests = append(pkgsNoTests, pkg)

			if params.FlagSkipEmpty {
				continue
			}

			coverages = append(coverages, 0)
			line = shColor("yellow:bold", "!") + " " + pkg
			line += strings.Repeat(" ", maxPkgLen-len(pkg)) + "   " + shColor("gray", sf("%6s", "0.0%"))
			line += "     " + shColor("yellow", "no tests")

		} else if regexPackageSummary.MatchString(line) {
			// Parse "ok/FAIL some/pkg" lines (package summary lines)
			parts := splitSummaryLine(line + "\t" + prevCoverage)

			pkgsFailed = sliceAppendIf(parts.result == "FAIL", pkgsFailed, parts.pkg)

			line = parts.result + shColor("reset:bold", parts.pkg)
			line = strings.Replace(line, "ok", shColor("green", "✔ "), 1)
			line = strings.Replace(line, "FAIL", shColor("red", "◼ "), 1)

			line += strings.Repeat(" ", maxPkgLen-len(parts.pkg))

			if regexCoverageNoStatements.MatchString(parts.coverage) {
				line += "   "
				line += shColor("gray", sf("%6s", "-")+"     no statements")
			} else {
				pkgCoverage := regexCoverageString.ReplaceAllString(parts.coverage, "$1")
				c := coverageParse(pkgCoverage)
				coverages = append(coverages, c)

				line += "   "
				line += shColor(coverageColor(c), sf("%6s", pkgCoverage))

				line += "     "
				line += strings.Replace(parts.elapsed, "(cached)", shColor("gray", "cached"), 1)
			}

			if params.FlagVerbose {
				line += "\n" + shColor("gray", strings.Repeat("-", maxPkgLen+22))
			}

		} else if regexTestSummary.MatchString(line) {
			// Parse "PASS/FAIL" lines (test summary lines)
			line = strings.Replace(line, "--- PASS: ", "✔ ", 1)
			line = strings.Replace(line, "--- FAIL: ", shColor("red", "✖ "), 1)

		} else {
			// All other lines
			line = strings.ReplaceAll(line, "\t", strings.Repeat(" ", params.IndentSpaces))
			line = shColor("whitesmoke", line)
		}

		params.LineOut(strings.TrimRightFunc(line, unicode.IsSpace))
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

func splitSummaryLine(line string) summaryLine {
	parts := strings.Split(line, "\t")
	return summaryLine{
		result:   strings.TrimSpace(sliceAt(parts, 0, "")),
		pkg:      strings.TrimSpace(sliceAt(parts, 1, "")),
		elapsed:  strings.TrimSpace(sliceAt(parts, 2, "")),
		coverage: strings.TrimSpace(sliceAt(parts, 3, "")),
	}
}

func coverageParse(cov string) float64 {
	c, err := strconv.ParseFloat(strings.Trim(cov, "% \t\n"), 64)
	return ifelse(err != nil, 0, c)
}

func coverageColor(cov float64) string {
	return ifelse(cov < 50, "red", ifelse(cov < 75, "yellow", "green"))
}
