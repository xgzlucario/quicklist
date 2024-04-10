package quicklist

import (
	"encoding/binary"
	"math/bits"
	"slices"
)

func appendUvarint(b []byte, n int, reverse bool) []byte {
	if !reverse {
		return binary.AppendUvarint(b, uint64(n))
	}
	bb := binary.AppendUvarint(bpool.Get(binary.MaxVarintLen32)[:0], uint64(n))
	if len(bb) > 1 {
		slices.Reverse(bb)
	}
	b = append(b, bb...)
	bpool.Put(bb)
	return b
}

// uvarintReverse is the reverse version from binary.Uvarint.
func uvarintReverse(buf []byte) (uint64, int) {
	var x uint64
	var s uint
	for i := range buf {
		b := buf[len(buf)-1-i]
		if i == binary.MaxVarintLen64 {
			return 0, -(i + 1) // overflow
		}
		if b < 0x80 {
			if i == binary.MaxVarintLen64-1 && b > 1 {
				return 0, -(i + 1) // overflow
			}
			return x | uint64(b)<<s, i + 1
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, 0
}

// SizeUvarint
// See https://go-review.googlesource.com/c/go/+/572196/1/src/encoding/binary/varint.go#174
func SizeUvarint(x uint64) int {
	return int(9*uint32(bits.Len64(x))+64) / 64
}

// SizeVarint
func SizeVarint(x int64) int {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return SizeUvarint(ux)
}
