package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapHasKey(t *testing.T) {
	t.Run("with non existent keys", func(t *testing.T) {
		m := map[string]bool{"hello": true, "world": true}
		actual := mapHasKey(m, "foo")
		assert.False(t, actual)
	})

	t.Run("with existent keys", func(t *testing.T) {
		m := map[string]bool{"hello": true, "world": true}
		actual := mapHasKey(m, "hello")
		assert.True(t, actual)
	})
}
