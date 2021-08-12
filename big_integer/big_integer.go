package big_integer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/SineyCoder/go_big_integer/tool"
	"github.com/SineyCoder/go_big_integer/types"
	"math"
	"strconv"
)

/**
 @author: nizhenxian
 @date: 2021/8/10 16:54:18
**/

const (
	MAX_CONSTANT                         = 16
	MAX_INT32                            = types.Int(0x7fffffff)
	MAX_INT64                            = types.Long(0x7fffffffffffffff)
	MIN_INT32                            = ^MAX_INT32
	MIN_INT64                            = ^MAX_INT64
	KARATSUBA_SQUARE_THRESHOLD           = 128
	TOOM_COOK_SQUARE_THRESHOLD           = 216
	MULTIPLY_SQUARE_THRESHOLD            = 20
	SCHOENHAGE_BASE_CONVERSION_THRESHOLD = 20
	BURNIKEL_ZIEGLER_THRESHOLD           = 80
	BURNIKEL_ZIEGLER_OFFSET              = 40
)

var (
	ZERO                = NewBigInteger([]types.Int{0}, 0)
	ONE                 = ValueOf(1)
	NEGATIVE_ONE        = ValueOf(-1)
	LOG_TWO             = types.Double(math.Log(2.0))
	LONG_MASK           = types.Long(0xffffffff)
	posConst            = make([]*BigInteger, MAX_CONSTANT+1)
	negConst            = make([]*BigInteger, MAX_CONSTANT+1)
	logCache            = make([]types.Double, 32+1)
	powerCache          = make([][]*BigInteger, 32+1)
	zeros               = "000000000000000000000000000000000000000000000000000000000000000" // the length of zeros, length=63
	lowestSetBitPlusTwo types.Int
	digitsPerLong       = []types.Int{0, 0,
		62, 39, 31, 27, 24, 22, 20, 19, 18, 18, 17, 17, 16, 16, 15, 15, 15, 14,
		14, 14, 14, 13, 13, 13, 13, 13, 13, 12, 12, 12, 12, 12, 12, 12, 12}
	longRadix = []*BigInteger{nil, nil,
		ValueOf(0x4000000000000000), ValueOf(0x383d9170b85ff80b),
		ValueOf(0x4000000000000000), ValueOf(0x6765c793fa10079d),
		ValueOf(0x41c21cb8e1000000), ValueOf(0x3642798750226111),
		ValueOf(0x1000000000000000), ValueOf(0x12bf307ae81ffd59),
		ValueOf(0xde0b6b3a7640000), ValueOf(0x4d28cb56c33fa539),
		ValueOf(0x1eca170c00000000), ValueOf(0x780c7372621bd74d),
		ValueOf(0x1e39a5057d810000), ValueOf(0x5b27ac993df97701),
		ValueOf(0x1000000000000000), ValueOf(0x27b95e997e21d9f1),
		ValueOf(0x5da0e1e53c5c8000), ValueOf(0xb16a458ef403f19),
		ValueOf(0x16bcc41e90000000), ValueOf(0x2d04b7fdd9c0ef49),
		ValueOf(0x5658597bcaa24000), ValueOf(0x6feb266931a75b7),
		ValueOf(0xc29e98000000000), ValueOf(0x14adf4b7320334b9),
		ValueOf(0x226ed36478bfa000), ValueOf(0x383d9170b85ff80b),
		ValueOf(0x5a3c23e39c000000), ValueOf(0x4e900abb53e6b71),
		ValueOf(0x7600ec618141000), ValueOf(0xaee5720ee830681),
		ValueOf(0x1000000000000000), ValueOf(0x172588ad4f5f0981),
		ValueOf(0x211e44f7d02c1000), ValueOf(0x2ee56725f06e5c71),
		ValueOf(0x41c21cb8e1000000)}
)

type BigInteger struct {
	Signum                    types.Int   // -1 for negative, 0 for zero, 1 for positive
	Mag                       []types.Int // order
	firstNonzeroIntNumPlusTwo types.Int
	bitLengthPlusOne          types.Int
}

func init() {
	// cache
	for i := types.Int(1); i <= MAX_CONSTANT; i++ {
		posConst[i] = &BigInteger{
			Signum: 1,
			Mag:    []types.Int{i},
		}
		negConst[i] = &BigInteger{
			Signum: -1,
			Mag:    []types.Int{i},
		}
	}

	for i := 2; i <= 32; i++ {
		powerCache[i] = []*BigInteger{
			ValueOf(types.Long(i)),
		}
		logCache[i] = types.Double(math.Log(float64(i)))
	}
}

func NewBigIntegerInt(val types.Long) *BigInteger {
	bigInteger := &BigInteger{}
	if val < 0 {
		val = -val
		bigInteger.Signum = -1 // set Signum as -1, value is negative
	} else {
		bigInteger.Signum = 1
	}

	highBit := (val >> 32).ToInt()
	if highBit == 0 {
		bigInteger.Mag = []types.Int{val.ToInt()} // high 32 bits is all zero
	} else {
		bigInteger.Mag = []types.Int{highBit, val.ToInt()}
	}
	return bigInteger
}

func ValueOf(val types.Long) *BigInteger {
	if val == 0 {
		return ZERO
	}
	if val > 0 && val <= MAX_CONSTANT {
		return posConst[val]
	} else if val < 0 && val >= -MAX_CONSTANT {
		return negConst[-val]
	}
	return NewBigIntegerInt(val)
}

