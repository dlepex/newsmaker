package news

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/willf/bitset"
)

var RotatorTickDefault time.Duration = 60 * time.Second

type Rotator struct {
	Tick  time.Duration
	Elems []RotatorElem
	guard Guard
}

type RotatorElem struct {
	Cooldown time.Duration
	// Fn should not panic.
	Fn   func(time.Time)
	last time.Time
}

func (rot *Rotator) Run(quit <-chan struct{}) {
	if !rot.guard.TryLock() {
		panic("rotator: already started")
	}
	if rot.Tick == 0 {
		rot.Tick = RotatorTickDefault
	}
	defer rot.guard.Unlock()
	slog.Infow("Rotator started", "elemCount", len(rot.Elems), "tick", rot.Tick)
	defer slog.Infow("Rotator finished", "elemCount", len(rot.Elems))
	if len(rot.Elems) == 0 {
		panic("rotator: no elements")
	}

	ready := make([]*RotatorElem, 0, len(rot.Elems))
	for {
		select {
		case <-time.After(rot.Tick):
			rot.onTick(ready, time.Now())
		case <-quit:
			return
		}
	}
}

func (rot *Rotator) onTick(ready []*RotatorElem, now time.Time) {
	ready = ready[:0]
	for i, _ := range rot.Elems {
		e := &rot.Elems[i]
		if now.Sub(e.last) >= e.Cooldown {
			ready = append(ready, e)
		}
	}
	var elem *RotatorElem
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

type DayInterval struct {
	set *bitset.BitSet
}

func checkHour(h int) {
	if h < 0 || h > 23 {
		panic("hour must belong 0..23")
	}
}

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

func (di DayInterval) ContainsHour(h int) bool {
	if di.set == nil {
		return false
	}
	return di.set.Test(uint(h))
}

func (di DayInterval) ContainsTime(t time.Time) bool {
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
