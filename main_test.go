package main

import (
	"github.com/sineycoder/go-bigger/bigger"
	"github.com/sineycoder/go-bigger/types"
	"math/big"
	"testing"
)

/**
 @author: nizhenxian
 @date: 2021/8/18 11:29:54
**/

// testing Integer, bigger.bigInteger vs bigInt
func BenchmarkBiggerIntegerAdd(bb *testing.B) {
	a := bigger.BigIntegerValueOf(types.Long(534151451245))
	b := bigger.BigIntegerValueOf(types.Long(18979412))
	a.Add(b)
}
func BenchmarkBigintIntegerAdd(bb *testing.B) {
	a := big.NewInt(534151451245)
	b := big.NewInt(18979412)
	a.Add(a, b)
}

func BenchmarkBiggerIntegerSubtract(bb *testing.B) {
	a := bigger.BigIntegerValueOf(types.Long(534151451245))
	b := bigger.BigIntegerValueOf(types.Long(18979412))
	a.Subtract(b)
}
func BenchmarkBigintIntegerSubtract(bb *testing.B) {
	a := big.NewInt(534151451245)
	b := big.NewInt(18979412)
	a.Sub(a, b)
}

func BenchmarkBiggerIntegerMultiply(bb *testing.B) {
	a := bigger.BigIntegerValueOf(types.Long(534151451245))
	b := bigger.BigIntegerValueOf(types.Long(18979412))
	a.Multiply(b)
}
func BenchmarkBigintIntegerMultiply(bb *testing.B) {
	a := big.NewInt(534151451245)
	b := big.NewInt(18979412)
	a.Mul(a, b)
}

func BenchmarkBiggerIntegerDivide(bb *testing.B) {
	a := bigger.BigIntegerValueOf(types.Long(534151451245))
	b := bigger.BigIntegerValueOf(types.Long(18979412))
	a.Divide(b)
}
func BenchmarkBigintIntegerDivide(bb *testing.B) {
	a := big.NewInt(534151451245)
	b := big.NewInt(18979412)
	a.Div(a, b)
}

// testing Float, bigger.bigDecimal vs bigFloat
func BenchmarkBiggerFloatAdd(bb *testing.B) {
	a := bigger.NewBigDecimalString("534151451245")
	b := bigger.NewBigDecimalString("18979412")
	a.Add(b)
}
func BenchmarkBigintFloatAdd(bb *testing.B) {
	a := big.NewFloat(534151451245)
	b := big.NewFloat(18979412)
	a.Add(a, b)
}

func BenchmarkBiggerFloatSubtract(bb *testing.B) {
	a := bigger.NewBigDecimalString("534151451245")
	b := bigger.NewBigDecimalString("18979412")
	a.Subtract(b)
}
func BenchmarkBigintFloatSubtract(bb *testing.B) {
	a := big.NewFloat(534151451245)
	b := big.NewFloat(18979412)
	a.Sub(a, b)
}

func BenchmarkBiggerFloatMultiply(bb *testing.B) {
	a := bigger.NewBigDecimalString("534151451245")
	b := bigger.NewBigDecimalString("18979412")
	a.Multiply(b)
}
func BenchmarkBigintFloatMultiply(bb *testing.B) {
	a := big.NewFloat(534151451245)
	b := big.NewFloat(18979412)
	a.Mul(a, b)
}

func BenchmarkBiggerFloatDivide(bb *testing.B) {
	a := bigger.NewBigDecimalString("1")
	b := bigger.NewBigDecimalString("3")
	a.Divide(b, 5, bigger.ROUND_HALF_UP)
}
func BenchmarkBigintFloatDivide(bb *testing.B) {
	a := big.NewFloat(0.5)
	b := big.NewFloat(0.4)
	a.Quo(a, b)
}
