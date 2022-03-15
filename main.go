package main

import (
	"fmt"

	"github.com/sineycoder/go-bigger/bigger"
)

/**
 @author: nizhenxian
 @date: 2021/8/12 18:10:43
**/
func main() {
	a := bigger.NewBigIntegerBytes([]byte{1})
	c := a.Xor(a)
	fmt.Println(a, c.String())
}
