package quicklist

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func genList(start, end int) *QuickList {
	lp := New()
	for i := start; i < end; i++ {
		lp.RPush(genKey(i))
	}
	return lp
}

func TestList(t *testing.T) {
	assert := assert.New(t)
	const N = 1000
	SetEachNodeMaxSize(128)

	t.Run("rpush", func(t *testing.T) {
		ls := New()
		for i := 0; i < N; i++ {
			assert.Equal(ls.Size(), i)
			ls.RPush(genKey(i))
		}
		for i := 0; i < N; i++ {
			v, ok := ls.Index(i)
			assert.Equal(genKey(i), v)
			assert.True(ok)
		}
		// check each node length
		for cur := ls.head; cur != nil; cur = cur.next {
			assert.LessOrEqual(len(cur.data), maxListPackSize)
		}
	})

	t.Run("lpush", func(t *testing.T) {
		ls := New()
		for i := 0; i < N; i++ {
			assert.Equal(ls.Size(), i)
			ls.LPush(genKey(i))
		}
		for i := 0; i < N; i++ {
			v, ok := ls.Index(N - 1 - i)
			assert.Equal(genKey(i), v)
			assert.True(ok)
		}
		// check each node length
		for cur := ls.head; cur != nil; cur = cur.next {
			assert.LessOrEqual(len(cur.data), maxListPackSize)
		}
	})

	t.Run("lpop", func(t *testing.T) {
		ls := genList(0, N)
		for i := 0; i < N; i++ {
			assert.Equal(ls.Size(), N-i)
			key, ok := ls.LPop()
			assert.Equal(key, genKey(i))
			assert.True(ok)
		}
		// pop empty list
		for i := 0; i < N; i++ {
			key, ok := ls.LPop()
			assert.Equal(key, "")
			assert.False(ok)
		}
	})

	t.Run("rpop", func(t *testing.T) {
		ls := genList(0, N)
		for i := 0; i < N; i++ {
			assert.Equal(ls.Size(), N-i)
			key, ok := ls.RPop()
			assert.Equal(key, genKey(N-i-1))
			assert.True(ok)
		}
		// pop empty list
		for i := 0; i < N; i++ {
			key, ok := ls.RPop()
			assert.Equal(key, "")
			assert.False(ok)
		}
	})

	t.Run("len", func(t *testing.T) {
		ls := New()
		for i := 0; i < N; i++ {
			ls.RPush(genKey(i))
			assert.Equal(ls.Size(), i+1)
		}
	})

	t.Run("set", func(t *testing.T) {
		ls := genList(0, N)
		for i := 0; i < N; i++ {
			newK := fmt.Sprintf("newkk-%x", i)
			ok := ls.Set(i, newK)
			assert.True(ok, i)
		}
		var count int
		ls.Range(0, -1, func(b []byte) bool {
			targetK := fmt.Sprintf("newkk-%x", count)
			assert.Equal(string(b), targetK)
			count++
			return false
		})
		assert.Equal(N, count)

		ok := ls.Set(N+1, "new")
		assert.False(ok)
	})

	t.Run("marshal", func(t *testing.T) {
		ls := genList(0, N)
		data, err := ls.MarshalJSON()
		assert.Nil(err)

		ls2 := New()
		err = ls2.UnmarshalJSON(data)
		assert.Nil(err)

		for i := 0; i < N; i++ {
			v, ok := ls.Index(i)
			assert.Equal(genKey(i), v)
			assert.True(ok)
		}
	})

	t.Run("range", func(t *testing.T) {
		ls := New()
		ls.Range(1, 2, func(s []byte) bool {
			panic("should not call")
		})
		ls = genList(0, N)

		var count int
		ls.Range(0, -1, func(s []byte) bool {
			assert.Equal(string(s), genKey(count))
			count++
			return false
		})
		assert.Equal(count, N)

		ls.Range(1, 1, func(s []byte) bool {
			panic("should not call")
		})
		ls.Range(-1, -1, func(s []byte) bool {
			panic("should not call")
		})
	})

	t.Run("revrange", func(t *testing.T) {
		ls := New()
		ls.RevRange(1, 2, func(s []byte) bool {
			panic("should not call")
		})
		ls = genList(0, N)

		var count int
		ls.RevRange(0, -1, func(s []byte) bool {
			assert.Equal(string(s), genKey(N-count-1))
			count++
			return false
		})
		assert.Equal(count, N)

		ls.RevRange(1, 1, func(s []byte) bool {
			panic("should not call")
		})
		ls.RevRange(-1, -1, func(s []byte) bool {
			panic("should not call")
		})
	})
}

func FuzzList(f *testing.F) {
	ls := New()
	vls := make([]string, 0, 4096)

	f.Fuzz(func(t *testing.T, key string) {
		assert := assert.New(t)

		switch rand.IntN(15) {
		// RPush
		case 0, 1, 2:
			k := strconv.Itoa(rand.Int())
			ls.RPush(k)
			vls = append(vls, k)

		// LPush
		case 3, 4, 5:
			k := strconv.Itoa(rand.Int())
			ls.LPush(k)
			vls = append([]string{k}, vls...)

		// LPop
		case 6, 7:
			val, ok := ls.LPop()
			if len(vls) > 0 {
				valVls := vls[0]
				vls = vls[1:]
				assert.Equal(val, valVls)
				assert.True(ok)
			} else {
				assert.Equal(val, "")
				assert.False(ok)
			}

		// RPop
		case 8, 9:
			val, ok := ls.RPop()
			if len(vls) > 0 {
				valVls := vls[len(vls)-1]
				vls = vls[:len(vls)-1]
				assert.Equal(val, valVls)
				assert.True(ok)
			} else {
				assert.Equal(val, "")
				assert.False(ok)
			}

		// Set
		case 10:
			if len(vls) > 0 {
				index := rand.IntN(len(vls))
				randKey := fmt.Sprintf("%d", rand.Uint32())
				ok := ls.Set(index, randKey)
				assert.True(ok)
				vls[index] = randKey
			}

		// Index
		case 11:
			if len(vls) > 0 {
				index := rand.IntN(len(vls))
				val, ok := ls.Index(index)
				vlsVal := vls[index]
				assert.Equal(val, vlsVal)
				assert.True(ok)
			}

		// Delete
		case 12:
			// if len(vls) > 0 {
			// index := rand.IntN(len(vls))
			// val, ok := ls.Delete(index)
			// assert.Equal(val, vls[index])
			// assert.True(ok)
			// vls = append(vls[:index], vls[index+1:]...)
			// }

		// Range
		case 13:
			if len(vls) > 2 {
				start := rand.IntN(len(vls) / 2)
				end := len(vls)/2 + rand.IntN(len(vls)/2)

				keys := make([]string, 0, end-start)
				ls.Range(start, end, func(s []byte) bool {
					keys = append(keys, string(s))
					return false
				})
				assert.Equal(keys, vls[start:end], fmt.Sprintf("start: %v, end: %v", start, end))
			}

		// Marshal
		case 14:
			// data := ls.Marshal()
			// nls := New()
			// err := nls.Unmarshal(data)
			// assert.Nil(err)

			// assert.Equal(len(vls), nls.Size())
			// if len(vls) == 0 {
			// 	assert.Equal(vls, []string{})
			// } else {
			// 	assert.Equal(vls, nls.Keys())
			// }
		}
	})
}
