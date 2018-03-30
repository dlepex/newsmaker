// nolint
package news

import (
	"sync"
	"sync/atomic"
)

// Guard implements "try-lock pattern" and abstracts CAS operation from user code.
type Guard struct {
	locked int32
	noCopy sync.Mutex // for linting purpose only
}

// CanLock aka trylock
func (g *Guard) CanLock() bool {
	return atomic.CompareAndSwapInt32(&g.locked, 0, 1)
}

func (g *Guard) Unlock() {
	atomic.StoreInt32(&g.locked, 0)
}

func GoWG(wg *sync.WaitGroup, fn func()) { //nolint:golint
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

// Go - locks guard and goes fn if guard is not locked
func (g *Guard) Go(fn func()) bool {
	if !g.CanLock() {
		return false
	}
	go func() {
		defer g.Unlock()
		fn()
	}()
	return true
}
