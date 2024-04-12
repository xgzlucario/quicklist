package quicklist

import (
	"encoding/binary"
	"slices"

	"github.com/klauspost/compress/zstd"
)

var (
	maxListPackSize = 8 * 1024

	bpool = NewBufferPool()

	encoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
	decoder, _ = zstd.NewReader(nil)
)

type EncodeType byte

const (
	EncodeRaw = iota + 1
	EncodeCompressed
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
	encode     EncodeType
	size       uint16
	data       []byte
	prev, next *ListPack
}

func NewListPack() *ListPack {
	return &ListPack{
		encode: EncodeRaw,
		data:   bpool.Get(maxListPackSize)[:0],
	}
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

func (lp *ListPack) RPop() (res string, ok bool) {
	lp.iterBack(0, 1, func(data []byte, entryStartPos, _ int) bool {
		res, ok = string(data), true
		lp.data = lp.data[:entryStartPos]
		lp.size--
		return true
	})
	return
}

func (lp *ListPack) LPop() (res string, ok bool) {
	lp.iterFront(0, 1, func(data []byte, _, entryEndPos int) bool {
		res, ok = string(data), true
		lp.data = append(lp.data[:0], lp.data[entryEndPos:]...)
		lp.size--
		return true
	})
	return
}

func (lp *ListPack) Size() int {
	return int(lp.size)
}

type lpIterator func(data []byte, entryStartPos, entryEndPos int) (stop bool)

func (lp *ListPack) iterFront(start, end int, f lpIterator) {
	if end == -1 {
		end = lp.Size()
	}
	var index int
	for i := 0; i < end; i++ {
		/*
			  index     dataStartPos    dataEndPos            indexNext
				|            |              |                     |
				+------------+--------------+---------------------+-----+
			--> |  data_len  |     data     |      entry_len      | ... |
				+------------+--------------+---------------------+-----+
				|<--- n ---->|<- data_len ->|<-- size_entry_len ->|
		*/
		dataLen, n := binary.Uvarint(lp.data[index:])
		dataStartPos := index + n
		dataEndPos := dataStartPos + int(dataLen)
		data := lp.data[dataStartPos:dataEndPos]
		indexNext := dataEndPos + SizeUvarint(dataLen+uint64(n))

		if i >= start && f(data, index, indexNext) {
			return
		}
		index = indexNext
	}
}

func (lp *ListPack) iterBack(start, end int, f lpIterator) {
	if end == -1 {
		end = lp.Size()
	}
	var index = len(lp.data)
	for i := 0; i < end; i++ {
		/*
			  indexNext  dataStartPos    dataEndPos               index
				  |            |              |                     |
			+-----+------------+--------------+---------------------+
			| ... |  data_len  |     data     |      entry_len      | <--
			+-----+------------+--------------+---------------------+
				  |<--- n ---->|<- data_len ->|<-- size_entry_len ->|
				  |<------ entry_len -------->|
		*/
		entryLen, sizeEntryLen := uvarintReverse(lp.data[:index])
		indexNext := index - int(entryLen) - sizeEntryLen
		dataLen, n := binary.Uvarint(lp.data[indexNext:])
		dataStartPos := indexNext + n
		dataEndPos := dataStartPos + int(dataLen)
		data := lp.data[dataStartPos:dataEndPos]

		if i >= start && f(data, indexNext, index) {
			return
		}
		index = indexNext
	}
}

func (lp *ListPack) Encode(encodeType EncodeType) error {
	if lp.encode == encodeType {
		return nil
	}
	lp.encode = encodeType

	switch encodeType {
	case EncodeCompressed:
		alloc := bpool.Get(len(lp.data))
		alloc = encoder.EncodeAll(lp.data, alloc)
		bpool.Put(lp.data)
		lp.data = alloc

	case EncodeRaw:
		alloc := bpool.Get(maxListPackSize)
		alloc, err := decoder.DecodeAll(lp.data, alloc)
		if err != nil {
			return err
		}
		bpool.Put(lp.data)
		lp.data = alloc
	}
	return nil
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
