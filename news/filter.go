package news

import (
	"errors"
	"strings"

	"github.com/dlepex/newsmaker/strext"
	"github.com/dlepex/newsmaker/words"
	"github.com/willf/bitset"
)

// DNF: or -> conj -> seq
type dnfCond [][][]words.Pattern

type Filter struct {
	Cond    string
	Sources []string
	Pubs    []string

	dnf  dnfCond
	pubs []string
}

func (f *Filter) init() error {
	dnf, e := parseCondition(f.Cond)
	if len(dnf) == 0 {
		return errors.New("Empty filter condition")
	}
	if e != nil {
		return e
	}
	f.dnf = dnf
	return nil
}

func parseCondition(s string) (dnfCond, error) {
	ors := strext.SplitAndTrimSpace(s, ";")
	dnf := make([][][]words.Pattern, 0, len(ors))
	for _, or := range ors {
		ands := strext.SplitAndTrimSpace(or, "&")
		conj := make([][]words.Pattern, 0, len(ands))
		for _, and := range ands {
			seq := strext.SplitAndTrimSpace(and, " ")
			ptrn := make([]words.Pattern, len(seq))
			for i, pstr := range seq {
				p, e := words.NewPattern(pstr)
				if e != nil {
					return nil, e
				}
				ptrn[i] = p
			}
			conj = append(conj, ptrn)
		}
		dnf = append(dnf, conj)
	}
	return dnf, nil
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

type wordMatcher struct {
	*Filter
	dnfState []*bitset.BitSet
	matched  bool
}

func (m *wordMatcher) init() {
	m.matched = false
	if len(m.dnfState) == 0 {
		m.dnfState = make([]*bitset.BitSet, len(m.dnf))
		for i, conj := range m.dnf {
			m.dnfState[i] = bitset.New(uint(len(conj)))
		}
	} else {
		for _, set := range m.dnfState {
			set.ClearAll()
		}
	}
}

func (m *wordMatcher) tryMatch(text []string) {
	if m.matched {
		return
	}
	for ci, conj := range m.dnf {
		for si, seq := range conj {
			if wordsPrefixMatch(text, seq) {
				set := m.dnfState[ci]
				set.Set(uint(si))
				if set.Count() == uint(len(conj)) {
					m.matched = true
				}
			}
		}
	}
}

func wordsPrefixMatch(text []string, seq []words.Pattern) bool {
	if len(text) < len(seq) {
		return false
	}
	for i := range seq {
		if !seq[i].Match(text[i]) {
			return false
		}
	}
	return true
}
