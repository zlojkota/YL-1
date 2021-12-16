package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOther(t *testing.T) {
	t.Run("no test", func(t *testing.T) {
		assert.True(t, true)
	})
}
