package quicklist

import (
	"encoding/binary"
	"math"
	"sync"
)

// List is double linked ziplist.
type List struct {
	mu         sync.RWMutex
	head, tail *ListPack
}

func SetEachNodeMaxSize(s int) {
	maxListPackSize = s
}

// NewList
func NewList() *List {
	lp := NewListPack()
	return &List{head: lp, tail: lp}
}

func (l *List) lpush(key string) {
	if len(l.head.data)+len(key) >= maxListPackSize {
		lp := NewListPack()
		lp.next = l.head
		l.head.prev = lp
		l.head = lp
	}
	l.head.LPush(key)
}

// LPush
func (l *List) LPush(keys ...string) {
	l.mu.Lock()
	for _, k := range keys {
		l.lpush(k)
	}
	l.mu.Unlock()
}

func (l *List) rpush(key string) {
	if len(l.tail.data)+len(key) >= maxListPackSize {
		lp := NewListPack()
		l.tail.next = lp
		lp.prev = l.tail
		l.tail = lp
	}
	l.tail.RPush(key)
}

// RPush
func (l *List) RPush(keys ...string) {
	l.mu.Lock()
	for _, k := range keys {
		l.rpush(k)
	}
	l.mu.Unlock()
}

// Index
func (l *List) Index(i int) (val string, ok bool) {
	l.Range(i, i+1, func(key string) bool {
		val, ok = key, true
		return true
	})
	return
}

// LPop
func (l *List) LPop() (key string, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// remove empty head node
	for l.head.size == 0 {
		if l.head.next == nil {
			return
		}
		if cap(l.head.data) != maxListPackSize {
			panic("bytes cap not equal")
		}
		bpool.Put(l.head.data)
		l.head = l.head.next
		l.head.prev = nil
	}

	return l.head.LPop()
}

// RPop
func (l *List) RPop() (key string, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// remove empty tail node
	for l.tail.size == 0 {
		if l.tail.prev == nil {
			return
		}
		if cap(l.tail.data) != maxListPackSize {
			panic("bytes cap not equal")
		}
		bpool.Put(l.tail.data)
		l.tail = l.tail.prev
		l.tail.next = nil
	}

	return l.tail.RPop()
}

// Delete
// func (l *List) Delete(index int) (key string, ok bool) {
// 	l.mu.Lock()
// 	l.iter(index, index+1, func(node *lnode, dataStart, dataEnd int, bkey []byte) bool {
// 		key = string(bkey)
// 		node.data = slices.Delete(node.data, dataStart, dataEnd)
// 		node.n--
// 		ok = true
// 		return true
// 	})
// 	l.mu.Unlock()
// 	return
// }

// Set
// func (l *List) Set(index int, key string) (ok bool) {
// l.mu.Lock()
// l.iter(index, index+1, func(node *lnode, dataStart, dataEnd int, _ []byte) bool {
// 	alloc := bpool.Get(len(key) + 5)[:0]
// 	alloc = append(
// 		binary.AppendUvarint(alloc, uint64(len(key))),
// 		key...,
// 	)
// 	node.data = slices.Replace(node.data, dataStart, dataEnd, alloc...)
// 	bpool.Put(alloc)
// 	ok = true
// 	return true
// })
// l.mu.Unlock()
// return
// }

// Size
func (l *List) Size() (n int) {
	l.mu.RLock()
	for cur := l.head; cur != nil; cur = cur.next {
		n += cur.Size()
	}
	l.mu.RUnlock()
	return
}

func (l *List) iterFront(start, end int, f lpIterator) {
	// param check
	count := end - start
	if end == -1 {
		count = math.MaxInt
	}
	if start < 0 || count < 0 {
		return
	}

	cur := l.head
	// skip nodes
	for start > cur.Size() {
		start -= cur.Size()
		cur = cur.next
		if cur == nil {
			return
		}
	}

	var stop bool
	for !stop && count > 0 && cur != nil {
		cur.iterFront(start, -1, func(data []byte, entryStartPos, entryEndPos int) bool {
			// if start
			stop = f(data, entryStartPos, entryEndPos)
			count--
			return stop || count == 0
		})
		cur = cur.next
		start = 0
	}
}

// Range
func (l *List) Range(start, end int, f func(string) (stop bool)) {
	l.mu.RLock()
	l.iterFront(start, end, func(key []byte, _, _ int) bool {
		return f(string(key))
	})
	l.mu.RUnlock()
}

// Keys
func (l *List) Keys() (keys []string) {
	l.Range(0, -1, func(key string) bool {
		keys = append(keys, key)
		return false
	})
	return
}

// Marshal
func (l *List) Marshal() []byte {
	buf := bpool.Get(maxListPackSize)[:0]
	l.mu.RLock()
	for cur := l.head; cur != nil; cur = cur.next {
		buf = append(buf, cur.data...)
	}
	l.mu.RUnlock()
	return buf
}

// Unmarshal requires an initialized List.
func (l *List) Unmarshal(src []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var index int
	for index < len(src) {
		// klen
		klen, n := binary.Uvarint(src[index:])
		index += n
		// key
		key := src[index : index+int(klen)]
		l.rpush(string(key))
		index += int(klen)
	}
	return nil
}
