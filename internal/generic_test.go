package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type aStruct struct {
	one bool
	two string
}

func TestSF(t *testing.T) {
	t.Run("with no args", func(t *testing.T) {
		expected := "unchanged 45%s"
		actual := sf(expected)
		assert.Equal(t, expected, actual)
	})

	t.Run("with args", func(t *testing.T) {
		expected := "unchanged 45$"
		actual := sf("unchanged 45%s", "$")
		assert.Equal(t, expected, actual)
	})
}

func TestIsZeroValue(t *testing.T) {

	t.Run("zero-value string", func(t *testing.T) {
		assert.True(t, isZeroValue(""))
	})

	t.Run("non zero-value string", func(t *testing.T) {
		assert.False(t, isZeroValue("some string"))
	})

	t.Run("zero-value int", func(t *testing.T) {
		assert.True(t, isZeroValue(0))
	})

	t.Run("non zero-value int", func(t *testing.T) {
		assert.False(t, isZeroValue(123))
	})

	t.Run("zero-value struct", func(t *testing.T) {
		assert.True(t, isZeroValue(aStruct{}))
	})

	t.Run("non zero-value struct", func(t *testing.T) {
		assert.False(t, isZeroValue(aStruct{one: true}))
		assert.False(t, isZeroValue(aStruct{two: "some string"}))
	})
}

func TestZvfb(t *testing.T) {
	t.Run("with zero-value", func(t *testing.T) {
		assert.Equal(t, "oops", zvfb("", "oops"))
		assert.Equal(t, 123, zvfb(0, 123))
		assert.Equal(t, aStruct{one: true, two: ""}, zvfb(aStruct{}, aStruct{one: true}))
	})

	t.Run("with non zero-value", func(t *testing.T) {
		assert.Equal(t, "some string", zvfb("some string", "oops"))
		assert.Equal(t, 987, zvfb(987, 123))
		assert.Equal(t, aStruct{two: "hello"}, zvfb(aStruct{one: false, two: "hello"}, aStruct{one: true}))
	})
}

func TestIfelse(t *testing.T) {
	t.Run("with cond false", func(t *testing.T) {
		expected := "else"
		actual := ifelse(1 < 0, "if", "else")
		assert.Equal(t, expected, actual)
	})

	t.Run("with cond true", func(t *testing.T) {
		expected := "if"
		actual := ifelse(1 < 2, "if", "else")
		assert.Equal(t, expected, actual)
	})
}
