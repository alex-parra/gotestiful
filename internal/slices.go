package internal

import (
	"strings"
	"unicode"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

// splitLines splits a string by new lines
func splitLines(str string) []string {
	str = strings.TrimRightFunc(str, unicode.IsSpace)
	return strings.Split(strings.ReplaceAll(str, "\r\n", "\n"), "\n")
}

// sliceAvg calculates the average of a list of numbers
func sliceAvg[T Number](nums []T) T {
	total := T(0.0)
	for _, n := range nums {
		total = total + n
	}

	if total == 0 {
		return total
	}

	return total / T(len(nums))
}

// sliceAt returns the value at index 'idx' or 'fallback' if out of range
func sliceAt[T any](lst []T, idx int, fallback T) T {
	if idx < 0 || len(lst) <= idx {
		return fallback
	}

	return lst[idx]
}

// sliceAppendIf appends 'args' if 'cond' is true
func sliceAppendIf[T any](cond bool, lst []T, args ...T) []T {
	if cond {
		return append(lst, args...)
	}
	return lst
}
