# quicklist
Implement redis quicklist data structure, based on listpack rather than ziplist to optimize cascade update.

# Usage

```go
package main

import (
	"fmt"
	"quicklist"
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
pkg: quicklist
cpu: 13th Gen Intel(R) Core(TM) i5-13600KF
BenchmarkList/lpush-20         	 6683034	       184.5 ns/op	      74 B/op	       4 allocs/op
BenchmarkList/rpush-20         	10769210	       110.3 ns/op	      50 B/op	       3 allocs/op
BenchmarkList/lpop-20          	 7403005	       159.8 ns/op	      58 B/op	       4 allocs/op
BenchmarkList/rpop-20          	 8414138	       140.5 ns/op	      58 B/op	       4 allocs/op
BenchmarkList/indexFront-20    	  584559	      2431 ns/op	      58 B/op	       4 allocs/op
PASS
ok  	quicklist	6.862s
```