func (b *BigInteger) Add(val *BigInteger) *BigInteger {
	if val.Signum == 0 {
		return b
	}
	if b.Signum == 0 {
		return val
	}
	if val.Signum == b.Signum {
		return NewBigInteger(add(b.Mag, val.Mag), b.Signum)
	}

	cmp := b.compareMagnitute(val)
	if cmp == 0 {
		return ZERO
	}
	var resultMag []types.Int
	if cmp > 0 {
		resultMag = subtract_(b.Mag, val.Mag)
	} else {
		resultMag = subtract_(val.Mag, b.Mag)
	}
	resultMag = trustedStripLeadingZeroInts(resultMag)
	if cmp == b.Signum {
		return NewBigInteger(resultMag, 1)
	} else {
		return NewBigInteger(resultMag, -1)
	}
}

func subtract_(big, little []types.Int) []types.Int {
	bigIndex := types.Int(len(big))
	result := make([]types.Int, bigIndex)
	littleIndex := types.Int(len(little))
	difference := types.Long(0)

	for littleIndex > 0 {
		bigIndex--
		littleIndex--
		difference = (big[bigIndex].ToLong() & LONG_MASK) -
			(little[littleIndex].ToLong() & LONG_MASK) +
			(difference >> 32)
		result[bigIndex] = difference.ToInt()
	}

	borrow := difference>>32 != 0
	for bigIndex > 0 && borrow {
		bigIndex--
		result[bigIndex] = big[bigIndex] - 1
		borrow = result[bigIndex] == -1
	}

	for bigIndex > 0 {
		bigIndex--
		result[bigIndex] = big[bigIndex]
	}

	return result
}

func (b *BigInteger) String() string {
	return b.StringRadix(10)
}

func (b *BigInteger) StringRadix(radix types.Int) string {
	if b.Signum == 0 {
		return "0"
	}
	radix = 10 // current only supports 10
	//if radix < 2 || radix > 36 {
	//	radix = 10
	//}

	if len(b.Mag) <= SCHOENHAGE_BASE_CONVERSION_THRESHOLD {
		return b.smallToString(radix)
	}

	var buf bytes.Buffer
	if b.Signum < 0 {
		toString(b.negate(), &buf, radix, 0)
		return "-" + buf.String()
	} else {
		toString(b, &buf, radix, 0)
	}
	return buf.String()
}

func NewBigInteger(magnitude []types.Int, signum types.Int) *BigInteger {
	bigInteger := &BigInteger{}
	if len(magnitude) == 0 {
		bigInteger.Signum = 0
	} else {
		bigInteger.Signum = signum
	}
	bigInteger.Mag = magnitude
	return bigInteger
}

func Arraycopy(src []types.Int, srcPos types.Int, dest []types.Int, destPos, length types.Int) {
	for i := types.Int(0); i < length; i++ {
		dest[destPos+i] = src[srcPos+i]
	}
}

func (b *BigInteger) bitLength() types.Int {
	n := b.bitLengthPlusOne - 1
	if n == -1 {
		var m []types.Int
		length := types.Int(len(m))
		if length == 0 {
			n = 0
		} else {
			magBitLength := ((length - 1) << 5) + BitLengthForInt(b.Mag[0])
			if b.Signum < 0 {
				pow2 := bitCount(b.Mag[0]) == 1
				for i := types.Int(1); i < length && pow2; i++ {
					pow2 = b.Mag[i] == 0
				}

				if pow2 {
					n = magBitLength - 1
				} else {
					n = magBitLength
				}
			} else {
				n = magBitLength
			}
		}
		b.bitLengthPlusOne = n + 1
	}
	return n
}

func bitCount(i types.Int) types.Int {
	i = i - (i.ShiftR(1) & 0x55555555)
	i = (i & 0x33333333) + (i.ShiftR(2) & 0x33333333)
	i = i + (i.ShiftR(4))&0x0f0f0f0f
	i = i + i.ShiftR(8)
	i = i + i.ShiftR(16)
	return i & 0x3f
}

func BitLengthForInt(n types.Int) types.Int {
	return 32 - NumberOfLeadingZeros(n)
}

func NumberOfLeadingZeros(i types.Int) types.Int {
	if i <= 0 {
		if i == 0 {
			return 32
		} else {
			return 0
		}
	}
	n := types.Int(31)
	if i >= 1<<16 {
		n -= 16
		i = i.ShiftR(16)
	}
	if i >= 1<<8 {
		n -= 8
		i = i.ShiftR(8)
	}
	if i >= 1<<4 {
		n -= 4
		i = i.ShiftR(4)
	}
	if i >= 1<<2 {
		n -= 2
		i = i.ShiftR(2)
	}
	return n - i.ShiftR(1)
}

/*
toString Converts the specified BigInteger to a string and appends to buf.
*/
func toString(u *BigInteger, buf *bytes.Buffer, radix types.Int, digits types.Int) {

	if len(u.Mag) <= SCHOENHAGE_BASE_CONVERSION_THRESHOLD {
		s := u.smallToString(radix)

		if (types.Int(len(s)) < digits) && buf.Len() > 0 {
			for i := types.Int(len(s)); i < digits; i++ {
				buf.WriteString("0")
			}
		}
		buf.WriteString(s)
		return
	}

	var b, n types.Int
	b = u.bitLength()

	// Calculate a value for n in the equation radix^(2^n) = u
	n = types.Int(math.Round(math.Log(float64(types.Double(b)*LOG_TWO))/float64(LOG_TWO) - 1.0))
	v := getRadixConversionCache(radix, n)
	var result []*BigInteger
	result = u.DivideAndRemainder(v)

	expectedDigits := types.Int(1 << n)

	toString(result[0], buf, radix, digits-expectedDigits)
	toString(result[1], buf, radix, expectedDigits)
}

