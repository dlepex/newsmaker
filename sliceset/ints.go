package sliceset

type Ints []int

func (s Ints) IndexOf(el int) int {
	for i, v := range s {
		if el == v {
			return i
		}
	}
	return -1
}

func (s Ints) Contains(el int) bool { return s.IndexOf(el) >= 0 }

func (s Ints) Append(el int) Ints {
	if s.IndexOf(el) >= 0 {
		return s
	}
	return append(s, el)
}
