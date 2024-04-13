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

func SetEachNodeMaxSize(s int) {
	maxListPackSize = s
}

// New create a quicklist instance.
func New() *QuickList {
	lp := NewListPack()
	return &QuickList{head: lp, tail: lp}
}

func (l *QuickList) lpush(key string) {
	if len(l.head.data)+len(key) >= maxListPackSize {
		lp := NewListPack()
		lp.next = l.head
		l.head.prev = lp
		l.head = lp
	}
	l.head.LPush(key)
}

// LPush
func (l *QuickList) LPush(keys ...string) {
	l.mu.Lock()
	for _, k := range keys {
		l.lpush(k)
	}
	l.mu.Unlock()
}

func (l *QuickList) rpush(key string) {
	if len(l.tail.data)+len(key) >= maxListPackSize {
		lp := NewListPack()
		l.tail.next = lp
		lp.prev = l.tail
		l.tail = lp
	}
	l.tail.RPush(key)
}

// RPush
func (l *QuickList) RPush(keys ...string) {
	l.mu.Lock()
	for _, k := range keys {
		l.rpush(k)
	}
	l.mu.Unlock()
}

// Index
func (l *QuickList) Index(i int) (val string, ok bool) {
	l.Range(i, i+1, func(key []byte) bool {
		val, ok = string(key), true
		return true
	})
	return
}

// LPop
func (l *QuickList) LPop() (key string, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// remove empty head node
	for l.head.size == 0 {
		if l.head.next == nil {
			return
		}
		bpool.Put(l.head.data)
		l.head = l.head.next
		l.head.prev = nil
	}

	return l.head.LPop()
}

// RPop
func (l *QuickList) RPop() (key string, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// remove empty tail node
	for l.tail.size == 0 {
		if l.tail.prev == nil {
			return
		}
		bpool.Put(l.tail.data)
		l.tail = l.tail.prev
		l.tail.next = nil
	}

	return l.tail.RPop()
}

// Set
func (l *QuickList) Set(index int, key string) (ok bool) {
	l.mu.Lock()
	cur := l.head
	for index >= cur.Size() {
		index -= cur.Size()
		cur = cur.next
		if cur == nil {
			return
		}
	}
	ok = cur.Set(index, key)
	l.mu.Unlock()
	return
}

// Size
func (l *QuickList) Size() (n int) {
	l.mu.RLock()
	for cur := l.head; cur != nil; cur = cur.next {
		n += cur.Size()
	}
	l.mu.RUnlock()
	return
}

type lsIterator func(data []byte) (stop bool)

func (l *QuickList) iterFront(start, end int, f lsIterator) {
	count := end - start
	if end == -1 {
		count = math.MaxInt
	}
	if start < 0 || count < 0 {
		return
	}

	cur := l.head
	for start > cur.Size() {
		start -= cur.Size()
		cur = cur.next
		if cur == nil {
			return
		}
	}

	var stop bool
	for !stop && count > 0 && cur != nil {
		cur.Range(start, -1, func(data []byte) bool {
			stop = f(data)
			count--
			return stop || count == 0
		})
		cur = cur.next
		start = 0
	}
}

func (l *QuickList) iterBack(start, end int, f lsIterator) {
	count := end - start
	if end == -1 {
		count = math.MaxInt
	}
	if start < 0 || count < 0 {
		return
	}

	cur := l.tail
	for start > cur.Size() {
		start -= cur.Size()
		cur = cur.prev
		if cur == nil {
			return
		}
	}

	var stop bool
	for !stop && count > 0 && cur != nil {
		cur.RevRange(start, -1, func(data []byte) bool {
			stop = f(data)
			count--
			return stop || count == 0
		})
		cur = cur.prev
		start = 0
	}
}

// Range
func (l *QuickList) Range(start, end int, f lsIterator) {
	l.mu.RLock()
	l.iterFront(start, end, f)
	l.mu.RUnlock()
}

// RevRange
func (l *QuickList) RevRange(start, end int, f lsIterator) {
	l.mu.RLock()
	l.iterBack(start, end, f)
	l.mu.RUnlock()
}

type binListPack struct {
	E EncodeType
	N uint16
	D []byte
}

// MarshalJSON
func (l *QuickList) MarshalJSON() ([]byte, error) {
	data := make([]binListPack, 0)
	l.mu.RLock()
	defer l.mu.RUnlock()

	for cur := l.head; cur != nil; cur = cur.next {
		data = append(data, binListPack{
			E: cur.encode,
			N: cur.size,
			D: cur.data,
		})
	}
	return sonic.Marshal(data)
}

// UnmarshalJSON
func (l *QuickList) UnmarshalJSON(src []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var data []binListPack
	if err := sonic.Unmarshal(src, &data); err != nil {
		return err
	}

	var last *ListPack
	for _, item := range data {
		lp := &ListPack{
			encode: item.E,
			size:   item.N,
			data:   item.D,
			prev:   last,
		}
		if last != nil {
			last.next = lp
		}
		if l.head == nil {
			l.head = lp
		}
		l.tail = lp
		last = lp
	}

	return nil
}
