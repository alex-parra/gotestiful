package internal

import (
	"strings"
	"sync"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func runTests(p *processOutputParams, outLines ...string) []string {
	c := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)

	out := []string{}
	lineOut := func(str ...string) {
		out = append(out, strings.Join(str, " "))
	}

	go func() {
		processOutput(&processOutputParams{
			OutputChannel:   c,
			LineOut:         lineOut,
			ToTestPackages:  p.ToTestPackages,
			IgnoredPackages: p.IgnoredPackages,
			FlagVerbose:     p.FlagVerbose,
			FlagSkipEmpty:   p.FlagSkipEmpty,
			FlagListEmpty:   p.FlagListEmpty,
			FlagListIgnored: p.FlagListIgnored,
			IndentSpaces:    2,
		})
		wg.Done()
	}()

	for _, l := range outLines {
		c <- l
	}

	close(c)
	wg.Wait()
	return out
}

func TestProcessOutput(t *testing.T) {
	color.NoColor = true

	t.Run("PASS only line", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"PASS",
		)

		assert.Equal(t, out, []string{
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("PASS line", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"    --- PASS: SomeTest",
		)

		assert.Equal(t, out, []string{
			"  ✔ SomeTest",
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("FAIL only line", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"FAIL",
		)

		assert.Equal(t, out, []string{
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("FAIL line", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"    --- FAIL: SomeTest",
		)

		assert.Equal(t, out, []string{
			"  ✖ SomeTest",
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("=== RUN only line", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"=== RUN SomeTest ...",
		)

		assert.Equal(t, out, []string{
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("no tests line (no skip)", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"? some/awesome/pkg [no test files]",
		)

		assert.Equal(t, out, []string{
			"! some/awesome/pkg     0.0%     no tests",
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 1    excluded: 0",
			"",
		})
	})

	t.Run("no tests line (skip)", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}, FlagSkipEmpty: true},
			"? some/awesome/pkg [no test files]",
		)

		assert.Equal(t, out, []string{
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 1    excluded: 0",
			"",
		})
	})

	t.Run("no tests (list)", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}, FlagListEmpty: true},
			"? some/awesome/pkg [no test files]",
		)

		assert.Equal(t, out, []string{
			"! some/awesome/pkg     0.0%     no tests",
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 1    excluded: 0",
			"",
			"Packages with no tests:",
			"- some/awesome/pkg",
		})
	})

	t.Run("ok line", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"ok  \tsome/awesome/pkg\t0.123s\tcoverage: 54.3% of statements",
		)

		assert.Equal(t, out, []string{
			"✔ some/awesome/pkg    54.3%     0.123s",
			"",
			"❯ Coverage: 54.30%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("coverage no statements", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"ok  \tsome/awesome/pkg\t0.123s\tcoverage: [no statements]",
		)

		assert.Equal(t, out, []string{
			"✔ some/awesome/pkg        -     no statements",
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("fail line (with coverage from previous line)", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"coverage: 12.3% of statements",
			"FAIL\tsome/awesome/pkg\t0.123s",
		)

		assert.Equal(t, out, []string{
			"◼ some/awesome/pkg    12.3%     0.123s",
			"",
			"❯ Coverage: 12.30%",
			"❯ Pkgs: tested: 1    failed: 1    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("unmatched lines", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}},
			"Some test info\t123 is not 234",
			"    more test debug infos...",
		)

		assert.Equal(t, out, []string{
			"Some test info  123 is not 234",
			"  more test debug infos...",
			"",
			"❯ Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("verbose", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}, FlagVerbose: true},
			"ok  \tsome/awesome/pkg\t(cached)\tcoverage: 54.3% of statements",
		)

		assert.Equal(t, out, []string{
			"✔ some/awesome/pkg    54.3%     cached\n" + strings.Repeat("-", 38),
			"",
			"❯ Coverage: 54.30%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		})
	})

	t.Run("list ignored", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"some/awesome/pkg"}, IgnoredPackages: []string{"a/pkg/to/ignore"}, FlagListIgnored: true},
			"ok  \tsome/awesome/pkg\t(cached)\tcoverage: 54.3% of statements",
		)

		assert.Equal(t, out, []string{
			"✔ some/awesome/pkg    54.3%     cached",
			"",
			"❯ Coverage: 54.30%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 1",
			"",
			"Packages ignored:",
			"- a/pkg/to/ignore",
		})
	})

}

func TestSplitSummaryLine(t *testing.T) {
	t.Run("with missing elapsed and coverage", func(t *testing.T) {
		expected := summaryLine{result: "res", pkg: "some/awe/some/pkg", elapsed: "", coverage: ""}
		actual := splitSummaryLine("   res\t some/awe/some/pkg   \t   ")
		assert.Equal(t, expected, actual)
	})

	t.Run("with missing coverage", func(t *testing.T) {
		expected := summaryLine{result: "res", pkg: "some/awe/some/pkg", elapsed: "0.3214s", coverage: ""}
		actual := splitSummaryLine("   res\t some/awe/some/pkg   \t  0.3214s ")
		assert.Equal(t, expected, actual)
	})

	t.Run("all present", func(t *testing.T) {
		expected := summaryLine{result: "res", pkg: "some/awe/some/pkg", elapsed: "0.1234s", coverage: "coverage: 50.0% of statements"}
		actual := splitSummaryLine("   res\t some/awe/some/pkg   \t  0.1234s    \t coverage: 50.0% of statements")
		assert.Equal(t, expected, actual)
	})
}

func TestCoverageParse(t *testing.T) {
	assert.Equal(t, 12.3, coverageParse("  12.30% "))
	assert.Equal(t, 3.2, coverageParse("3.20%\n"))
	assert.Equal(t, 100.0, coverageParse("\t100.00%"))
}

func TestCoverageColor(t *testing.T) {
	assert.Equal(t, "green", coverageColor(95.0))
	assert.Equal(t, "green", coverageColor(75.0))

	assert.Equal(t, "yellow", coverageColor(74.9))
	assert.Equal(t, "yellow", coverageColor(50.0))

	assert.Equal(t, "red", coverageColor(49.9))
	assert.Equal(t, "red", coverageColor(20))

	assert.Equal(t, "red", coverageColor(0))
}
