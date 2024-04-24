package quicklist

import (
	"fmt"
	"testing"
)

func genListPack(start, end int) *ListPack {
	lp := NewListPack()
	for i := start; i < end; i++ {
		lp.Insert(-1, genKey(i))
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
			lp.Insert(-1, genKey(i))
		}
		for i := 0; i < N; i++ {
			val, ok := lp.Remove(0)
			equal(t, val, genKey(i))
			equal(t, true, ok)
		}
	})

	t.Run("lpush/rpop", func(t *testing.T) {
		lp := NewListPack()
		for i := 0; i < N; i++ {
			equal(t, lp.Size(), i)
			lp.Insert(0, genKey(i))
		}
		for i := 0; i < N; i++ {
			val, ok := lp.Remove(-1)
			equal(t, val, genKey(i))
			equal(t, true, ok)
		}
	})

	t.Run("iterFront", func(t *testing.T) {
		lp := genListPack(0, N)

		// iter [0, -1]
		var i int
		lp.iterFront(0, -1, func(data []byte, _, _, _ int) bool {
			equal(t, string(data), genKey(i))
			i++
			return false
		})
		equal(t, i, N)

		// iter [0, N/2]
		i = 0
		lp.iterFront(0, N/2, func(data []byte, _, _, _ int) bool {
			equal(t, string(data), genKey(i))
			i++
			return false
		})
		equal(t, i, N/2)

		// iter [N/2, -1]
		i = 0
		lp.iterFront(N/2, -1, func(data []byte, _, _, _ int) bool {
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
		lp.iterBack(0, -1, func(data []byte, _, _, _ int) bool {
			equal(t, string(data), genKey(N-1-i))
			i++
			return false
		})
		equal(t, i, N)

		// iter [0, N/2]
		i = 0
		lp.iterBack(0, N/2, func(data []byte, _, _, _ int) bool {
			equal(t, string(data), genKey(N-1-i))
			i++
			return false
		})
		equal(t, i, N/2)

		// iter [N/2, -1]
		i = 0
		lp.iterBack(N/2, -1, func(data []byte, _, _, _ int) bool {
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

		res, ok := lp.Remove(0)
		equal(t, res, genKey(1))
		equal(t, true, ok)
	})

	t.Run("removeFirst", func(t *testing.T) {
		lp := genListPack(0, N)

		index, ok := lp.RemoveFirst(genKey(N))
		if index != 0 {
			t.Error(index, N)
		}
		if ok {
			t.Error(ok)
		}

		index, ok = lp.RemoveFirst(genKey(0))
		equal(t, index, 0)
		equal(t, true, ok)

		res, ok := lp.Remove(0)
		equal(t, res, genKey(1))
		equal(t, true, ok)
	})

	t.Run("set", func(t *testing.T) {
		lp := genListPack(0, N)

		for i := 0; i < N; i++ {
			newKey := fmt.Sprintf("newkey-%d", i)
			ok := lp.Set(i, newKey)
			if !ok {
				t.Error(ok)
			}

			index, ok := lp.First(newKey)
			if index != i {
				t.Error(index, i)
			}
			if !ok {
				t.Error(ok)
			}
		}

		// set -1
		ok := lp.Set(-1, "last")
		if !ok {
			t.Error(ok)
		}

		index, ok := lp.First("last")
		if index != N-1 {
			t.Error(index, N-1)
		}
		if !ok {
			t.Error(ok)
		}

		// out of bound
		ok = lp.Set(N+1, "newKey")
		equal(t, false, ok)
	})

	t.Run("insert", func(t *testing.T) {
		lp := genListPack(0, 10)

		// insert to 0
		lp.Insert(0, "test0")
		val, ok := lp.Remove(0)
		equal(t, val, "test0")
		equal(t, ok, true)

		// insert to 5
		lp.Insert(5, "test1")
		val, ok = lp.Remove(5)
		equal(t, val, "test1")
		equal(t, ok, true)

		// insert 11
		lp.Insert(11, "test2")
		val, ok = lp.Remove(-1)
		notEqual(t, val, "test2")
		equal(t, ok, true)
	})
}
