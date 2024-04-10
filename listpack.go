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

	Using this structure, it is fast to iterate data from both sides.
*/
type ListPack struct {
	encode EncodeType
	size   uint16
	data   []byte
	// prev, next *ListPack
}

func NewListPack() *ListPack {
	return &ListPack{
		encode: EncodeRaw,
		data:   bpool.Get(maxListPackSize)[:0],
	}
}

func (lp *ListPack) RPush(data string) {
	entry := encodeEntry(data)
	lp.data = append(lp.data, entry...)
	lp.size++
	// reuse
	bpool.Put(entry)
}

func (lp *ListPack) LPush(data string) {
	entry := encodeEntry(data)
	lp.data = slices.Insert(lp.data, 0, entry...)
	lp.size++
	// reuse
	bpool.Put(entry)
}

func (lp *ListPack) RPop() (string, bool) {
	if lp.size == 0 {
		return "", false
	}
	entryLen, n1 := uvarintReverse(lp.data)
	index := len(lp.data) - int(entryLen) - n1
	dataLen, n2 := binary.Uvarint(lp.data[index:])
	data := lp.data[index+n2 : index+int(dataLen)+n2]
	lp.data = lp.data[:index]
	lp.size--
	return string(data), true
}

func (lp *ListPack) LPop() (string, bool) {
	if lp.size == 0 {
		return "", false
	}
	dataLen, n := binary.Uvarint(lp.data)
	tail := n + int(dataLen)
	data := lp.data[n:tail]
	lp.data = lp.data[tail+varintLength(tail):]
	lp.size--
	return string(data), true
}

func (lp *ListPack) Size() int {
	return int(lp.size)
}

// encode data to [data_len, data, entry_len].
func encodeEntry(data string) []byte {
	n := len(data)
	want := varintLength(n)*2 + 1 + n
	b := bpool.Get(want)[:0]
	b = appendUvarint(b, len(data), false)
	b = append(b, data...)
	return appendUvarint(b, len(b), true)
}
