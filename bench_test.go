package quicklist

import (
	"fmt"
	"testing"
)

func getListPack(n int) *ListPack {
	lp := NewListPack()
	for i := 0; i < n; i++ {
		lp.RPush(fmt.Sprintf("%08x", i))
	}
	return lp
}

func BenchmarkListPack(b *testing.B) {
	b.Run("lpop1", func(b *testing.B) {
		lp := getListPack(b.N)
		for i := 0; i < b.N; i++ {
			lp.LPop()
		}
	})
}
