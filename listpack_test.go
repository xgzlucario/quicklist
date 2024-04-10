package quicklist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListPack(t *testing.T) {
	assert := assert.New(t)
	const N = 1000

	t.Run("rpush/lpop", func(t *testing.T) {
		lp := NewListPack()
		for i := 0; i < N; i++ {
			assert.Equal(lp.Size(), i)
			lp.RPush(fmt.Sprintf("%08d", i))
		}
		for i := 0; i < N; i++ {
			val, ok := lp.LPop()
			assert.Equal(val, fmt.Sprintf("%08d", i))
			assert.True(ok)
		}
	})

	t.Run("lpush/rpop", func(t *testing.T) {
		lp := NewListPack()
		for i := 0; i < N; i++ {
			assert.Equal(lp.Size(), i)
			lp.LPush(fmt.Sprintf("%08d", i))
		}
		for i := 0; i < N; i++ {
			val, ok := lp.RPop()
			assert.Equal(val, fmt.Sprintf("%08d", i))
			assert.True(ok)
		}
	})
}
