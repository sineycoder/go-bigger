package types

import (
	"fmt"
	"math"
)

/**
 @author: nizhenxian
 @date: 2021/8/11 19:18:00
**/
type Int int32
type Long int64
type Float float32
type Double float64

const (
	POSITIVE_INFINITY = Double(math.MaxFloat64)
	NEGATIVE_INFINITY = -POSITIVE_INFINITY
)

// e.g. right shift, append 0 to high bit

func DoubleFromBits(bits Long) Double {
	return Double(math.Float64frombits(uint64(bits)))
}

func (i Int) ShiftR(val Int) Int {
	return Int(uint32(i) >> val)
}

func (i Int) ToLong() Long {
	return Long(i)
}

func (i Int) String() string {
	return fmt.Sprintf("%d", i)
}

func (l Long) ShiftR(val Int) Long {
	return Long(uint64(l) >> val)
}

func (l Long) ToInt() Int {
	return Int(l)
}

func (l Long) ToDouble() Double {
	return Double(l)
}

func (l Long) String() string {
	return fmt.Sprintf("%d", l)
}

func (l Long) Abs() Long {
	return Long(math.Abs(float64(l)))
}

func (l Long) Signum() Int {
	return ((l >> 63) | (-l).ShiftR(63)).ToInt()
}
