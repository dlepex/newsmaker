package news

import (
	"sync"
	"sync/atomic"
)

type Guard struct {
	locked int32
	noCopy sync.Mutex
}

func (g *Guard) TryLock() bool {
	return atomic.CompareAndSwapInt32(&g.locked, 0, 1)
}

func (g *Guard) Unlock() {
	atomic.StoreInt32(&g.locked, 0)
}

func GoWG(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

func (g *Guard) Go(fn func()) bool {
	if !g.TryLock() {
		return false
	}
	go func() {
		defer g.Unlock()
		fn()
	}()
	return true
}