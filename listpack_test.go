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
	return fmt.Sprintf("%08d", i)
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
}
