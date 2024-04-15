package quicklist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert := assert.New(t)
	const N = 1000

	t.Run("rpush/lpop", func(t *testing.T) {
		lp := NewListPack()
		for i := 0; i < N; i++ {
			assert.Equal(lp.Size(), i)
			lp.RPush(genKey(i))
		}
		for i := 0; i < N; i++ {
			val, ok := lp.LPop()
			assert.Equal(val, genKey(i))
			assert.True(ok)
		}
	})

	t.Run("lpush/rpop", func(t *testing.T) {
		lp := NewListPack()
		for i := 0; i < N; i++ {
			assert.Equal(lp.Size(), i)
			lp.LPush(genKey(i))
		}
		for i := 0; i < N; i++ {
			val, ok := lp.RPop()
			assert.Equal(val, genKey(i))
			assert.True(ok)
		}
	})

	t.Run("iterFront", func(t *testing.T) {
		lp := genListPack(0, N)

		// iter [0, -1]
		var i int
		lp.iterFront(0, -1, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(i))
			i++
			return false
		})
		assert.Equal(i, N)

		// iter [0, N/2]
		i = 0
		lp.iterFront(0, N/2, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(i))
			i++
			return false
		})
		assert.Equal(i, N/2)

		// iter [N/2, -1]
		i = 0
		lp.iterFront(N/2, -1, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(i+N/2))
			i++
			return false
		})
		assert.Equal(i, N/2)
	})

	t.Run("iterBack", func(t *testing.T) {
		lp := genListPack(0, N)

		// iter [0, -1]
		var i int
		lp.iterBack(0, -1, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(N-1-i))
			i++
			return false
		})
		assert.Equal(i, N)

		// iter [0, N/2]
		i = 0
		lp.iterBack(0, N/2, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(N-1-i))
			i++
			return false
		})
		assert.Equal(i, N/2)

		// iter [N/2, -1]
		i = 0
		lp.iterBack(N/2, -1, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(N/2-i-1))
			i++
			return false
		})
		assert.Equal(i, N/2)
	})

	t.Run("remove", func(t *testing.T) {
		lp := genListPack(0, N)

		val, ok := lp.Remove(N)
		assert.Equal(val, "")
		assert.False(ok)

		val, ok = lp.Remove(0)
		assert.Equal(val, genKey(0))
		assert.True(ok)

		res, ok := lp.LPop()
		assert.Equal(res, genKey(1))
		assert.True(ok)
	})

	t.Run("removeElem", func(t *testing.T) {
		lp := genListPack(0, N)

		ok := lp.RemoveElem(genKey(N))
		assert.False(ok)

		ok = lp.RemoveElem(genKey(0))
		assert.True(ok)

		res, ok := lp.LPop()
		assert.Equal(res, genKey(1))
		assert.True(ok)
	})

	t.Run("removeRange", func(t *testing.T) {
		lp := genListPack(0, N)

		// case1
		n := lp.RemoveRange(0, N/2)
		assert.Equal(n, N/2)

		val, ok := lp.LPop()
		assert.Equal(val, genKey(N/2))
		assert.True(ok)

		// case2
		lp = genListPack(0, N)
		n = lp.RemoveRange(N/2, N*2)
		assert.Equal(n, N/2)

		val, ok = lp.RPop()
		assert.Equal(val, genKey(N/2-1))
		assert.True(ok)
	})

	t.Run("comressed", func(t *testing.T) {
		lp := genListPack(0, N)
		sizeBefore := len(lp.data)

		// encode same
		assert.Nil(lp.Encode(EncodeRaw))

		// compress
		err := lp.Encode(EncodeCompressed)
		assert.Nil(err)
		sizeNow := len(lp.data)

		assert.Less(sizeNow, sizeBefore)

		// decompress
		err = lp.Encode(EncodeRaw)
		assert.Nil(err)
		sizeNow = len(lp.data)

		assert.Equal(sizeNow, sizeBefore)

		// check
		var i int
		lp.iterFront(0, -1, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(i))
			i++
			return false
		})

		// decode error
		lp = genListPack(0, N)
		lp.encode = EncodeCompressed
		err = lp.Encode(EncodeRaw)
		assert.NotNil(err)
	})

	t.Run("set", func(t *testing.T) {
		lp := genListPack(0, N)
		for i := 0; i < N; i++ {
			ok := lp.Set(i, fmt.Sprintf("newkey-%d", i))
			assert.True(ok)
		}

		var i int
		lp.iterFront(0, -1, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), fmt.Sprintf("newkey-%d", i))
			i++
			return false
		})
		assert.Equal(i, N)

		ok := lp.Set(N+1, "newKey")
		assert.False(ok)
	})
}
