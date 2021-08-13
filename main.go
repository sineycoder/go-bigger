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
	a := big_integer.ValueOf(6782613786431)
	b := big_integer.ValueOf(-678261378231)
	res := a.Multiply(b)
	fmt.Println(res.String())
}
