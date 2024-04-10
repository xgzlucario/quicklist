package main

import (
	"fmt"
	"quicklist"
)

func main() {
	lp := quicklist.NewListPack()

	for i := 0; i < 100; i++ {
		lp.RPush(fmt.Sprintf("%08d", i))
	}
}
