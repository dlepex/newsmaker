package sliceset

import (
	"github.com/dlepex/newsmaker/generic"
)

//go:generate typeinst
type _typeinst struct { // nolint
	Ints func(E int) generic.SliceSet
}
