package quicklist

import (
	"math"
	"sync"

	"github.com/bytedance/sonic"
)

//	 +------------------------------ QuickList -----------------------------+
//	 |	     +-----------+     +-----------+             +-----------+      |
//	head --- | listpack0 | <-> | listpack1 | <-> ... <-> | listpackN | --- tail
//	         +-----------+     +-----------+             +-----------+
//
// QuickList is double linked listpack.
type QuickList struct {
	mu         sync.RWMutex
	head, tail *ListPack
}

func SetMaxListPackSize(s int) {
	maxListPackSize = s
}

func SetDefaultListPackCap(s int) {
	defaultListPackCap = s
}

// New create a quicklist instance.
func New() *QuickList {
	lp := NewListPack()
	return &QuickList{head: lp, tail: lp}
}

func (ls *QuickList) lpush(key string) {
	if len(ls.head.data)+len(key) >= maxListPackSize {
		lp := NewListPack()
		lp.next = ls.head
		ls.head.prev = lp
		ls.head = lp
	}
	ls.head.LPush(key)
}

// LPush
func (ls *QuickList) LPush(keys ...string) {
	ls.mu.Lock()
	for _, k := range keys {
		ls.lpush(k)
	}
	ls.mu.Unlock()
}

func (ls *QuickList) rpush(key string) {
	if len(ls.tail.data)+len(key) >= maxListPackSize {
		lp := NewListPack()
		ls.tail.next = lp
		lp.prev = ls.tail
		ls.tail = lp
	}
	ls.tail.RPush(key)
}

// RPush
func (ls *QuickList) RPush(keys ...string) {
	ls.mu.Lock()
	for _, k := range keys {
		ls.rpush(k)
	}
	ls.mu.Unlock()
}

// Index
func (ls *QuickList) Index(i int) (val string, ok bool) {
	ls.Range(i, i+1, func(key []byte) bool {
		val, ok = string(key), true
		return true
	})
	return
}

// LPop
func (ls *QuickList) LPop() (string, bool) {
	return ls.Remove(0)
}

// RPop
func (ls *QuickList) RPop() (key string, ok bool) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	for lp := ls.tail; lp != nil; lp = lp.prev {
		if lp.size > 0 {
			return lp.RPop()
		}
		ls.free(lp)
	}
	return
}

// free release empty listpack.
func (ls *QuickList) free(lp *ListPack) {
	if lp.size == 0 && lp.prev != nil && lp.next != nil {
		lp.prev.next = lp.next
		lp.next.prev = lp.prev
		bpool.Put(lp.data)
		lp = nil
	}
}

// find quickly locates `listpack` and it `indexInternal` based on index.
func (ls *QuickList) find(index int) (*ListPack, int) {
	var lp *ListPack
	for lp = ls.head; lp != nil && index >= lp.Size(); lp = lp.next {
		index -= lp.Size()
	}
	return lp, index
}

// Set
func (ls *QuickList) Set(index int, key string) bool {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	lp, indexInternal := ls.find(index)
	if lp != nil {
		return lp.Set(indexInternal, key)
	}
	return false
}

// Remove
func (ls *QuickList) Remove(index int) (val string, ok bool) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	lp, indexInternal := ls.find(index)
	if lp != nil {
		val, ok = lp.Remove(indexInternal)
		ls.free(lp)
	}
	return
}

// RemoveRange
func (ls *QuickList) RemoveRange(index, count int) (n int) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	lp, indexInternal := ls.find(index)
	for n < count && lp != nil {
		n += lp.RemoveRange(indexInternal, count)
		ls.free(lp)
		indexInternal = 0
		lp = lp.next
	}
	return
}

// RemoveFirst
func (ls *QuickList) RemoveFirst(key string) bool {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	for lp := ls.head; lp != nil; lp = lp.next {
		if lp.size == 0 {
			ls.free(lp)
		} else if lp.RemoveElem(key) {
			return true
		}
	}
	return false
}

// Size
func (ls *QuickList) Size() (n int) {
	ls.mu.RLock()
	for lp := ls.head; lp != nil; lp = lp.next {
		n += lp.Size()
	}
	ls.mu.RUnlock()
	return
}

type lsIterator func(data []byte) (stop bool)

func (ls *QuickList) iterFront(start, end int, f lsIterator) {
	count := end - start
	if end == -1 {
		count = math.MaxInt
	}
	if start < 0 || count < 0 {
		return
	}

	lp, indexInternal := ls.find(start)

	var stop bool
	for !stop && count > 0 && lp != nil {
		lp.iterFront(indexInternal, -1, func(data []byte, _, _ int) bool {
			stop = f(data)
			count--
			return stop || count == 0
		})
		lp = lp.next
		indexInternal = 0
	}
}

func (ls *QuickList) iterBack(start, end int, f lsIterator) {
	count := end - start
	if end == -1 {
		count = math.MaxInt
	}
	if start < 0 || count < 0 {
		return
	}

	lp := ls.tail
	for start > lp.Size() {
		start -= lp.Size()
		lp = lp.prev
		if lp == nil {
			return
		}
	}

	var stop bool
	for !stop && count > 0 && lp != nil {
		lp.iterBack(start, -1, func(data []byte, _, _ int) bool {
			stop = f(data)
			count--
			return stop || count == 0
		})
		lp = lp.prev
		start = 0
	}
}

// Range
func (ls *QuickList) Range(start, end int, f lsIterator) {
	ls.mu.RLock()
	ls.iterFront(start, end, f)
	ls.mu.RUnlock()
}

// RevRange
func (ls *QuickList) RevRange(start, end int, f lsIterator) {
	ls.mu.RLock()
	ls.iterBack(start, end, f)
	ls.mu.RUnlock()
}

type binListPack struct {
	N uint32
	D []byte
}

// MarshalJSON
func (ls *QuickList) MarshalJSON() ([]byte, error) {
	data := make([]binListPack, 0, 8)
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	for lp := ls.head; lp != nil && lp.size > 0; lp = lp.next {
		data = append(data, binListPack{
			N: lp.size,
			D: lp.data,
		})
	}
	return sonic.Marshal(data)
}

// UnmarshalJSON
func (ls *QuickList) UnmarshalJSON(src []byte) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	var data []binListPack
	if err := sonic.Unmarshal(src, &data); err != nil {
		return err
	}

	var last *ListPack
	for i, item := range data {
		lp := &ListPack{
			size: item.N,
			data: item.D,
			prev: last,
		}
		if last != nil {
			last.next = lp
		}
		if i == 0 {
			ls.head = lp
		}
		ls.tail = lp
		last = lp
	}

	return nil
}