// getRadixConversionCache
// Returns the value radix^(2^exponent) from cache. If this value not exist, it is added.
func getRadixConversionCache(radix types.Int, exponent types.Int) *BigInteger {
	cacheLine := powerCache[radix]
	if exponent < types.Int(len(cacheLine)) {
		return cacheLine[exponent]
	}

	oldLength := types.Int(len(cacheLine))
	for i := oldLength; i <= exponent; i++ {
		cacheLine = append(cacheLine, cacheLine[i-1].pow(2))
	}

	if exponent >= types.Int(len(powerCache[radix])) {
		powerCache[radix] = cacheLine
	}

	return cacheLine[exponent]
}

func (bi *BigInteger) getLowestSetBit() types.Int {
	lsb := lowestSetBitPlusTwo - 2
	if lsb == -2 { // lsb not initialized yet
		lsb = 0
		if bi.Signum == 0 {
			lsb -= 1
		} else {
			var i, b types.Int
			for i = types.Int(0); ; i++ {
				b = bi.getInt(i)
				if b != 0 {
					break
				}
			}
			lsb += (i << 5) + NumberOfTrailingZeros(b)
		}
		lowestSetBitPlusTwo = lsb + 2
	}
	return lsb
}

func (b *BigInteger) getInt(n types.Int) types.Int {
	if n < 0 {
		return 0
	}
	if n >= types.Int(len(b.Mag)) {
		return b.sigInt()
	}

	magInt := b.Mag[types.Int(len(b.Mag))-n-1]

	if b.Signum >= 0 {
		return magInt
	} else {
		if n <= b.firstNonzeroIntNum() {
			return -magInt
		} else {
			return ^magInt //
		}
	}
}

func (b *BigInteger) firstNonzeroIntNum() types.Int {
	fn := b.firstNonzeroIntNumPlusTwo - 2
	if fn == -2 {
		var i, mlen types.Int
		mlen = types.Int(len(b.Mag))
		for i = mlen - 1; i >= 0 && b.Mag[i] == 0; i-- {
		}
		fn = mlen - i - 1
		b.firstNonzeroIntNumPlusTwo = fn + 2 // offset by two to initialize
	}
	return fn
}

func (b *BigInteger) sigInt() types.Int {
	if b.Signum < 0 {
		return -1
	} else {
		return 0
	}
}

// Returns a negative BigInteger
func (b *BigInteger) negate() *BigInteger {
	return NewBigInteger(b.Mag, -b.Signum)
}

// add Adds the contents of the int arrays x and y
func add(x, y []types.Int) []types.Int {
	// if x is shorter, swap
	if len(x) < len(y) {
		var tmp = x
		x = y
		y = tmp
	}

	xIndex := len(x)
	yIndex := len(y)
	result := make([]types.Int, xIndex)
	var sum types.Long
	if yIndex == 1 {
		sum = types.Long(x[xIndex-1]) + types.Long(y[0])
		xIndex--
		result[xIndex] = sum.ToInt()
	} else {
		// Add common parts of both numbers
		for yIndex > 0 {
			xIndex--
			yIndex--
			sum = types.Long(x[xIndex]) +
				types.Long(y[yIndex]) + (sum.ShiftR(32)) // make sure positive
			result[xIndex] = sum.ToInt()
		}
	}

	carry := (sum.ShiftR(32)) != 0

	for xIndex > 0 && carry {
		xIndex--
		result[xIndex] = x[xIndex] + 1
		carry = result[xIndex] == 0
	}

	for xIndex > 0 {
		xIndex--
		result[xIndex] = x[xIndex]
	}

	if carry {
		bigger := make([]types.Int, len(result)+1)
		Arraycopy(result, 0, bigger, 1, types.Int(len(result)))
		bigger[0] = 0x01
		return bigger
	}
	return result
}

func (b *BigInteger) abs() *BigInteger {
	if b.Signum >= 0 {
		return b
	} else {
		return b.negate()
	}
}

