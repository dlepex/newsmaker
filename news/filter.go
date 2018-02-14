package news

import (
	"strings"

	"github.com/dlepex/newsmaker/words"
)

type Filter struct {
	Cond    string
	Sources []string
	Pubs    []string

	dnf  *words.Expr
	pubs []string
}

func (f *Filter) init() error {
	dnf, e := words.NewExpr(f.Cond)
	if e != nil {
		return e
	}
	f.dnf = dnf
	return nil
}

func matchAnyGlob(s string, globs []string) bool {
	if len(globs) == 0 {
		return true
	}
	//todo use glob instead of simple prefix/suffix match
	for _, g := range globs {
		if strings.HasPrefix(s, g) || strings.HasSuffix(s, g) {
			return true
		}
	}
	return false
}

func matchAnyGlobAny(ss []string, globs []string) bool {
	if len(ss) == 0 {
		return true
	}
	if len(globs) == 0 {
		return true
	}
	for _, s := range ss {
		if matchAnyGlob(s, globs) {
			return true
		}
	}
	return false
}

func (f *Filter) matchSrc(src *SourceInfo) bool {
	return matchAnyGlob(src.Name, f.Sources)
}

func (f *Filter) matchPub(pub *PubInfo) bool {
	return matchAnyGlob(pub.Name, f.Pubs)
}

func chooseFilters(ff []*Filter, match func(*Filter) bool) []int {
	ind := []int{}
	for i, f := range ff {
		if match(f) {
			ind = append(ind, i)
		}
	}
	return ind
}

func matchAnyPub(ff []*Filter, p *PubInfo) bool {
	for _, f := range ff {
		if f.matchPub(p) {
			return true
		}
	}
	return false
}
