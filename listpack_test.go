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

		var count int
		lp.iterFront(0, -1, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(count))
			count++
			return false
		})
		assert.Equal(count, N)

		count = 0
		lp.iterFront(0, N/2, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(count))
			count++
			return false
		})
		assert.Equal(count, N/2)

		count = 0
		lp.iterFront(N/2, -1, func(data []byte, _, _ int) bool {
			assert.Equal(string(data), genKey(count+N/2))
			count++
			return false
		})
		assert.Equal(count, N/2)
	})
}