func (b *BigInteger) pow(exponent types.Int) *BigInteger {
	if exponent < 0 {
		return nil
	}

	if b.Signum == 0 {
		if exponent == 0 {
			return ONE
		} else {
			return b
		}
	}

	partToSquare := b.abs()

	powersOfTwo := partToSquare.getLowestSetBit()
	bitsToShiftLong := types.Long(powersOfTwo * exponent)
	if bitsToShiftLong > LONG_MASK {
		panic(errors.New("overflow"))
	}
	bitsToShift := bitsToShiftLong.ToInt()

	var remainingBits types.Int

	if powersOfTwo > 0 {
		partToSquare = partToSquare.shiftRight(powersOfTwo)
		remainingBits = partToSquare.bitLength()
		if remainingBits == 1 {
			if b.Signum < 0 && (exponent&1) == 1 {
				return NEGATIVE_ONE.shiftLeft(bitsToShift)
			} else {
				return ONE.shiftLeft(bitsToShift)
			}
		}
	} else {
		remainingBits = partToSquare.bitLength()
		if remainingBits == 1 {
			if b.Signum < 0 && (exponent&1) == 1 {
				return NEGATIVE_ONE
			} else {
				return ONE
			}
		}
	}

	scaleFactor := types.Long(remainingBits * exponent)

	if len(partToSquare.Mag) == 1 && scaleFactor <= 62 {
		var newSign types.Int
		if b.Signum < 0 && (exponent&1) == 1 {
			newSign = -1
		} else {
			newSign = 1
		}

		result := types.Long(1)
		baseToPow2 := types.Long(partToSquare.Mag[0]) & LONG_MASK

		workingExponent := exponent

		for workingExponent != 0 {
			if (workingExponent & 1) != 0 {
				result *= baseToPow2
			}

			workingExponent = workingExponent.ShiftR(1)
			if workingExponent != 0 {
				baseToPow2 *= baseToPow2
			}
		}

		if powersOfTwo > 0 {
			if types.Long(bitsToShift)+scaleFactor <= 62 {
				return ValueOf((result << bitsToShiftLong) * newSign.ToLong())
			} else {
				return ValueOf(result * types.Long(newSign)).shiftLeft(bitsToShift)
			}
		} else {
			return ValueOf(result * types.Long(newSign))
		}
	} else {
		if types.Int(types.Long(b.bitLength())*types.Long(exponent)/types.Long(32)) > MAX_INT32/32+1 {
			panic(errors.New("overflow"))
		}

		answer := ONE

		workingExponent := exponent

		for workingExponent != 0 {
			if (workingExponent & 1) == 1 {
				answer = answer.Multiply(partToSquare)
			}

			workingExponent = workingExponent.ShiftR(1)
			if workingExponent != 0 {
				partToSquare = partToSquare.square()
			}
		}

		if powersOfTwo > 0 {
			answer = answer.shiftLeft(bitsToShift)
		}

		if b.Signum < 0 && (exponent&1) == 1 {
			return answer.negate()
		} else {
			return answer
		}
	}

}

func (b *BigInteger) shiftRight(n types.Int) *BigInteger {
	if b.Signum == 0 {
		return ZERO
	}
	if n > 0 {
		return b.shiftRightImpl(n)
	} else if n == 0 {
		return b
	} else {
		return NewBigInteger(shiftLeft(b.Mag, -n), b.Signum)
	}
}

func shiftLeft(mag []types.Int, n types.Int) []types.Int {
	nInts := n.ShiftR(5)
	nBits := n & 0x1f
	magLen := types.Int(len(mag))
	var newMag []types.Int

	if nBits == 0 {
		newMag = make([]types.Int, magLen+nInts)
		Arraycopy(mag, 0, newMag, 0, magLen)
	} else {
		i := 0
		nBits2 := 32 - nBits
		highBits := mag[0].ShiftR(nBits2)
		if highBits != 0 {
			newMag = make([]types.Int, magLen+nInts+1)
			newMag[i] = highBits
			i++
		} else {
			newMag = make([]types.Int, magLen+nInts)
		}
		j := types.Int(0)
		for j < magLen-1 {
			newMag[i] = mag[j]<<nBits | (mag[j+1].ShiftR(nBits2))
			i++
			j++
		}
		newMag[i] = mag[j] << nBits
	}

	return newMag
}

func (b *BigInteger) shiftRightImpl(n types.Int) *BigInteger {
	nInts := n.ShiftR(5)
	nBits := n & 0x1f
	magLen := types.Int(len(b.Mag))
	var newMag []types.Int

	if nInts >= magLen {
		if b.Signum >= 0 {
			return ZERO
		} else {
			return negConst[1]
		}
	}

	if nBits == 0 {
		newMagLen := magLen - nInts
		copy(newMag, b.Mag)
		newMag = tool.Copy(newMag, newMagLen)
	} else {
		i := 0
		highBits := b.Mag[0].ShiftR(nBits)
		if highBits != 0 {
			newMag = make([]types.Int, magLen-nInts)
			newMag[i] = highBits
			i++
		} else {
			newMag = make([]types.Int, magLen-nInts-1)
		}

		nBits2 := 32 - nBits
		j := types.Int(0)
		for j < magLen-nInts-1 {
			newMag[i] = (b.Mag[j] << nBits2) | (b.Mag[j+1].ShiftR(nBits))
			i++
			j++
		}
	}

	if b.Signum < 0 {
		onesLost := false
		i := magLen - 1
		j := magLen - nInts
		for ; i >= j && !onesLost; i-- {
			onesLost = b.Mag[i] != 0
		}

		if !onesLost && nBits != 0 {
			onesLost = b.Mag[magLen-nInts-1]<<(32-nBits) != 0
		}

		if onesLost {
			newMag = javaIncrement(newMag)
		}
	}

	return NewBigInteger(newMag, b.Signum)
}

func (b *BigInteger) shiftLeft(n types.Int) *BigInteger {
	if b.Signum == 0 {
		return ZERO
	}
	if n > 0 {
		return NewBigInteger(shiftLeft(b.Mag, n), b.Signum)
	} else if n == 0 {
		return b
	} else {
		return b.shiftRightImpl(-n)
	}
}

func (b *BigInteger) square() *BigInteger {
	return b.squareRec(false)
}

func (b *BigInteger) squareRec(isRecursion bool) *BigInteger {
	if b.Signum == 0 {
		return ZERO
	}
	length := types.Int(len(b.Mag))

	if length < KARATSUBA_SQUARE_THRESHOLD {
		z := squareToLen(b.Mag, length, nil)
		return NewBigInteger(trustedStripLeadingZeroInts(z), 1)
	} else {
		if length < TOOM_COOK_SQUARE_THRESHOLD {
			return b.squareKaratsuba()
		} else {
			if !isRecursion {
				if types.Long(bitLength(b.Mag, types.Int(len(b.Mag)))) > types.Long(16)*(MAX_INT32/32+1).ToLong() {
					panic(errors.New("overflow"))
				}
			}

			return b.squareToomCook3()
		}
	}
}

