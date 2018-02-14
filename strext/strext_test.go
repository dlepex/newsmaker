package strext

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitAndTrimSpace(t *testing.T) {
	s := "this, was so,, nice,\tof,  ,, , you \t "
	es := []string{"this", "was so", "nice", "of", "you"}
	fs := SplitAndTrimSpace(s, ",")

	if !reflect.DeepEqual(fs, es) {
		t.Errorf("Expected %v found %v", es, fs)
	}
}

func TestIsBlank(t *testing.T) {
	xs := []string{" \t ", "  s  ", "", " "}

	for i, x := range xs {
		if IsBlank(x) != (strings.TrimSpace(x) == "") {
			t.Errorf("Equivalence failed at %d", i)
		}
	}
}

func BenchmarkIsBlank(b *testing.B) {
	benchStrPred(b, func(s string) bool {
		return IsBlank(s)
	})
}

func BenchmarkIsBlankTs(b *testing.B) {
	benchStrPred(b, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
}

func BenchmarkEmptyStr(b *testing.B) {
	benchStrPred(b, func(s string) bool {
		return s == ""
	})
}

func benchStrPred(b *testing.B, pred func(string) bool) {
	xs := []string{" \t ", "  s  ", "", " "}
	l := len(xs)
	// run the Fib function b.N times
	sum := 0
	for n := 0; n < b.N; n++ {
		i := n % l
		if pred(xs[i]) {
			sum++
		}
	}
	b.Log(sum)
}
