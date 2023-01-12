package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitLines(t *testing.T) {
	expected := []string{"Line 1", "Line 2", "Line 3", "Line 4"}
	actual := splitLines("Line 1\nLine 2\nLine 3\r\nLine 4")
	assert.Equal(t, expected, actual)
}

func TestSliceAvg(t *testing.T) {
	t.Run("with ints", func(t *testing.T) {
		expected := 4
		actual := sliceAvg([]int{1, 2, 6, 7})
		assert.Equal(t, expected, actual)
	})

	t.Run("with floats", func(t *testing.T) {
		expected := 4.0
		actual := sliceAvg([]float64{1.5, 2, 6, 6.5})
		assert.Equal(t, expected, actual)
	})

	t.Run("with empty list", func(t *testing.T) {
		expected := 0
		actual := sliceAvg([]int{})
		assert.Equal(t, expected, actual)
	})
}

func TestSliceAt(t *testing.T) {
	t.Run("returns fallback with negative index", func(t *testing.T) {
		expected := "oops"
		actual := sliceAt([]string{}, -2, "oops")
		assert.Equal(t, expected, actual)
	})

	t.Run("returns fallback with index out-of-range", func(t *testing.T) {
		expected := "oops"
		actual := sliceAt([]string{}, 0, "oops")
		assert.Equal(t, expected, actual)
	})

	t.Run("returns value with index ok", func(t *testing.T) {
		expected := "ok"
		actual := sliceAt([]string{"ok"}, 0, "oops")
		assert.Equal(t, expected, actual)
	})
}

func TestSliceAppendIf(t *testing.T) {
	t.Run("with condition false", func(t *testing.T) {
		expected := []string{"one"}
		actual := sliceAppendIf(1 < 0, []string{"one"}, "two")
		assert.Equal(t, expected, actual)
	})

	t.Run("with condition true", func(t *testing.T) {
		expected := []string{"one", "two"}
		actual := sliceAppendIf(1 < 2, []string{"one"}, "two")
		assert.Equal(t, expected, actual)
	})
}

func TestSliceExclude(t *testing.T) {
	t.Run("with empty excludes", func(t *testing.T) {
		expected := []string{"one", "two", "three"}
		actual := sliceExclude([]string{"one", "two", "three"}, []string{})
		assert.Equal(t, expected, actual)
	})

	t.Run("with one excludes", func(t *testing.T) {
		expected := []string{"one", "three"}
		actual := sliceExclude([]string{"one", "two", "three"}, []string{"two"})
		assert.Equal(t, expected, actual)
	})

	t.Run("with multiple excludes", func(t *testing.T) {
		expected := []string{"one", "three", "five"}
		actual := sliceExclude([]string{"one", "two", "three", "four", "five"}, []string{"two", "four"})
		assert.Equal(t, expected, actual)
	})
}
