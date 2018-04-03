package words

import (
	"sort"
	"strings"
	"testing"

	"github.com/dlepex/newsmaker/strext"
)

func TestSplit(t *testing.T) {
	given := `«Cобака-23»   с начала года провела: на 18,94% больше теле-шоу о NNN.!!! В размере до 10.007 P&G`
	expected := `Cобака-23|с|начала|года|провела|на|18,94|больше|теле-шоу|о|NNN|В|размере|до|10.007|P&G`
	get := strings.Join(Split(given), "|")
	if get != expected {
		t.Error(get)
	}

	given = ` ,:!. ... ,. `
	if len(Split(given)) != 0 {
		t.Error("empty slice expected")
	}
}

func TestPattern(t *testing.T) {

	tests := [][]string{
		{`S(\x26)P`, "S&P", "!S%P", "!s&p", "!sp", "S&Pxxx", "!xxxxS&P"},
		{"*tion", "situation", "Tion", "lllTiONrr"},
		{"*tion$", "situation", "!tiona"},
		{"tion", "!situation", "tiona"},
		{"ноч$ь", "ночью"},
		{"д.ч(ер)?$Ь", "дичью", "дочерях", "дочь", "Д1ЧЕРЯМ"},
	}

	testMatchers(t, tests, func(s string) (matcher, error) {
		return NewPattern(s)
	})
}

func TestExpr(t *testing.T) {
	tests := [][]string{
		{`hello dot;world & s(\x26)p;*tion$ xa & in$ & Ferr`, "xx S&P world xx", "xx world xx s&P", "hello DoT", "!dot hello", "sition xai in FERR",
			"!Ferr xai in sitution", "Ferr in tion XA", "!Ferr tion XA"},
		{"aa;bbb;ccc", "xxx aa xxx", "AAa", "!xxxBCxx", "BbBbbb sss s s s"},
		{"aa & bb; cc & bb", "BB AA", "CCC Bbb", "!aaa ccccc"},
	}

	testMatchers(t, tests, func(s string) (matcher, error) {
		expr, err := NewExpr(s)
		if err == nil {
			t.Log(expr.conjSizes, expr.elems)
		}
		return expr, err
	})
}

type matcher interface {
	Match(string) bool
}

func testMatchers(t *testing.T, tests [][]string, New func(string) (matcher, error)) {
	for i, test := range tests {
		p, e := New(test[0])
		if e != nil {
			t.Errorf("New error: [%v] in %d", e, i)
			continue
		}
		for _, sample := range test[1:] {
			not := (sample[0] == '!')
			if not {
				if p.Match(sample[1:]) {
					t.Errorf("[%s] must NOT match %s", test[0], sample)
				}
			} else {
				if !p.Match(sample) {
					t.Errorf("[%s] should match  %s", test[0], sample)
				}
			}
		}
	}
}

func TestSort(t *testing.T) {
	s := "aa&v;aaaaaaaaaa;z&a&x&d;bbbsssssssssb;cxx;b&v;11"
	ors := strext.SplitAndTrimSpace(s, ";")
	t.Log(ors)
	sort.Sort(strSlice(ors))
	t.Log(ors)
}
