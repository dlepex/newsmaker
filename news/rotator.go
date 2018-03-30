package news

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/willf/bitset"
)

// rotTickDefault - default tick duration
var rotTickDefault = 60 * time.Second

// DayInterval is a (bit) set 0..23 of day hours.
type DayInterval struct {
	set *bitset.BitSet
}

type rotator struct {
	Tick  time.Duration
	Elems []rotatorElem
	guard Guard
}

type rotatorElem struct {
	Cooldown time.Duration
	// Fn should not panic.
	Fn   func(time.Time)
	last time.Time
}

func (rot *rotator) run(quit <-chan struct{}) {
	if !rot.guard.CanLock() {
		panic("rotator: already started")
	}
	defer rot.guard.Unlock()
	if rot.Tick == 0 {
		rot.Tick = rotTickDefault
	}
	slog.Infow("Rotator started", "elemCount", len(rot.Elems), "tick", rot.Tick)
	defer slog.Infow("Rotator finished", "elemCount", len(rot.Elems))
	if len(rot.Elems) == 0 {
		panic("rotator: no elements")
	}

	ready := make([]*rotatorElem, 0, len(rot.Elems))
	for {
		select {
		case <-time.After(rot.Tick):
			rot.onTick(ready, time.Now())
		case <-quit:
			return
		}
	}
}

func (rot *rotator) onTick(ready []*rotatorElem, now time.Time) {
	ready = ready[:0]
	for i := range rot.Elems {
		e := &rot.Elems[i]
		if now.Sub(e.last) >= e.Cooldown {
			ready = append(ready, e)
		}
	}
	var elem *rotatorElem
	switch len(ready) {
	case 0:
		return
	case 1:
		elem = ready[0]
	default:
		elem = ready[rand.Intn(len(ready))]
	}
	elem.last = now
	elem.Fn(now)
}

func checkHour(h int) {
	if !(0 <= h && h <= 23) {
		panic("hour must be in range: 0..23")
	}
}

// DayHoursFromTo creates DayInterval from begin(incl.) to end (incl.) hour
// begin, end are day hours 0..23
// end may be less that begin, since day hours are are cyclic :-)
func DayHoursFromTo(begin, end int) DayInterval {
	checkHour(begin)
	checkHour(end)
	set := bitset.New(24)
	b, e := uint(begin), uint(end)
	h := b
	for {
		set.Set(h)
		h++
		if h == 24 {
			h = 0
		}
		if h == b || h == e {
			break
		}
	}
	return DayInterval{set}
}

func (di DayInterval) ContainsHour(h int) bool { //nolint
	if di.set == nil {
		return false
	}
	return di.set.Test(uint(h))
}

func (di DayInterval) ContainsTime(t time.Time) bool { //nolint
	return di.ContainsHour(t.Hour())
}

func (di DayInterval) String() string {
	if di.set == nil {
		return "{}"
	}
	a := make([]int8, 0, di.set.Count())
	for i := 0; i < 24; i++ {
		if di.set.Test(uint(i)) {
			a = append(a, int8(i))
		}
	}
	return fmt.Sprintf("%v", a)
}
