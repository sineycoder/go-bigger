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
	a := bigger.NewBigIntegerString("1")
	b := bigger.NewBigIntegerString("2")
	fmt.Println(a.And(b))
}
