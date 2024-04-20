package quicklist

import (
	"fmt"
	"testing"
)

func genListPack(start, end int) *ListPack {
	lp := NewListPack()
	for i := start; i < end; i++ {
		lp.RPush(genKey(i))
	}
	return lp
}

func genKey(i int) string {
	return fmt.Sprintf("%08x", i)
}

func TestListPack(t *testing.T) {
	const N = 1000

	t.Run("rpush/lpop", func(t *testing.T) {
		lp := NewListPack()
		for i := 0; i < N; i++ {
			equal(t, lp.Size(), i)
			lp.RPush(genKey(i))
		}
		for i := 0; i < N; i++ {
			val, ok := lp.LPop()
			equal(t, val, genKey(i))
			equal(t, true, ok)
		}
	})

	t.Run("lpush/rpop", func(t *testing.T) {
		lp := NewListPack()
		for i := 0; i < N; i++ {
			equal(t, lp.Size(), i)
			lp.LPush(genKey(i))
		}
		for i := 0; i < N; i++ {
			val, ok := lp.RPop()
			equal(t, val, genKey(i))
			equal(t, true, ok)
		}
	})

	t.Run("iterFront", func(t *testing.T) {
		lp := genListPack(0, N)

		// iter [0, -1]
		var i int
		lp.iterFront(0, -1, func(data []byte, _, _ int) bool {
			equal(t, string(data), genKey(i))
			i++
			return false
		})
		equal(t, i, N)

		// iter [0, N/2]
		i = 0
		lp.iterFront(0, N/2, func(data []byte, _, _ int) bool {
			equal(t, string(data), genKey(i))
			i++
			return false
		})
		equal(t, i, N/2)

		// iter [N/2, -1]
		i = 0
		lp.iterFront(N/2, -1, func(data []byte, _, _ int) bool {
			equal(t, string(data), genKey(i+N/2))
			i++
			return false
		})
		equal(t, i, N/2)
	})

	t.Run("iterBack", func(t *testing.T) {
		lp := genListPack(0, N)

		// iter [0, -1]
		var i int
		lp.iterBack(0, -1, func(data []byte, _, _ int) bool {
			equal(t, string(data), genKey(N-1-i))
			i++
			return false
		})
		equal(t, i, N)

		// iter [0, N/2]
		i = 0
		lp.iterBack(0, N/2, func(data []byte, _, _ int) bool {
			equal(t, string(data), genKey(N-1-i))
			i++
			return false
		})
		equal(t, i, N/2)

		// iter [N/2, -1]
		i = 0
		lp.iterBack(N/2, -1, func(data []byte, _, _ int) bool {
			equal(t, string(data), genKey(N/2-i-1))
			i++
			return false
		})
		equal(t, i, N/2)
	})

	t.Run("remove", func(t *testing.T) {
		lp := genListPack(0, N)

		val, ok := lp.Remove(N)
		equal(t, val, "")
		equal(t, false, ok)

		val, ok = lp.Remove(0)
		equal(t, val, genKey(0))
		equal(t, true, ok)

		res, ok := lp.LPop()
		equal(t, res, genKey(1))
		equal(t, true, ok)
	})

	t.Run("removeElem", func(t *testing.T) {
		lp := genListPack(0, N)

		ok := lp.RemoveElem(genKey(N))
		equal(t, false, ok)

		ok = lp.RemoveElem(genKey(0))
		equal(t, true, ok)

		res, ok := lp.LPop()
		equal(t, res, genKey(1))
		equal(t, true, ok)
	})

	t.Run("removeRange", func(t *testing.T) {
		lp := genListPack(0, N)

		// case1
		n := lp.RemoveRange(0, N/2)
		equal(t, n, N/2)

		val, ok := lp.LPop()
		equal(t, val, genKey(N/2))
		equal(t, true, ok)

		// case2
		lp = genListPack(0, N)
		n = lp.RemoveRange(N/2, N*2)
		equal(t, n, N/2)

		val, ok = lp.RPop()
		equal(t, val, genKey(N/2-1))
		equal(t, true, ok)
	})

	t.Run("set", func(t *testing.T) {
		lp := genListPack(0, N)
		for i := 0; i < N; i++ {
			ok := lp.Set(i, fmt.Sprintf("newkey-%d", i))
			equal(t, true, ok)
		}

		var i int
		lp.iterFront(0, -1, func(data []byte, _, _ int) bool {
			equal(t, string(data), fmt.Sprintf("newkey-%d", i))
			i++
			return false
		})
		equal(t, i, N)

		ok := lp.Set(N+1, "newKey")
		equal(t, false, ok)
	})
}
