package news

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDedupKey(t *testing.T) {
	assert.NotEqual(t, StrToDedupKey("wo", "rld"), StrToDedupKey("worl", "d"))
}

func TestDedup(t *testing.T) {
	d := NewDedup(4)
	keep := func(s string) bool {
		return d.Keep(StrToDedupKey(s))
	}
	s := []string{"11", "bb", "cc", "22", "zz"}
	s0 := s[0]
	assert.True(t, keep(s0))
	assert.False(t, keep(s0))
	for _, v := range s[1:] {
		assert.True(t, keep(v))
	}
	assert.False(t, keep(s[2]))
}
