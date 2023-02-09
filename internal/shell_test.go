package internal

import (
	"sync"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestShCmd(t *testing.T) {
	expected := "Hello\n"
	actual := shCmd("echo", shArgs{"Hello"}, "")
	assert.Equal(t, expected, actual)
}

func TestShJSONPipe(t *testing.T) {
	type Inp struct {
		Foo string
		Bar string
	}
	out := []Inp{}
	c := make(chan Inp)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for line := range c {
			out = append(out, line)
		}
		wg.Done()
	}()

	shJSONPipe("echo", shArgs{`{"Foo": "1",
		"Bar":"2"
	}
	{"Foo": "3", "Bar": "4"}
	{"Foo":"5"}{"Bar": "6"}`}, "", c)
	wg.Wait()

	assert.Equal(t, out, []Inp{
		{Foo: "1", Bar: "2"},
		{Foo: "3", Bar: "4"},
		{Foo: "5"},
		{Bar: "6"},
	})
}

func TestShColor(t *testing.T) {
	t.Run("no color", func(t *testing.T) {
		color.NoColor = true
		actual := shColor("red", "Hello")
		assert.Equal(t, "Hello", actual)
		color.NoColor = false
	})

	t.Run("unsupported color", func(t *testing.T) {
		actual := shColor("blank", "Hello")
		assert.Equal(t, "Hello", actual)
	})

	t.Run("red", func(t *testing.T) {
		actual := shColor("red", "Hello")
		assert.Contains(t, actual, "\033[31mHello")
	})

	t.Run("red:bold", func(t *testing.T) {
		actual := shColor("red:bold", "Hello")
		assert.Contains(t, actual, "\033[31m\033[1mHello")
	})

	t.Run("green", func(t *testing.T) {
		actual := shColor("green", "Hello")
		assert.Contains(t, actual, "\033[32mHello")
	})

	t.Run("yellow", func(t *testing.T) {
		actual := shColor("yellow", "Hello")
		assert.Contains(t, actual, "\033[33mHello")
	})

	t.Run("whitesmoke", func(t *testing.T) {
		actual := shColor("whitesmoke", "Hello %d", 123)
		assert.Contains(t, actual, "\033[38;2;180;180;180mHello 123")
	})

	t.Run("gray", func(t *testing.T) {
		actual := shColor("gray", "Hello %s", "World")
		assert.Contains(t, actual, "\033[38;2;85;85;85mHello World")
	})
}
