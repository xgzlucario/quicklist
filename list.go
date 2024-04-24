package quicklist

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"
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
	ls.head.Insert(0, key)
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
	ls.tail.Insert(-1, key)
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
			return lp.Remove(-1)
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

// RemoveFirst
func (ls *QuickList) RemoveFirst(key string) (res int, ok bool) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	for lp := ls.head; lp != nil; lp = lp.next {
		if lp.size == 0 {
			ls.free(lp)

		} else {
			n, ok := lp.RemoveFirst(key)
			if ok {
				return res + n, true
			} else {
				res += lp.Size()
			}
		}
	}
	return 0, false
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
		lp.iterFront(indexInternal, -1, func(data []byte, _, _, _ int) bool {
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
		lp.iterBack(start, -1, func(data []byte, _, _, _ int) bool {
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

var order = binary.LittleEndian

// MarshalBinary
func (ls *QuickList) MarshalBinary() ([]byte, error) {
	data := bpool.Get(1024)[:0]
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	for lp := ls.head; lp != nil && lp.size > 0; lp = lp.next {
		// append [size, len_data, data]
		data = order.AppendUint32(data, lp.size)
		data = order.AppendUint32(data, uint32(len(lp.data)))
		data = append(data, lp.data...)
	}
	return data, nil
}

var ErrOutOfRange = errors.New("unmarshal error: index out of range")

// UnmarshalBinary
func (ls *QuickList) UnmarshalBinary(src []byte) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.head = nil
	var last *ListPack

	for index := 0; index < len(src); {
		if len(src)-index < 8 {
			return ErrOutOfRange
		}
		// size
		size := order.Uint32(src[index:])
		index += 4
		// length
		length := order.Uint32(src[index:])
		index += 4
		// data
		if index+int(length) > len(src) {
			return ErrOutOfRange
		}
		data := src[index : index+int(length)]
		index += int(length)

		lp := &ListPack{
			size: size,
			data: data,
			prev: last,
		}
		if ls.head == nil {
			ls.head = lp
		}
		if last != nil {
			last.next = lp
		}
		ls.tail = lp
		last = lp
	}
	return nil
}
