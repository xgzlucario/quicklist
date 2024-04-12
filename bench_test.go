package quicklist

import (
	"testing"
)

func BenchmarkList(b *testing.B) {
	b.Run("lpush", func(b *testing.B) {
		lp := NewList()
		for i := 0; i < b.N; i++ {
			lp.LPush(genKey(i))
		}
	})
	b.Run("rpush", func(b *testing.B) {
		lp := NewList()
		for i := 0; i < b.N; i++ {
			lp.RPush(genKey(i))
		}
	})
	b.Run("lpop", func(b *testing.B) {
		lp := genList(0, b.N)
		for i := 0; i < b.N; i++ {
			lp.LPop()
		}
	})
	b.Run("rpop", func(b *testing.B) {
		lp := genList(0, b.N)
		for i := 0; i < b.N; i++ {
			lp.RPop()
		}
	})
	b.Run("index", func(b *testing.B) {
		lp := genList(0, b.N)
		for i := 0; i < b.N; i++ {
			lp.Index(i)
		}
	})
}
