package main

import (
	"fmt"
	"github.com/SineyCoder/go_big_integer/big_integer"
)

/**
 @author: nizhenxian
 @date: 2021/8/12 18:10:43
**/
func main() {
	a := big_integer.NewBigIntegerString("9867816478612964983216")
	b := big_integer.NewBigIntegerString("1231231231231231231231")
	res := a.Add(b)
	fmt.Println(res.String())
}
