# quicklist

[![Go Report Card](https://goreportcard.com/badge/github.com/xgzlucario/quicklist)](https://goreportcard.com/report/github.com/xgzlucario/quicklist) [![Go Reference](https://pkg.go.dev/badge/github.com/xgzlucario/quicklist.svg)](https://pkg.go.dev/github.com/xgzlucario/quicklist) ![](https://img.shields.io/badge/go-1.22-orange.svg) ![](https://img.shields.io/github/languages/code-size/xgzlucario/quicklist.svg) [![codecov](https://codecov.io/gh/xgzlucario/GigaCache/graph/badge.svg?token=yC1xELYaM2)](https://codecov.io/gh/xgzlucario/quicklist) [![Test](https://github.com/xgzlucario/quicklist/actions/workflows/go.yml/badge.svg)](https://github.com/xgzlucario/quicklist/actions/workflows/go.yml)

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
	fmt.Println("Len:", ls.Size()) // 200

	// Index
	val, ok := ls.Index(50)
	fmt.Println("Index:", val, ok)

	// Set
	ok = ls.Set(0, "newValue")
	fmt.Println("Set:", ok) // true

	// LPop
	val, ok = ls.LPop()
	fmt.Println("LPop:", val, ok) // newValue, true
	// RPop
	val, ok = ls.RPop()
	fmt.Println("RPop:", val, ok) // 00099, true

	// Range
	ls.Range(0, -1, func(s []byte) (stop bool) {
		// do something
		return false
	})
	ls.RevRange(0, -1, func(s []byte) (stop bool) {
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
BenchmarkList/lpush-20           6714782             177.5 ns/op            81 B/op          4 allocs/op
BenchmarkList/rpush-20           9673924             122.9 ns/op            57 B/op          3 allocs/op
BenchmarkList/lpop-20           33424344             36.82 ns/op            12 B/op          1 allocs/op
BenchmarkList/rpop-20           34043060             35.22 ns/op            11 B/op          1 allocs/op
BenchmarkList/index-20           2519694             478.2 ns/op             8 B/op          1 allocs/op
BenchmarkList/set-20             2319103             520.9 ns/op            16 B/op          1 allocs/op
BenchmarkList/range-20             26865             44864 ns/op             0 B/op          0 allocs/op
BenchmarkList/revrange-20                  23131           52130 ns/op               0 B/op          0 allocs/op
BenchmarkListPack/set/same-len-20        1000000            1009 ns/op              16 B/op          1 allocs/op
BenchmarkListPack/set/less-len-20        1000000            1011 ns/op              16 B/op          2 allocs/op
BenchmarkListPack/set/great-len-20       1000000            1020 ns/op              24 B/op          2 allocs/op
PASS
ok      github.com/xgzlucario/quicklist 23.947s
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