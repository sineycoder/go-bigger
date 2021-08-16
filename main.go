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
	a := bigger.NewBigDecimalString("0.7")
	b := bigger.NewBigDecimalString("0.2")
	res := a.Add(b)
	res = res.SetScale(30, bigger.ROUND_HALF_UP)
	fmt.Println(res.String())
}