func (b *BigInteger) squareToomCook3() *BigInteger {
	length := types.Int(len(b.Mag))
	k := (length + 2) / 3
	r := length - 2*k

	var a0, a1, a2 *BigInteger
	a2 = b.getToomSlice(k, r, 0, length)
	a1 = b.getToomSlice(k, r, 1, length)
	a0 = b.getToomSlice(k, r, 2, length)
	var v0, v1, v2, vm1, vinf, t1, t2, tm1, da1 *BigInteger

	v0 = a0.squareRec(true)
	da1 = a2.Add(a0)
	vm1 = da1.subtract(a1).squareRec(true)
	da1 = da1.Add(a1)
	v1 = da1.squareRec(true)
	vinf = a2.squareRec(true)
	v2 = da1.Add(a2).shiftLeft(1).subtract(a0).squareRec(true)

	t2 = v2.subtract(vm1).exactDivideBy3()
	tm1 = v1.subtract(vm1).shiftRight(1)
	t1 = v1.subtract(v0)
	t2 = t2.subtract(t1).shiftRight(1)
	t1 = t1.subtract(tm1).subtract(vinf)
	t2 = t2.subtract(vinf.shiftLeft(1))
	tm1 = tm1.subtract(t2)

	ss := k * 32
	return vinf.shiftLeft(ss).Add(t2).shiftLeft(ss).Add(t1).shiftLeft(ss).Add(tm1).shiftLeft(ss).Add(v0)
}

func (b *BigInteger) squareKaratsuba() *BigInteger {
	half := types.Int(len(b.Mag)+1) / 2

	xl := b.getLower(half)
	xh := b.getUpper(half)

	xhs := xh.square() // xhs = xh ^ 2
	xls := xl.square() // xls = xl ^ 2

	return xhs.shiftLeft(half * 32).Add(xl.Add(xh).square().subtract(xhs.Add(xls))).shiftLeft(half * 32).Add(xls)
}

func (b *BigInteger) getLower(n types.Int) *BigInteger {
	length := types.Int(len(b.Mag))
	if length <= n {
		return b.abs()
	}

	lowerInts := make([]types.Int, n)
	Arraycopy(b.Mag, length-n, lowerInts, 0, n)

	return NewBigInteger(trustedStripLeadingZeroInts(lowerInts), 1)
}

func (b *BigInteger) getUpper(n types.Int) *BigInteger {
	length := types.Int(len(b.Mag))
	if length <= n {
		return ZERO
	}

	upperLen := length - n
	upperInts := make([]types.Int, upperLen)
	Arraycopy(b.Mag, 0, upperInts, 0, upperLen)

	return NewBigInteger(trustedStripLeadingZeroInts(upperInts), 1)
}

func (b *BigInteger) getToomSlice(lowerSize types.Int, upperSize types.Int, slice types.Int, fullsize types.Int) *BigInteger {
	var start, end, sliceSize, length, offset types.Int

	length = types.Int(len(b.Mag))
	offset = fullsize - length

	if slice == 0 {
		start = 0 - offset
		end = upperSize - 1 - offset
	} else {
		start = upperSize + (slice-1)*lowerSize - offset
		end = start + lowerSize - 1
	}

	if start < 0 {
		start = 0
	}
	if end < 0 {
		return ZERO
	}

	sliceSize = (end - start) + 1

	if sliceSize <= 0 {
		return ZERO
	}

	if start == 0 && sliceSize >= length {
		return b.abs()
	}

	intSlice := make([]types.Int, sliceSize)
	Arraycopy(b.Mag, start, intSlice, 0, sliceSize)

	return NewBigInteger(trustedStripLeadingZeroInts(intSlice), 1)

}

func (b *BigInteger) subtract(a1 *BigInteger) *BigInteger {
	return nil
}

func (b *BigInteger) exactDivideBy3() *BigInteger {
	length := types.Int(len(b.Mag))
	result := make([]types.Int, length)
	var x, w, q, borrow types.Long
	borrow = 0
	for i := length - 1; i >= 0; i-- {
		x = types.Long(b.Mag[i]) & LONG_MASK
		w = x - borrow
		if borrow > x {
			borrow = 1
		} else {
			borrow = 0
		}

		q = (w * 0xAAAAAAAB) & LONG_MASK
		result[i] = q.ToInt()

		if q >= 0x55555556 {
			borrow++
			if q >= 0xAAAAAAAB {
				borrow++
			}
		}
	}
	result = trustedStripLeadingZeroInts(result)
	return NewBigInteger(result, b.Signum)
}

func (b *BigInteger) Multiply(val *BigInteger) *BigInteger {
	return b.multiplyRec(val, false)
}

