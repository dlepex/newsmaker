package news

import (
	"encoding/hex"
	"hash/fnv"
	"log"
	"sync"
)

const DedupKeySize = 16

// DedupKey is array due to GC reasons (map will not be scanned), although string keys may be faster to get
type DedupKey [DedupKeySize]byte

// Deduplicator is very primitive LRU cache impl
type Deduplicator struct {
	m       map[DedupKey]struct{}
	q       []DedupKey
	MaxSize int
	qr      int
	qw      int
}

type SyncDeduplicator struct {
	*Deduplicator
	mu sync.Mutex
}

func NewDedup(maxSize int) *Deduplicator {
	if maxSize <= 0 {
		log.Fatalf("illegal maxSize %d", maxSize)
	}
	return &Deduplicator{make(map[DedupKey]struct{}), make([]DedupKey, maxSize), maxSize, 0, 0}
}

// Checks if key is duplicate (true), if not - adds it to the cache (false)
func (d *Deduplicator) Check(k DedupKey) bool {
	m := d.m
	if _, has := m[k]; has {
		return true
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
	return false
}

func (d *SyncDeduplicator) Check(k DedupKey) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Deduplicator.Check(k)
}

func getAndInc(ptr *int, max int) int {
	v := *ptr
	if v < max {
		*ptr = v + 1
		return v
	} else {
		*ptr = 0
		return max
	}
}

func StrToDedupKey(xs ...string) DedupKey {
	if len(xs) == 0 {
		return DedupKey{}
	}
	h := fnv.New128()
	h.Write([]byte(xs[0]))
	for _, x := range xs[1:] {
		h.Write([]byte{0})
		h.Write([]byte(x))
	}

	var k DedupKey
	hash := h.Sum(nil)
	copy(k[:], hash)
	return k
}

func (k DedupKey) String() string {
	return hex.EncodeToString(k[:])
}
