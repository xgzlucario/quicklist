package quicklist

import (
	"testing"
)

func BenchmarkListPack(b *testing.B) {
	b.Run("lpop", func(b *testing.B) {
		lp := genListPack(0, b.N)
		for i := 0; i < b.N; i++ {
			lp.LPop()
		}
	})
	b.Run("rpop", func(b *testing.B) {
		lp := genListPack(0, b.N)
		for i := 0; i < b.N; i++ {
			lp.RPop()
		}
	})
}
