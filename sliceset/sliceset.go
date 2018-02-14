package sliceset

type T interface{}
type SliceSet []T

func (s SliceSet) IndexOf(el T) int {
	for i, v := range s {
		if el == v {
			return i
		}
	}
	return -1
}

func (s SliceSet) Contains(el T) bool { return s.IndexOf(el) >= 0 }

func (s SliceSet) Append(el T) SliceSet {
	if s.IndexOf(el) >= 0 {
		return s
	}
	return append(s, el)
}
