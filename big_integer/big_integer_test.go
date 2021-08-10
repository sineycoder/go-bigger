package big_integer

import "testing"

/**
 @author: nizhenxian
 @date: 2021/8/10 16:59:02
**/

func TestValueOf(t *testing.T) {
	a := ValueOf(1231212321312211233)
	b := ValueOf(123123)
	res := a.Multiply(b)
	println(res.String())
}
