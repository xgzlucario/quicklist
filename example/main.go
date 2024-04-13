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
