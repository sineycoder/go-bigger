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
	a := bigger.NewBigIntegerString("5.112")
	b := bigger.NewBigIntegerString("-2.12")
	res := a.Add(b)
	fmt.Println(res.String())
}