func (b *BigInteger) multiplyRec(val *BigInteger, isRecursion bool) *BigInteger {
	if val.Signum == 0 || b.Signum == 0 {
		return ZERO
	}

	xlen := types.Int(len(b.Mag))
	if val == b && xlen > MULTIPLY_SQUARE_THRESHOLD {
		return b.square()
	}

	ylen := types.Int(len(val.Mag))

	if (xlen < KARATSUBA_SQUARE_THRESHOLD) || (ylen < KARATSUBA_SQUARE_THRESHOLD) {
		var resultSign types.Int
		if val.Signum == b.Signum {
			resultSign = 1
		} else {
			resultSign = -1
		}
		if len(val.Mag) == 1 {
			return multiplyByInt(b.Mag, b.Mag[0], resultSign)
		}
		if len(b.Mag) == 1 {
			return multiplyByInt(val.Mag, b.Mag[0], resultSign)
		}
		result := multiplyToLen(b.Mag, xlen, val.Mag, ylen, nil)
		result = trustedStripLeadingZeroInts(result)
		return NewBigInteger(result, resultSign)
	} else {
		if (xlen < TOOM_COOK_SQUARE_THRESHOLD) && (ylen < TOOM_COOK_SQUARE_THRESHOLD) {
			return multiplyKaratsuba(b, val)
		} else {
			if !isRecursion {
				if types.Long(bitLength(b.Mag, types.Int(len(b.Mag)))+bitLength(val.Mag, types.Int(len(val.Mag)))) > types.Long(32)*(MAX_INT32/32+1).ToLong() {
					panic("overflow")
				}
			}

			return multiplyToomCook3(b, val)
		}
	}

}

func (b *BigInteger) smallToString(radix types.Int) string {
	if b.Signum == 0 {
		return "0"
	}

	maxNumDigitGroups := (4*len(b.Mag) + 6) / 7
	digitGroup := make([]string, maxNumDigitGroups)

	tmp := b.abs()
	numGroups := 0
	for tmp.Signum != 0 {
		d := longRadix[radix]

		q, a, b2 := NewMutableBigIntegerDefault(),
			NewMutableBigIntegerArray(tmp.Mag),
			NewMutableBigIntegerArray(d.Mag)

		r := a.Divide(b2, q)
		q2 := q.ToBigInteger(tmp.Signum * d.Signum)
		r2 := r.ToBigInteger(tmp.Signum * d.Signum)

		digitGroup[numGroups] = strconv.FormatInt(int64(r2.LongValue()), int(radix))
		numGroups++
		tmp = q2
	}

	var buf bytes.Buffer
	if b.Signum < 0 {
		buf.WriteString("-")
	}
	buf.WriteString(digitGroup[numGroups-1])
	for i := numGroups - 2; i >= 0; i-- {
		numLeadingZeros := digitsPerLong[radix] - types.Int(len(digitGroup[i]))
		if numLeadingZeros != 0 {
			buf.WriteString(zeros[:numLeadingZeros])
		}
		buf.WriteString(digitGroup[i])
	}
	return buf.String()
}

func (b *BigInteger) LongValue() types.Long {
	result := types.Long(0)
	for i := types.Int(1); i >= 0; i-- {
		result = (result << 32) + (b.getInt(i).ToLong() & LONG_MASK)
	}
	return result
}

func (b *BigInteger) DivideAndRemainder(val *BigInteger) []*BigInteger {
	if len(val.Mag) < BURNIKEL_ZIEGLER_THRESHOLD || len(b.Mag)-len(val.Mag) < BURNIKEL_ZIEGLER_OFFSET {
		return b.DivideAndRemainderKnuth(val)
	} else {
		return b.DivideAndRemainderBurnikelZiegler(val)
	}
}

func (b *BigInteger) DivideAndRemainderKnuth(val *BigInteger) []*BigInteger {
	result := make([]*BigInteger, 2)
	q := NewMutableBigIntegerDefault()
	a := NewMutableBigIntegerArray(b.Mag)
	bb := NewMutableBigIntegerArray(val.Mag)
	r := a.DivideKnuth(bb, q, true)
	if b.Signum == val.Signum {
		result[0] = q.ToBigInteger(1)
	} else {
		result[0] = q.ToBigInteger(0)
	}
	result[1] = r.ToBigInteger(b.Signum)
	return result
}

func (b *BigInteger) DivideAndRemainderBurnikelZiegler(val *BigInteger) []*BigInteger {
	q := NewMutableBigIntegerDefault()
	r := NewMutableBigIntegerByBigInteger(b).DivideAndRemainderBurnikelZiegler(NewMutableBigIntegerByBigInteger(val), q)
	var qBigInt, rBigInt *BigInteger
	if q.IsZero() {
		qBigInt = ZERO
	} else {
		qBigInt = q.ToBigInteger(b.Signum * val.Signum)
	}
	if r.IsZero() {
		rBigInt = ZERO
	} else {
		rBigInt = r.ToBigInteger(b.Signum)
	}
	return []*BigInteger{qBigInt, rBigInt}
}

func (b *BigInteger) compareMagnitute(val *BigInteger) types.Int {
	m1 := b.Mag
	len1 := types.Int(len(m1))
	m2 := val.Mag
	len2 := types.Int(len(m2))
	if len1 < len2 {
		return -1
	}
	if len1 > len2 {
		return 1
	}
	for i := types.Int(0); i < len1; i++ {
		a := m1[i]
		b2 := m2[i]
		if a != b2 {
			if (a.ToLong() & LONG_MASK) < (b2.ToLong() & LONG_MASK) {
				return -1
			} else {
				return 1
			}
		}
	}
	return 0
}

