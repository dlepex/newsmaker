package news

import (
	"encoding/hex"
	"hash/fnv"
	"log"
	"sync"
)

// Deduplicator is used to filter out repeated elements by their key (unique id or hash)
// Keep - returns true if key was kept (i.e. key is new) and false if key is duplicate.
type Deduplicator interface {
	Keep(key DedupKey) bool
}

// DedupKeySize - number of bytes in dedupkey byte arr
const DedupKeySize = 16

// DedupKey is array due to GC reasons (map will not be scanned), although string keys may be faster to get
type DedupKey [DedupKeySize]byte

// memDedup is primitive LRU cache impl
type memDedup struct {
	m       map[DedupKey]struct{}
	q       []DedupKey
	MaxSize int
	qr      int
	qw      int
}

// syncDedup -  thread-safe wrapper
type syncDedup struct {
	dedup Deduplicator
	mu    sync.Mutex
}

//NewDedup - creates new in-memory deduplicator
func NewDedup(maxSize int) Deduplicator {
	if maxSize <= 0 {
		log.Fatalf("illegal maxSize %d", maxSize)
	}
	return &memDedup{make(map[DedupKey]struct{}), make([]DedupKey, maxSize), maxSize, 0, 0}
}

//DedupSync returns concurrent-safe (mutex-based) wrapper
//if already wrapped does nothing.
func DedupSync(d Deduplicator) Deduplicator {
	if d == nil {
		panic("wrapping nil dedup")
	}
	if _, ok := d.(*syncDedup); ok {
		return d
	}
	return &syncDedup{dedup: d}
}

func (d *memDedup) Keep(k DedupKey) bool {
	m := d.m
	if _, has := m[k]; has {
		return false
	}
	max := d.MaxSize - 1
	if len(m) > max {
		oldk := d.q[getAndInc(&d.qr, max)]
		delete(m, oldk)
		m[k] = struct{}{}
		d.q[getAndInc(&d.qw, max)] = k
	} else {
		m[k] = struct{}{}
		d.q[getAndInc(&d.qw, max)] = k
	}
	return true
}

func (d *syncDedup) Keep(k DedupKey) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dedup.Keep(k)
}

func getAndInc(ptr *int, max int) int {
	v := *ptr
	if v < max {
		*ptr = v + 1
		return v
	}
	*ptr = 0
	return max

}

//StrToDedupKey - calculates dedupKey for strings (by hashing)
func StrToDedupKey(xs ...string) DedupKey {
	if len(xs) == 0 {
		return DedupKey{}
	}
	h := fnv.New128()
	h.Write([]byte(xs[0])) // nolint:errcheck
	for _, x := range xs[1:] {
		h.Write([]byte{0}) // nolint:errcheck
		h.Write([]byte(x)) // nolint:errcheck
	}

	var k DedupKey
	hash := h.Sum(nil)
	copy(k[:], hash)
	return k
}

func (k DedupKey) String() string {
	return hex.EncodeToString(k[:])
}
