package quicklist

import (
	"bytes"
	"encoding/binary"
	"slices"
)

var (
	defaultListPackCap = 128

	maxListPackSize = 1024 * 1024

	bpool = NewBufferPool()
)

// ListPack is a lists of strings serialization format on Redis.
/*
	ListPack data content:
	+--------+--------+-----+--------+
	| entry0 | entry1 | ... | entryN |
	+--------+--------+-----+--------+
	    |
	  entry0 content:
	+------------+--------------+---------------------+
	|  data_len  |     data     |      entry_len      |
	+------------+--------------+---------------------+
	|<- varint ->|<- data_len ->|<- varint(reverse) ->|
	|<------- entry_len ------->|

	Using this structure, it is fast to iterate from both sides.
*/
type ListPack struct {
	size       uint32
	data       []byte
	prev, next *ListPack
}

func NewListPack() *ListPack {
	return &ListPack{data: make([]byte, 0, defaultListPackCap)}
}

func (lp *ListPack) RPush(data string) {
	lp.data = appendEntry(lp.data, data)
	lp.size++
}

func (lp *ListPack) LPush(data string) {
	entry := appendEntry(nil, data)
	lp.data = slices.Insert(lp.data, 0, entry...)
	lp.size++
	// reuse
	bpool.Put(entry)
}

func (lp *ListPack) RPop() (string, bool) {
	return lp.Remove(lp.Size() - 1)
}

func (lp *ListPack) LPop() (string, bool) {
	return lp.Remove(0)
}

func (lp *ListPack) Size() int {
	return int(lp.size)
}

// lpIterator is listpack iterator.
type lpIterator func(data []byte, entryStartPos, entryEndPos int) (stop bool)

func (lp *ListPack) iterFront(start, end int, f lpIterator) {
	if end == -1 {
		end = lp.Size()
	}
	var index int
	for i := 0; i < end && index < len(lp.data); i++ {
		//
		//    index     dataStartPos    dataEndPos            indexNext
		//      |            |              |                     |
		//      +------------+--------------+---------------------+-----+
		//  --> |  data_len  |     data     |      entry_len      | ... |
		//      +------------+--------------+---------------------+-----+
		//      |<--- n ---->|<- data_len ->|<-- size_entry_len ->|
		//
		dataLen, n := binary.Uvarint(lp.data[index:])
		indexNext := index + n + int(dataLen) + +SizeUvarint(dataLen+uint64(n))

		if i >= start {
			dataStartPos := index + n
			dataEndPos := dataStartPos + int(dataLen)

			data := lp.data[dataStartPos:dataEndPos]
			if f(data, index, indexNext) {
				return
			}
		}
		index = indexNext
	}
}

func (lp *ListPack) iterBack(start, end int, f lpIterator) {
	if end == -1 {
		end = lp.Size()
	}
	var index = len(lp.data)
	for i := 0; i < end && index > 0; i++ {
		//
		//    indexNext  dataStartPos    dataEndPos               index
		//        |            |              |                     |
		//  +-----+------------+--------------+---------------------+
		//  | ... |  data_len  |     data     |      entry_len      | <--
		//  +-----+------------+--------------+---------------------+
		//        |<--- n ---->|<- data_len ->|<-- size_entry_len ->|
		//        |<------ entry_len -------->|
		//
		entryLen, sizeEntryLen := uvarintReverse(lp.data[:index])
		indexNext := index - int(entryLen) - sizeEntryLen

		if i >= start {
			dataLen, n := binary.Uvarint(lp.data[indexNext:])
			dataStartPos := indexNext + n
			dataEndPos := dataStartPos + int(dataLen)

			data := lp.data[dataStartPos:dataEndPos]
			if f(data, indexNext, index) {
				return
			}
		}
		index = indexNext
	}
}

// find quickly locates the element based on index.
// When the target index is in the first half, use forward traversal;
// otherwise, use reverse traversal.
func (lp *ListPack) find(index int, fn func(old []byte, entryStartPos, entryEndPos int)) {
	if lp.size == 0 || index >= lp.Size() {
		return
	}
	if index <= lp.Size()/2 {
		lp.iterFront(index, index+1, func(old []byte, entryStartPos, entryEndPos int) bool {
			fn(old, entryStartPos, entryEndPos)
			return true
		})
	} else {
		index = lp.Size() - index - 1
		lp.iterBack(index, index+1, func(old []byte, entryStartPos, entryEndPos int) bool {
			fn(old, entryStartPos, entryEndPos)
			return true
		})
	}
}

func (lp *ListPack) Set(index int, data string) (ok bool) {
	lp.find(index, func(old []byte, entryStartPos, entryEndPos int) {
		if len(data) == len(old) {
			copy(old, data)
		} else {
			alloc := appendEntry(nil, data)
			lp.data = slices.Replace(lp.data, entryStartPos, entryEndPos, alloc...)
			bpool.Put(alloc)
		}
		ok = true
	})
	return
}

func (lp *ListPack) Remove(index int) (val string, ok bool) {
	lp.find(index, func(data []byte, entryStartPos, entryEndPos int) {
		val = string(data)
		lp.data = slices.Delete(lp.data, entryStartPos, entryEndPos)
		lp.size--
		ok = true
	})
	return
}

func (lp *ListPack) RemoveRange(index, count int) (n int) {
	var start, end int
	var flag bool
	lp.iterFront(index, index+count, func(_ []byte, entryStartPos, entryEndPos int) bool {
		if !flag {
			start = entryStartPos
			flag = true
		}
		end = entryEndPos
		n++
		return false
	})
	lp.data = slices.Delete(lp.data, start, end)
	lp.size -= uint32(n)
	return
}

func (lp *ListPack) RemoveElem(data string) (ok bool) {
	lp.iterFront(0, -1, func(old []byte, entryStartPos, entryEndPos int) bool {
		if bytes.Equal(s2b(&data), old) {
			lp.data = slices.Delete(lp.data, entryStartPos, entryEndPos)
			lp.size--
			ok = true
		}
		return ok
	})
	return
}

// encode data to [data_len, data, entry_len].
func appendEntry(dst []byte, data string) []byte {
	if dst == nil {
		dst = bpool.Get(maxListPackSize)[:0]
	}
	before := len(dst)
	dst = appendUvarint(dst, len(data), false)
	dst = append(dst, data...)
	return appendUvarint(dst, len(dst)-before, true)
}
