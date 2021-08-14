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
	a := big_integer.ValueOf(4000000000)
	res := a.SqrtAndRemainder()
	b := big_integer.ValueOf(63245)
	b = b.Multiply(b)
	b = b.Add(big_integer.ValueOf(69975))
	//b = b.Multiply(b).Add(big_integer.ValueOf(69975))
	fmt.Println(res, b)
}
