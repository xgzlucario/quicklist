package quicklist

import (
	"bytes"
	"encoding/binary"
	"math"
	"slices"
	"testing"
)

func TestUtils(t *testing.T) {
	t.Run("appendVarint", func(t *testing.T) {
		for i := 0; i < math.MaxUint16; i++ {
			// append varint
			b1 := binary.AppendUvarint(nil, uint64(i))
			b2 := appendUvarint(nil, i, false)
			b3 := appendUvarint(nil, i, true)
			b4 := slices.Clone(b3)
			slices.Reverse(b4)

			equalBytes(t, b1, b2)
			equalBytes(t, b1, b2)
			equalBytes(t, b1, b4)

			// read uvarint
			x1, s1 := binary.Uvarint(b1)
			x2, s2 := uvarintReverse(b3)
			x3, s3 := uvarintReverse(append([]byte("something"), b3...))

			equal(t, x1, x2)
			equal(t, x1, x3)

			equal(t, s1, s2)
			equal(t, s1, s3)
		}
	})
}

// for test
func lessOrEqual[T int](t *testing.T, expected, actual T) {
	if expected > actual {
		t.Fatalf("[lessOrEqual] expected: %v, actual: %v", expected, actual)
	}
}

func notEqual[T comparable](t *testing.T, expected, actual T) {
	if expected == actual {
		t.Fatalf("[notEqual] expected: %v, actual: %v", expected, actual)
	}
}

func equal[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Fatalf("[equal] expected: %v, actual: %v", expected, actual)
	}
}

func equalBytes(t *testing.T, expected, actual []byte) {
	if !bytes.Equal(expected, actual) {
		t.Fatalf("[equalBytes] expected: %v, actual: %v", expected, actual)
	}
}

func isNil(t *testing.T, expected any) {
	if expected != nil {
		t.Fatalf("[isNil] expected nil: %v", expected)
	}
}

func isNotNil(t *testing.T, expected any) {
	if expected == nil {
		t.Fatalf("[isNotNil] expected not nil: %v", expected)
	}
}
