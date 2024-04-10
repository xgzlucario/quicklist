package quicklist

import (
	"encoding/binary"
	"math"
	"slices"
)

func appendUvarint(b []byte, n int, reverse bool) []byte {
	if !reverse {
		return binary.AppendUvarint(b, uint64(n))
	}
	bb := binary.AppendUvarint(bpool.Get(binary.MaxVarintLen32)[:0], uint64(n))
	slices.Reverse(bb)
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

func varintLength[T int](x T) (n int) {
	if x > math.MaxUint32 {
		panic("overflow")
	}
	if x < (1 << 7) {
		return 1
	} else if x < (1 << 14) {
		return 2
	} else if x < (1 << 21) {
		return 3
	} else if x < (1 << 28) {
		return 4
	}
	return 5
}
