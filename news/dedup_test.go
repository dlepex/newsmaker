package news

import (
	"hash/fnv"
	"testing"
)

func x(s string) []byte {
	h := fnv.New128()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func TestDedupKey(t *testing.T) {
	if StrToDedupKey("wo", "rld") == StrToDedupKey("worl", "d") {
		t.Error()
	}
}
func TestDedup(t *testing.T) {
	d := NewDedup(4)
	K := func(s string) DedupKey {
		return StrToDedupKey(s)
	}
	s := []string{"11", "bb", "cc", "22", "zz"}

	if d.Check(K(s[0])) {
		t.Error()
	}
	if !d.Check(K(s[0])) {
		t.Error()
	}
	dlog := func() {
		t.Log(d.qr, d.qw, d.q[0:])
	}
	dlog()
	for _, v := range s[1:] {
		if d.Check(K(v)) {
			t.Error()
		}
		dlog()
	}

	if len(d.m) != d.MaxSize {
		t.Error()
	}
	if d.Check(K(s[0])) {
		t.Error()
	}
	if d.Check(K(s[1])) {
		t.Error()
	}
	if !d.Check(K(s[3])) {
		t.Error()
	}
	dlog()
}
