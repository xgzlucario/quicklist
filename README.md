# quicklist
Implement redis quicklist data structure, based on listpack rather than ziplist to optimize cascade update.

# Usage

```go
package main

import (
	"fmt"

	"github.com/xgzlucario/quicklist"
)

func genKey(i int) string {
	return fmt.Sprintf("%05d", i)
}

func main() {
	ls := quicklist.New()

	// RPush
	for i := 0; i < 100; i++ {
		ls.RPush(genKey(i))
	}

	// LPush
	for i := 0; i < 100; i++ {
		ls.RPush(genKey(i))
	}

	// Len
	// Output: 200
	fmt.Println(ls.Size())

	// Index
	val, ok := ls.Index(50)
	fmt.Println("Index:", val, ok)

	// LPop
	val, ok = ls.LPop()
	fmt.Println("LPop:", val, ok)

	// RPop
	val, ok = ls.RPop()
	fmt.Println("RPop:", val, ok)

	// Range
	ls.Range(0, -1, func(s string) (stop bool) {
		// do something
		return false
	})
}
```

# Benchmark

```
goos: linux
goarch: amd64
pkg: github.com/xgzlucario/quicklist
cpu: 13th Gen Intel(R) Core(TM) i5-13600KF
BenchmarkList/lpush-20           6917678              171.4 ns/op            81 B/op          4 allocs/op
BenchmarkList/rpush-20           9794673              121.2 ns/op            57 B/op          3 allocs/op
BenchmarkList/lpop-20           33277419              36.08 ns/op            12 B/op          1 allocs/op
BenchmarkList/rpop-20           34612028              34.37 ns/op            11 B/op          1 allocs/op
BenchmarkList/index-20           2526723              477.1 ns/op             8 B/op          1 allocs/op
BenchmarkList/range-20             26451              44666 ns/op             0 B/op          0 allocs/op
BenchmarkList/revrange-20          22366              53787 ns/op             0 B/op          0 allocs/op
PASS
ok      github.com/xgzlucario/quicklist 19.181s
```

```
slice
entries: 20000000
alloc: 614 mb
gcsys: 10 mb
heap inuse: 614 mb
heap object: 19531 k
gc: 15
pause: 448.89µs
cost: 1.974449294s

quicklist
entries: 20000000
alloc: 261 mb
gcsys: 8 mb
heap inuse: 262 mb
heap object: 3589 k
gc: 27
pause: 1.492172ms
cost: 2.320623234s
```