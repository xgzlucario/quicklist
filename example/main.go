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
