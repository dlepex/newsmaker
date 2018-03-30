package words

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dlepex/newsmaker/sliceset"
	"github.com/dlepex/newsmaker/strext"
)

// Expr is a sentence filtering condition in a form of DNF of regex patterns.
// EBNF Grammar of Expr:
// --------------------------------------------
// Expr := Conj {";" Conj}
// Conj := Seq {"&" Seq}
// Seq := Pattern {" " Pattern}
// --------------------------------------------
// ; is OR, + is AND
// Seq is the sequence of patterns to match the sequence of words in the sentence
type Expr struct {
	elems []exprElem
	// sizes of conjuctions groups
	conjSizes []int
}

type exprElem struct {
	// pattern or sequence of patterns (seq)
	p []Pattern
	// what conjuctions groups this pattern belongs
	conj sliceset.Ints
}

// NewExpr - creates Expr from text (satisfying Expr grammar)
func NewExpr(s string) (*Expr, error) {
	ors := strext.SplitAndTrimSpace(s, ";")
	sort.Sort(strSlice(ors)) // evaluate shortest patterns first
	var elems []exprElem
	var seqs [][]string
	conj := 0
	var csizes []int
	for _, or := range ors {
		ands := strext.SplitAndTrimSpace(or, "&")
		isConj := len(ands) > 1
		if isConj {
			csizes = append(csizes, len(ands))
		}
		for _, and := range ands {
			seq := strext.SplitAndTrimSpace(and, " ")
			idx, prefix := indexOf(seqs, seq)

			if prefix {
				return nil, fmt.Errorf("Duplicate pattern seq prefix: %v -> %v", seq, seqs[idx])
			}

			if idx < 0 {
				seqs = append(seqs, seq)
				idx = len(seqs) - 1
				pseq, e := patterns(seq)
				if e != nil {
					return nil, e
				}
				var aconj []int
				if isConj {
					aconj = []int{conj}
				}
				elems = append(elems, exprElem{pseq, aconj})
			} else if isConj {
				e := &elems[idx]
				if len(e.conj) == 0 {
					return nil, fmt.Errorf("Found conj which is always true for pattern/seq: %v, please remove it", seq)
				}
				a := &elems[idx].conj
				*a = a.Append(conj)
			}
		}
		if isConj {
			conj++
		}
	}
	return &Expr{elems, csizes}, nil
}

//Match - matches expr against untokenized sentence
func (expr *Expr) Match(s string) bool {
	return expr.MatchWords(Split(s))
}

const mwStackSz = 128

//MatchWords - matches expr against tokenized sentence
func (expr *Expr) MatchWords(text []string) bool {
	var stack [mwStackSz]int16
	var kconj []int16

	conjNum := len(expr.conjSizes)
	if conjNum <= mwStackSz {
		kconj = stack[:conjNum]
	} else {
		kconj = make([]int16, conjNum)
	}

	for i, v := range expr.conjSizes {
		kconj[i] = int16(v)
	}
	var m map[int16]struct{}
	if len(kconj) > 0 {
		m = make(map[int16]struct{})
	}
	for w := range text {
		sub := text[w:]
		for idx, el := range expr.elems {
			if len(el.conj) == 0 {
				if el.matchSub(sub) {
					return true
				}
			} else {
				key := int16(idx)
				if _, ok := m[key]; ok {
					continue
				}
				if el.matchSub(sub) {
					m[key] = struct{}{}
					for _, ci := range el.conj {
						kconj[ci]--
						if kconj[ci] == 0 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (el *exprElem) matchSub(text []string) bool {
	if len(text) < len(el.p) {
		return false
	}
	for i, p := range el.p {
		if !p.Match(text[i]) {
			return false
		}
	}
	return true
}

func indexOf(patterns [][]string, val []string) (index int, prefix bool) {

	for i, seq := range patterns {
		match := true
		max := len(seq)
		for k, valk := range val {
			if k >= max {
				return i, true
			}
			if valk != seq[k] {
				match = false
				break
			}
		}
		if match {
			return i, false
		}
	}
	return -1, false
}

// order by string length.
type strSlice []string

func (p strSlice) Len() int { return len(p) }

func strWeight(s string) int {
	if strings.ContainsRune(s, '&') {
		return 1
	}
	return 0
}

func (p strSlice) Less(i, j int) bool {
	dif := strWeight(p[i]) - strWeight(p[j])
	if dif != 0 {
		return dif < 0
	}
	return len(p[i]) < len(p[j])
}

func (p strSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func patterns(seq []string) (ptrn []Pattern, err error) {
	ptrn = make([]Pattern, len(seq))
	for i, pstr := range seq {
		p, e := NewPattern(pstr)
		if e != nil {
			return nil, e
		}
		ptrn[i] = p
	}
	return
}
