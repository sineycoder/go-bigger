package big_integer

import (
	"fmt"
	"github.com/SineyCoder/go_bigger/types"
	"math/big"
	"testing"
)

/**
 @author: nizhenxian
 @date: 2021/8/10 16:59:02
**/

func TestValueOf(t *testing.T) {
	a := ValueOf(types.Long(534151451245))
	b := ValueOf(types.Long(18979412))
	res := a.Add(b)
	fmt.Printf("%+v", res)
}

func TestArraycopy(t *testing.T) {
	a := big.NewFloat(0.3)
	a.SetPrec(100)
	b := big.NewFloat(0.4)
	b.SetPrec(100)
	a = a.Add(a, b)
	fmt.Println(a)
}

func BenchmarkValueOf(bb *testing.B) {
	a := ValueOf(types.Long(97917234971231119))
	b := ValueOf(types.Long(-9791723497123222))
	a.Subtract(b)
}

func BenchmarkArraycopy(bb *testing.B) {
	a := big.NewInt(21)
	b := big.NewInt(3)
	a.Div(a, b)
	fmt.Println(a)
}
