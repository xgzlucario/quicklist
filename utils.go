package quicklist

import (
	"encoding/binary"
	"math/bits"
	"slices"
	"unsafe"
)

func appendUvarint(b []byte, n int, reverse bool) []byte {
	if !reverse {
		return binary.AppendUvarint(b, uint64(n))
	}
	before := len(b)
	b = binary.AppendUvarint(b, uint64(n))
	after := len(b)
	if after-before > 1 {
		slices.Reverse(b[before:])
	}
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

func s2b(str *string) []byte {
	strHeader := (*[2]uintptr)(unsafe.Pointer(str))
	byteSliceHeader := [3]uintptr{
		strHeader[0], strHeader[1], strHeader[1],
	}
	return *(*[]byte)(unsafe.Pointer(&byteSliceHeader))
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
