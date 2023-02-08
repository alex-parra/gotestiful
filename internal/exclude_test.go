package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExclude(t *testing.T) {
	t.Parallel()

	type tst struct {
		name string

		packages         []string
		excluded         []string
		expectedIncluded []string
		expectedIgnored  []string
	}

	tsts := []tst{{
		name:             "empty excludes",
		packages:         []string{"one", "two", "three"},
		excluded:         []string{},
		expectedIncluded: []string{"one", "two", "three"},
		expectedIgnored:  nil,
	}, {
		name:             "one exclude",
		packages:         []string{"one", "two", "three"},
		excluded:         []string{"two"},
		expectedIncluded: []string{"one", "three"},
		expectedIgnored:  []string{"two"},
	}, {
		name:             "prefix exclude",
		packages:         []string{"zero", "one/package", "one/other", "two", "not/one", "three"},
		excluded:         []string{"one"},
		expectedIncluded: []string{"zero", "two", "not/one", "three"},
		expectedIgnored:  []string{"one/package", "one/other"},
	},{
		name:             "regex",
		packages:         []string{"zero", "one/package", "one/other", "two/package", "two/other", "three"},
		excluded:         []string{"*/package"},
		expectedIncluded: []string{"zero", "one/other", "two/other", "three"},
		expectedIgnored:  []string{"one/package", "two/package"},
	}, {
		name:             "empty excludes - empty string ignored",
		packages:         []string{"one", "two", "three"},
		excluded:         []string{""},
		expectedIncluded: []string{"one", "two", "three"},
		expectedIgnored:  nil,
	}}

	for _, tst := range tsts {
		tst := tst
		t.Run(tst.name, func(t *testing.T) {
			t.Parallel()
			included, ignored := excludePackages(tst.packages, tst.excluded)
			assert.Equal(t, tst.expectedIncluded, included)
			assert.Equal(t, tst.expectedIgnored, ignored)
		})
	}
}