func multiplyToomCook3(a *BigInteger, b *BigInteger) *BigInteger {
	alen, blen := types.Int(len(a.Mag)), types.Int(len(b.Mag))
	largest := types.Int(math.Max(float64(alen), float64(blen)))
	k := (largest + 2) / 3
	r := largest - 2*k

	var a0, a1, a2, b0, b1, b2 *BigInteger
	a2 = a.getToomSlice(k, r, 0, largest)
	a1 = a.getToomSlice(k, r, 1, largest)
	a0 = a.getToomSlice(k, r, 2, largest)
	b2 = b.getToomSlice(k, r, 0, largest)
	b1 = b.getToomSlice(k, r, 1, largest)
	b0 = b.getToomSlice(k, r, 2, largest)

	var v0, v1, v2, vm1, vinf, t1, t2, tm1, da1, db1 *BigInteger
	v0 = a0.multiplyRec(b0, true)
	da1 = a2.Add(a0)
	db1 = b2.Add(b0)
	vm1 = da1.subtract(a1).multiplyRec(db1.subtract(b1), true)
	da1 = da1.Add(a1)
	db1 = db1.Add(b1)
	v1 = da1.multiplyRec(db1, true)
	v2 = da1.Add(a2).shiftLeft(1).subtract(a0).multiplyRec(
		db1.Add(b2).shiftLeft(1).subtract(b0), true)
	vinf = a2.multiplyRec(b2, true)

	t2 = v2.subtract(vm1).exactDivideBy3()
	tm1 = v1.subtract(vm1).shiftRight(1)
	t1 = v1.subtract(v0)
	t2 = t2.subtract(t1).shiftRight(1)
	t1 = t1.subtract(tm1).subtract(vinf)
	t2 = t2.subtract(vinf.shiftLeft(1))
	tm1 = tm1.subtract(t2)

	ss := k * 32
	result := vinf.shiftLeft(ss).Add(t2).shiftLeft(ss).Add(t1).shiftLeft(ss).Add(tm1).shiftLeft(ss).Add(v0)

	if a.Signum != b.Signum {
		return result.negate()
	} else {
		return result
	}
}

func multiplyKaratsuba(x *BigInteger, y *BigInteger) *BigInteger {
	xlen, ylen := types.Int(len(x.Mag)), types.Int(len(y.Mag))

	half := types.Int((math.Max(float64(xlen), float64(ylen)) + 1) / 2)

	xl := x.getLower(half)
	xh := x.getUpper(half)
	yl := y.getLower(half)
	yh := y.getUpper(half)

	p1 := xh.Multiply(yh) // p1 = xh*yh
	p2 := xl.Multiply(yl) // p2 = xl*yl

	// p3=(xh+xl)*(yh+yl)
	p3 := xh.Add(xl).Multiply(yh.Add(yl))

	// result = p1 * 2^(32*2*half) + (p3 - p1 - p2) * 2^(32*half) + p2
	result := p1.shiftLeft(32 * half).Add(p3.subtract(p1).subtract(p2)).shiftLeft(32 * half).Add(p2)

	if x.Signum != y.Signum {
		return result.negate()
	} else {
		return result
	}
}

func multiplyToLen(x []types.Int, xlen types.Int, y []types.Int, ylen types.Int, z []types.Int) []types.Int {
	multiplyToLenCheck(x, xlen)
	multiplyToLenCheck(y, ylen)
	return implMultiplyToLen(x, xlen, y, ylen, z)
}

func implMultiplyToLen(x []types.Int, xlen types.Int, y []types.Int, ylen types.Int, z []types.Int) []types.Int {
	xstart, ystart := xlen-1, ylen-1

	if z == nil || types.Int(len(z)) < (xlen+ylen) {
		z = make([]types.Int, xlen+ylen)
	}

	carry := types.Long(0)
	j, k := ystart, ystart+1+xstart
	for ; j >= 0; j-- {
		product := (types.Long(y[j])&LONG_MASK)*
			(types.Long(x[xstart])&LONG_MASK) + carry
		z[k] = product.ToInt()
		carry = product.ShiftR(32)
		k--
	}
	z[xstart] = carry.ToInt()

	for i := xstart - 1; i >= 0; i-- {
		carry = 0
		j, k = ystart, ystart+1+i
		for ; j >= 0; j-- {
			product := (types.Long(y[j])&LONG_MASK)*
				(types.Long(x[i])&LONG_MASK) +
				(types.Long(z[k]) & LONG_MASK) + carry
			z[k] = product.ToInt()
			k--
		}
		z[i] = carry.ToInt()
	}
	return z
}

func multiplyToLenCheck(array []types.Int, length types.Int) {
	if length <= 0 {
		return
	}
	if array == nil {
		panic(errors.New("object is nil"))
	}
	if length > types.Int(len(array)) {
		panic(errors.New(fmt.Sprintf("array index out of bound, index: %d", length-1)))
	}
}

func multiplyByInt(x []types.Int, y, sign types.Int) *BigInteger {
	if bitCount(y) == 1 {
		return NewBigInteger(shiftLeft(x, NumberOfTrailingZeros(y)), sign)
	}
	xlen := types.Int(len(x))
	rmag := make([]types.Int, xlen+1)
	carry, yl := types.Long(0), y.ToLong()&LONG_MASK
	rstart := types.Int(len(rmag)) - 1
	for i := xlen - 1; i >= 0; i-- {
		product := (x[i].ToLong()&LONG_MASK)*yl + carry
		rmag[rstart] = product.ToInt()
		rstart--
		carry = product.ShiftR(32)
	}
	if carry == 0 {
		rmag = tool.CopyRange(rmag, 1, types.Int(len(rmag)))
	} else {
		rmag[rstart] = carry.ToInt()
	}
	return NewBigInteger(rmag, sign)
}

func bitLength(val []types.Int, length types.Int) types.Int {
	if length == 0 {
		return 0
	}
	return ((length - 1) << 5) + BitLengthForInt(val[0])
}

func trustedStripLeadingZeroInts(val []types.Int) []types.Int {
	vlen := types.Int(len(val))
	var keep types.Int
	for keep = 0; keep < vlen && val[keep] == 0; keep++ {
	}
	if keep == 0 {
		return val
	} else {
		return tool.CopyRange(val, keep, vlen)
	}
}

