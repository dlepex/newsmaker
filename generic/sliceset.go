// nolint
package generic

type E interface{} //typeinst: typevar
type SliceSet []E

func (s SliceSet) IndexOf(el E) int {
	for i, v := range s {
		if el == v {
			return i
		}
	}
	return -1
}

func (s SliceSet) Contains(el E) bool { return s.IndexOf(el) >= 0 }

func (s SliceSet) Append(el E) SliceSet {
	if s.IndexOf(el) >= 0 {
		return s
	}
	return append(s, el)
}
