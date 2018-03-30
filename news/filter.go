package news

import (
	"strings"

	"github.com/dlepex/newsmaker/words"
)

type Filter struct { //nolint
	Cond    string
	Sources []string // this are "globs" (glob is a simplified pattern, right now it's either prefix or suffix match)
	Pubs    []string // same

	dnf  *words.Expr // news title match condition
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

// matchAnyGlob matches string s against "globs"
// true, if glob is either prefix or suffix of s.
// todo use glob instead of simple prefix/suffix match?
func matchAnyGlob(s string, globs []string) bool {
	if len(globs) == 0 {
		return true
	}

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
