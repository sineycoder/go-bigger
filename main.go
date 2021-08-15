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
	a := bigger.NewBigDecimalString("123123.12e10")
	fmt.Println(a.String())
}