func squareToLen(x []types.Int, length types.Int, z []types.Int) []types.Int {
	zlen := length << 1
	if z == nil || types.Int(len(z)) < zlen {
		z = make([]types.Int, zlen)
	}
	implSquareToLenChecks(x, length, z, zlen)
	return implSquareToLen(x, length, z, zlen)
}

func implSquareToLen(x []types.Int, length types.Int, z []types.Int, zlen types.Int) []types.Int {
	lastProductLowWord := types.Int(0)
	i := types.Int(0)
	j := types.Int(0)
	for ; j < length; j++ {
		piece := types.Long(x[j]) & LONG_MASK
		product := piece * piece
		z[i] = (lastProductLowWord << 31) | types.Int(product.ShiftR(33))
		i++
		z[i] = types.Int(product.ShiftR(1))
		i++
		lastProductLowWord = product.ToInt()
	}

	// Add in off-diagonal sums
	i = length
	offset := types.Int(1)
	for ; i > 0; i-- {
		t := x[i-1]
		t = mulAdd(z, x, offset, i-1, t)
		addOne(z, offset-1, i, t)
		offset += 2
	}

	primitiveLeftShift(z, zlen, 1)
	z[zlen-1] |= x[length-1] & 1

	return z
}

func primitiveLeftShift(a []types.Int, length types.Int, n types.Int) {
	if length == 0 || n == 0 {
		return
	}
	n2 := 32 - n
	i := types.Int(0)
	c := a[i]
	m := i + length - 1
	for ; i < m; i++ {
		b := c
		c = a[i+1]
		a[i] = (b << n) | (c.ShiftR(n2))
	}
	a[length-1] <<= n
}

func addOne(a []types.Int, offset types.Int, mlen types.Int, carry types.Int) types.Int {
	offset = types.Int(len(a)) - 1 - mlen - offset
	t := (types.Long(a[offset]) & LONG_MASK) + (types.Long(carry) & LONG_MASK)

	a[offset] = t.ToInt()
	if t.ShiftR(32) == 0 {
		return 0
	}
	for mlen-1 > 0 {
		mlen--
		offset--
		if offset < 0 {
			return 1
		} else {
			a[offset]++
			if a[offset] != 0 {
				return 0
			}
		}
	}
	return 1
}

func mulAdd(out []types.Int, in []types.Int, offset types.Int, length types.Int, k types.Int) types.Int {
	implMulAddCheck(out, in, offset, length, k)
	return implMulAdd(out, in, offset, length, k)
}

func implMulAdd(out []types.Int, in []types.Int, offset types.Int, length types.Int, k types.Int) types.Int {
	kLong := types.Long(k) & LONG_MASK
	carry := types.Long(0)

	offset = types.Int(len(out)) - offset - 1
	for j := length - 1; j >= 0; j-- {
		product := (types.Long(in[j])&LONG_MASK)*kLong +
			(types.Long(out[offset]) & LONG_MASK) + carry
		out[offset] = product.ToInt()
		offset--
		carry = product.ShiftR(32)
	}
	return carry.ToInt()
}

func implMulAddCheck(out []types.Int, in []types.Int, offset types.Int, length types.Int, k types.Int) {
	if length > types.Int(len(in)) {
		panic(errors.New(fmt.Sprintf("input length is out of bound: %d > %d", length, len(in))))
	}
	if offset > 0 {
		panic(errors.New(fmt.Sprintf("input offset is invalid: %d", offset)))
	}
	if offset > types.Int(len(out)-1) {
		panic(errors.New(fmt.Sprintf("input offset is out of bound: %d > %d", offset, len(out)-1)))
	}
	if length > types.Int(len(out))-offset {
		panic(errors.New(fmt.Sprintf("input length is out of bound: %d > %d", len(out), types.Int(len(out))-offset)))
	}
}

func implSquareToLenChecks(x []types.Int, length types.Int, z []types.Int, zlen types.Int) {
	if length < 1 {
		panic(errors.New(fmt.Sprintf("invalid input length: %d", length)))
	}
	if length > types.Int(len(x)) {
		panic(errors.New(fmt.Sprintf("input length out of bound: %d > %d", length, len(x))))

	}
	if length*2 > types.Int(len(z)) {
		panic(errors.New(fmt.Sprintf("input length out of bound: %d > %d", length*2, len(z))))
	}
	if zlen < 1 {
		panic(errors.New(fmt.Sprintf("invalid input length: %d", zlen)))
	}
	if zlen < types.Int(len(z)) {
		panic(errors.New(fmt.Sprintf("input length out of bound: %d > %d", length, len(z))))
	}
}

func javaIncrement(val []types.Int) []types.Int {
	lastSum := types.Int(0)
	for i := len(val) - 1; i >= 0 && lastSum == 0; i-- {
		val[i] += 1
		lastSum = val[i]
	}
	if lastSum == 0 {
		val = make([]types.Int, len(val)+1)
		val[0] = 1
	}
	return val
}

func NumberOfTrailingZeros(i types.Int) types.Int {
	var y, n types.Int
	if i == 0 {
		return 32
	}
	n = 31
	y = i << 16
	if y != 0 {
		n = n - 16
		i = y
	}
	y = i << 8
	if y != 0 {
		n = n - 8
		i = y
	}
	y = i << 4
	if y != 0 {
		n = n - 4
		i = y
	}
	y = i << 2
	if y != 0 {
		n = n - 2
		i = y
	}
	return n - (i << 1).ShiftR(31)
}
