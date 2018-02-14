package news

import (
	"testing"

	"github.com/dlepex/newsmaker/words"
)

type A struct {
	val int
}

func (a *A) set() {
	a.val = 1
}

func TestMatch(t *testing.T) {
	ww := words.Split(`«Газпромом»  заявил о росте конкуренции из-за запуска «Ямал СПГ»`)

	f := &Filter{Cond: "*пр.м & спг$ &яма"}
	e := f.init()
	if e != nil {
		panic(e)
	}
	wm := wordMatcher{Filter: f}
	wm.init()

	for i, _ := range ww {
		if !wm.matched {
			wm.tryMatch(ww[i:])
		}
	}

	a := A{val: 2}
	a.set()
	t.Log(a)
	t.Log(wm)
}

// систем$f систем$m

// газпром$m

func checkb(e bool) {
	if !e {
		panic(e)
	}
}
