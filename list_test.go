package quicklist

import (
	"crypto/md5"
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"
)

func genList(start, end int) *QuickList {
	lp := New()
	for i := start; i < end; i++ {
		lp.RPush(genKey(i))
	}
	return lp
}

func TestList(t *testing.T) {
	const N = 1000
	SetMaxListPackSize(128)
	SetDefaultListPackCap(128)

	t.Run("rpush", func(t *testing.T) {
		ls := New()
		for i := 0; i < N; i++ {
			equal(t, ls.Size(), i)
			ls.RPush(genKey(i))
		}
		for i := 0; i < N; i++ {
			v, ok := ls.Index(i)
			equal(t, genKey(i), v)
			equal(t, true, ok)
		}
		// check each node length
		for cur := ls.head; cur != nil; cur = cur.next {
			lessOrEqual(t, len(cur.data), maxListPackSize)
		}
	})

	t.Run("lpush", func(t *testing.T) {
		ls := New()
		for i := 0; i < N; i++ {
			equal(t, ls.Size(), i)
			ls.LPush(genKey(i))
		}
		for i := 0; i < N; i++ {
			v, ok := ls.Index(N - 1 - i)
			equal(t, genKey(i), v)
			equal(t, true, ok)
		}
		// check each node length
		for cur := ls.head; cur != nil; cur = cur.next {
			lessOrEqual(t, len(cur.data), maxListPackSize)
		}
	})

	t.Run("lpop", func(t *testing.T) {
		ls := genList(0, N)
		for i := 0; i < N; i++ {
			equal(t, ls.Size(), N-i)
			key, ok := ls.LPop()
			equal(t, key, genKey(i))
			equal(t, true, ok)
		}
		// pop empty list
		key, ok := ls.LPop()
		equal(t, key, "")
		equal(t, false, ok)
	})

	t.Run("rpop", func(t *testing.T) {
		ls := genList(0, N)
		for i := 0; i < N; i++ {
			equal(t, ls.Size(), N-i)
			key, ok := ls.RPop()
			equal(t, key, genKey(N-i-1))
			equal(t, true, ok)
		}
		// pop empty list
		key, ok := ls.RPop()
		equal(t, key, "")
		equal(t, false, ok)
	})

	t.Run("len", func(t *testing.T) {
		ls := New()
		for i := 0; i < N; i++ {
			ls.RPush(genKey(i))
			equal(t, ls.Size(), i+1)
		}
	})

	t.Run("set", func(t *testing.T) {
		ls := genList(0, N)
		for i := 0; i < N; i++ {
			newK := fmt.Sprintf("newkk-%x", i)
			ok := ls.Set(i, newK)
			equal(t, true, ok)
		}
		var count int
		ls.Range(0, -1, func(b []byte) bool {
			targetK := fmt.Sprintf("newkk-%x", count)
			equal(t, string(b), targetK)
			count++
			return false
		})
		equal(t, N, count)

		ok := ls.Set(N+1, "new")
		equal(t, false, ok)
	})

	t.Run("remove", func(t *testing.T) {
		ls := genList(0, N)
		for i := 0; i < N-1; i++ {
			val, ok := ls.Remove(0)
			equal(t, val, genKey(i))
			equal(t, true, ok)

			val, ok = ls.Index(0)
			equal(t, val, genKey(i+1))
			equal(t, true, ok)
		}

		equal(t, ls.head.Size(), 0)
		// only has 2 nodes.
		equal(t, ls.head.next, ls.tail)
		equal(t, ls.tail.Size(), 1)

		val, ok := ls.tail.Remove(-1)
		equal(t, val, genKey(N-1))
		equal(t, true, ok)
	})

	t.Run("removeFirst", func(t *testing.T) {
		ls := genList(0, N)

		// remove not exist item.
		index, ok := ls.RemoveFirst("none")
		if index != 0 {
			t.Error(index, 0)
		}
		if ok {
			t.Error(ok)
		}

		for i := 0; i < N-1; i++ {
			// same as LPop
			index, ok := ls.RemoveFirst(genKey(i))
			if index != 0 {
				t.Error(index, i)
			}
			if !ok {
				t.Error(ok)
			}

			val, ok := ls.Index(0)
			if val != genKey(i+1) {
				t.Error(val, genKey(i+1))
			}
			if !ok {
				t.Error(ok)
			}
		}

		equal(t, ls.head.Size(), 0)

		// only has 2 nodes.
		equal(t, ls.head.next, ls.tail)
		equal(t, ls.tail.Size(), 1)

		val, ok := ls.tail.Remove(-1)
		equal(t, val, genKey(N-1))
		equal(t, true, ok)
	})

	t.Run("marshal", func(t *testing.T) {
		ls := genList(0, N)
		data, err := ls.MarshalBinary()
		isNil(t, err)

		ls2 := New()
		err = ls2.UnmarshalBinary(data)
		isNil(t, err)

		for i := 0; i < N; i++ {
			v, ok := ls.Index(i)
			equal(t, genKey(i), v)
			equal(t, true, ok)
		}

		// unmarshal error
		data = md5.New().Sum(data)
		err = ls2.UnmarshalBinary(data)
		isNotNil(t, err)

		// unmarshal size error
		data = []byte{1, 1, 1, 1}
		err = ls2.UnmarshalBinary(data)
		isNotNil(t, err)
	})

	t.Run("range", func(t *testing.T) {
		ls := New()
		ls.Range(1, 2, func(s []byte) bool {
			panic("should not call")
		})
		ls = genList(0, N)

		var count int
		ls.Range(0, -1, func(s []byte) bool {
			equal(t, string(s), genKey(count))
			count++
			return false
		})
		equal(t, count, N)

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
			equal(t, string(s), genKey(N-count-1))
			count++
			return false
		})
		equal(t, count, N)

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
				equal(t, val, valVls)
				equal(t, true, ok)
			} else {
				equal(t, val, "")
				equal(t, false, ok)
			}

		// RPop
		case 8, 9:
			val, ok := ls.RPop()
			if len(vls) > 0 {
				valVls := vls[len(vls)-1]
				vls = vls[:len(vls)-1]
				equal(t, val, valVls)
				equal(t, true, ok)
			} else {
				equal(t, val, "")
				equal(t, false, ok)
			}

		// Set
		case 10:
			if len(vls) > 0 {
				index := rand.IntN(len(vls))
				randKey := fmt.Sprintf("%d", rand.Uint32())
				ok := ls.Set(index, randKey)
				equal(t, true, ok)
				vls[index] = randKey
			}

		// Index
		case 11:
			if len(vls) > 0 {
				index := rand.IntN(len(vls))
				val, ok := ls.Index(index)
				vlsVal := vls[index]
				equal(t, val, vlsVal)
				equal(t, true, ok)
			}

		// Remove
		case 12:
			if len(vls) > 0 {
				index := rand.IntN(len(vls))
				val, ok := ls.Remove(index)
				equal(t, val, vls[index])
				equal(t, true, ok)
				vls = append(vls[:index], vls[index+1:]...)
			}

		// Range
		case 13:
			if len(vls) > 0 {
				end := rand.IntN(len(vls))
				if end == 0 {
					return
				}
				start := rand.IntN(end)

				var count int
				ls.Range(start, end, func(data []byte) bool {
					equal(t, b2s(data), vls[start+count])
					count++
					return false
				})
			}

		// MarshalBinary
		case 14:
			data, _ := ls.MarshalBinary()
			nls := New()
			err := nls.UnmarshalBinary(data)
			isNil(t, err)

			var i int
			nls.Range(0, -1, func(data []byte) bool {
				equal(t, b2s(data), vls[i])
				i++
				return false
			})
		}
	})
}
