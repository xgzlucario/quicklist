package quicklist

import (
	"testing"
)

func BenchmarkList(b *testing.B) {
	b.Run("lpush", func(b *testing.B) {
		ls := New()
		for i := 0; i < b.N; i++ {
			ls.LPush(genKey(i))
		}
	})
	b.Run("rpush", func(b *testing.B) {
		ls := New()
		for i := 0; i < b.N; i++ {
			ls.RPush(genKey(i))
		}
	})
	b.Run("lpop", func(b *testing.B) {
		ls := genList(0, b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ls.LPop()
		}
	})
	b.Run("rpop", func(b *testing.B) {
		ls := genList(0, b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ls.RPop()
		}
	})
	b.Run("index", func(b *testing.B) {
		ls := genList(0, 10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ls.Index(i % 10000)
		}
	})
	b.Run("range", func(b *testing.B) {
		ls := genList(0, 10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ls.Range(0, -1, func(s []byte) (stop bool) {
				return false
			})
		}
	})
	b.Run("revrange", func(b *testing.B) {
		ls := genList(0, 10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ls.RevRange(0, -1, func(s []byte) (stop bool) {
				return false
			})
		}
	})
}
