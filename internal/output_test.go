package internal

import (
	"strings"
	"sync"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func runTests(p *processOutputParams, outLines ...TestEvent) []string {
	c := make(chan TestEvent)
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
			AverageCoverage: true,
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

	t.Run("one dummy test, no coverage flag", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}},

			TestEvent{Action: "run", Package: "tst", Test: "TestDummy"},
			TestEvent{Action: "output", Package: "tst", Test: "TestDummy", Output: "=== RUN   TestDummy\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestDummy", Output: "--- PASS: TestDummy (0.00s)\n"},
			TestEvent{Action: "pass", Package: "tst", Test: "TestDummy", Elapsed: 0},
			TestEvent{Action: "output", Package: "tst", Output: "PASS\n"},
			TestEvent{Action: "output", Package: "tst", Output: "ok  \ttst\t0.266s\n"},
			TestEvent{Action: "pass", Package: "tst", Elapsed: 0.266},
		)

		assert.Equal(t, []string{
			"✔ tst              0.266s",
			"",
			"❯ Average Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		}, out)
	})

	t.Run("one dummy test, coverage flag", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}},

			TestEvent{Action: "run", Package: "tst", Test: "TestDummy"},
			TestEvent{Action: "output", Package: "tst", Test: "TestDummy", Output: "=== RUN   TestDummy\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestDummy", Output: "--- PASS: TestDummy (0.00s)\n"},
			TestEvent{Action: "pass", Package: "tst", Test: "TestDummy", Elapsed: 0},
			TestEvent{Action: "output", Package: "tst", Output: "PASS\n"},
			TestEvent{Action: "output", Package: "tst", Output: "coverage: 0.0% of statements\n"},
			TestEvent{Action: "output", Package: "tst", Output: "ok  \ttst\t0.266s\n"},
			TestEvent{Action: "pass", Package: "tst", Elapsed: 0.266},
		)

		assert.Equal(t, []string{
			"✔ tst     0.0%     0.266s",
			"",
			"❯ Average Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		}, out)
	})

	// with the new logic, skipped are written only on verbose
	t.Run("one skipped", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}},

			TestEvent{Action: "run", Package: "tst", Test: "TestOther"},

			TestEvent{Action: "output", Package: "tst", Test: "TestOther", Output: "=== RUN   TestOther\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestOther", Output: "    code_test.go:10: some reason to skip\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestOther", Output: "--- SKIP: TestOther (0.00s)\n"},
			TestEvent{Action: "pass", Package: "tst", Test: "TestOther", Elapsed: 0},

			TestEvent{Action: "output", Package: "tst", Output: "PASS\n"},
			TestEvent{Action: "output", Package: "tst", Output: "ok  \ttst\t0.266s\n"},
			TestEvent{Action: "pass", Package: "tst", Elapsed: 0.266},
		)

		assert.Equal(t, []string{
			"≋ TestOther     skipped",
			"  code_test.go:10: some reason to skip",
			"✔ tst              0.266s",
			"",
			"❯ Average Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		}, out)
	})

	t.Run("failing test", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}},

			TestEvent{Action: "run", Package: "tst", Test: "TestFailing"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "=== RUN   TestFailing\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "if you having correctness problems i feel bad for you son\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "i've got 99 problems\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "    code_test.go:12: but a test ain't one\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "--- FAIL: TestFailing (0.00s)\n"},
			TestEvent{Action: "fail", Package: "tst", Test: "TestFailing", Elapsed: 0},
			TestEvent{Action: "output", Package: "tst", Output: "FAIL\n"},
			TestEvent{Action: "output", Package: "tst", Output: "exit status 1\n"},
			TestEvent{Action: "output", Package: "tst", Output: "FAIL\ttst\t0.308s\n"},
			TestEvent{Action: "fail", Package: "tst", Elapsed: 0.308},
		)

		assert.Equal(t, []string{
			"✖ TestFailing",
			"if you having correctness problems i feel bad for you son",
			"i've got 99 problems",
			"  code_test.go:12: but a test ain't one",
			"◼ tst              0.308s",
			"",
			"❯ Average Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 1    noTests: 0    excluded: 0",
			"",
		}, out)
	})

	t.Run("no tests line (no skip)", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}},

			TestEvent{Action: "output", Package: "tst", Output: "?   \ttst\t[no test files]\n"},
			TestEvent{Action: "skip", Package: "tst", Elapsed: 0},
		)

		assert.Equal(t, []string{
			"! tst     0.0%     no tests",
			"",
			"❯ Average Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 1    excluded: 0",
			"",
		}, out)
	})

	t.Run("no tests line (skip)", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}, FlagSkipEmpty: true},

			TestEvent{Action: "output", Package: "tst", Output: "?   \ttst\t[no test files]\n"},
			TestEvent{Action: "skip", Package: "tst", Elapsed: 0},
		)

		assert.Equal(t, []string{
			"",
			"❯ Average Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 1    excluded: 0",
			"",
		}, out)
	})

	t.Run("no tests line (list)", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}, FlagListEmpty: true},

			TestEvent{Action: "output", Package: "tst", Output: "?   \ttst\t[no test files]\n"},
			TestEvent{Action: "skip", Package: "tst", Elapsed: 0},
		)

		assert.Equal(t, []string{
			"! tst     0.0%     no tests",
			"",
			"❯ Average Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 1    excluded: 0",
			"",
			"Packages with no tests:",
			"- tst",
		}, out)
	})

	t.Run("ok line, coverage", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}},

			TestEvent{Action: "run", Package: "tst", Test: "TestGood"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "=== RUN   TestGood\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "this line should be ignored, as we are not verbose\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "--- PASS: TestGood (0.00s)\n"},
			TestEvent{Action: "pass", Package: "tst", Test: "TestGood", Elapsed: 0},
			TestEvent{Action: "output", Package: "tst", Output: "PASS\n"},
			TestEvent{Action: "output", Package: "tst", Output: "coverage: 50.0% of statements\n"},
			TestEvent{Action: "output", Package: "tst", Output: "ok  \ttst\t0.185s\n"},
			TestEvent{Action: "pass", Package: "tst", Elapsed: 0.186},
		)

		assert.Equal(t, []string{
			"✔ tst    50.0%     0.186s",
			"",
			"❯ Average Coverage: 50.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		}, out)
	})

	t.Run("coverage no statements", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}},

			TestEvent{Action: "run", Package: "tst", Test: "TestGood"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "=== RUN   TestGood\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "this line should be ignored, as we are not verbose\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "--- PASS: TestGood (0.00s)\n"},
			TestEvent{Action: "pass", Package: "tst", Test: "TestGood", Elapsed: 0},
			TestEvent{Action: "output", Package: "tst", Output: "PASS\n"},
			TestEvent{Action: "output", Package: "tst", Output: "coverage: [no statements]\n"},
			TestEvent{Action: "output", Package: "tst", Output: "ok  \ttst\t0.110s\n"},
			TestEvent{Action: "pass", Package: "tst", Elapsed: 0.11},
		)

		assert.Equal(t, []string{
			"✔ tst        -     no statements",
			"",
			"❯ Average Coverage: 0.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 0",
			"",
		}, out)
	})

	t.Run("one test fail, one succseccful, no verbose, coverage", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}},

			TestEvent{Action: "run", Package: "tst", Test: "TestFailing"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "=== RUN   TestFailing\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "if you having correctness problems i feel bad for you son\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "i've got 99 problems\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "    code_test.go:12: but a test ain't one\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "--- FAIL: TestFailing (0.00s)\n"},
			TestEvent{Action: "fail", Package: "tst", Test: "TestFailing", Elapsed: 0},
			TestEvent{Action: "run", Package: "tst", Test: "TestGood"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "=== RUN   TestGood\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "this one should not be printed\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "because it's successful\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "--- PASS: TestGood (0.00s)\n"},
			TestEvent{Action: "pass", Package: "tst", Test: "TestGood", Elapsed: 0},
			TestEvent{Action: "output", Package: "tst", Output: "FAIL\n"},
			TestEvent{Action: "output", Package: "tst", Output: "coverage: 50.0% of statements\n"},
			TestEvent{Action: "output", Package: "tst", Output: "exit status 1\n"},
			TestEvent{Action: "output", Package: "tst", Output: "FAIL\ttst\t0.108s\n"},
			TestEvent{Action: "fail", Package: "tst", Elapsed: 0.108},
		)

		assert.Equal(t, []string{
			"✖ TestFailing",
			"if you having correctness problems i feel bad for you son",
			"i've got 99 problems",
			"  code_test.go:12: but a test ain't one",
			"◼ tst    50.0%     0.108s",
			"",
			"❯ Average Coverage: 50.00%",
			"❯ Pkgs: tested: 1    failed: 1    noTests: 0    excluded: 0",
			"",
		}, out)
	})

	t.Run("one test fail, one succseccful, verbose, coverage", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}, FlagVerbose: true},

			TestEvent{Action: "run", Package: "tst", Test: "TestFailing"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "=== RUN   TestFailing\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "if you having correctness problems i feel bad for you son\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "i've got 99 problems\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "    code_test.go:12: but a test ain't one\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestFailing", Output: "--- FAIL: TestFailing (0.00s)\n"},
			TestEvent{Action: "fail", Package: "tst", Test: "TestFailing", Elapsed: 0},
			TestEvent{Action: "run", Package: "tst", Test: "TestGood"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "=== RUN   TestGood\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "this one should be printed\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "because it's successful\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "--- PASS: TestGood (0.00s)\n"},
			TestEvent{Action: "pass", Package: "tst", Test: "TestGood", Elapsed: 0},
			TestEvent{Action: "output", Package: "tst", Output: "FAIL\n"},
			TestEvent{Action: "output", Package: "tst", Output: "coverage: 50.0% of statements\n"},
			TestEvent{Action: "output", Package: "tst", Output: "exit status 1\n"},
			TestEvent{Action: "output", Package: "tst", Output: "FAIL\ttst\t0.108s\n"},
			TestEvent{Action: "fail", Package: "tst", Elapsed: 0.108},
		)

		assert.Equal(t, []string{
			"✖ TestFailing",
			"if you having correctness problems i feel bad for you son",
			"i've got 99 problems",
			"  code_test.go:12: but a test ain't one",
			"✔ TestGood",
			"this one should be printed",
			"because it's successful",
			"◼ tst    50.0%     0.108s",
			"-------------------------",
			"",
			"❯ Average Coverage: 50.00%",
			"❯ Pkgs: tested: 1    failed: 1    noTests: 0    excluded: 0",
			""}, out)
	})

	t.Run("ignored, no list", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}, IgnoredPackages: []string{"tst/ignored"}},

			TestEvent{Action: "run", Package: "tst", Test: "TestGood"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "=== RUN   TestGood\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "this line should be ignored, as we are not verbose\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "--- PASS: TestGood (0.00s)\n"},
			TestEvent{Action: "pass", Package: "tst", Test: "TestGood", Elapsed: 0},
			TestEvent{Action: "output", Package: "tst", Output: "PASS\n"},
			TestEvent{Action: "output", Package: "tst", Output: "coverage: 50.0% of statements\n"},
			TestEvent{Action: "output", Package: "tst", Output: "ok  \ttst\t0.185s\n"},
			TestEvent{Action: "pass", Package: "tst", Elapsed: 0.186},
		)

		assert.Equal(t, []string{
			"✔ tst    50.0%     0.186s",
			"",
			"❯ Average Coverage: 50.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 1",
			"",
		}, out)
	})

	t.Run("ignored, list", func(t *testing.T) {
		out := runTests(
			&processOutputParams{ToTestPackages: []string{"tst"}, IgnoredPackages: []string{"tst/ignored"}, FlagListIgnored: true},

			TestEvent{Action: "run", Package: "tst", Test: "TestGood"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "=== RUN   TestGood\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "this line should be ignored, as we are not verbose\n"},
			TestEvent{Action: "output", Package: "tst", Test: "TestGood", Output: "--- PASS: TestGood (0.00s)\n"},
			TestEvent{Action: "pass", Package: "tst", Test: "TestGood", Elapsed: 0},
			TestEvent{Action: "output", Package: "tst", Output: "PASS\n"},
			TestEvent{Action: "output", Package: "tst", Output: "coverage: 50.0% of statements\n"},
			TestEvent{Action: "output", Package: "tst", Output: "ok  \ttst\t0.185s\n"},
			TestEvent{Action: "pass", Package: "tst", Elapsed: 0.186},
		)

		assert.Equal(t, []string{
			"✔ tst    50.0%     0.186s",
			"",
			"❯ Average Coverage: 50.00%",
			"❯ Pkgs: tested: 1    failed: 0    noTests: 0    excluded: 1",
			"",
			"Packages ignored:",
			"- tst/ignored",
		}, out)
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
