package bigger

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sineycoder/go-bigger/tool"
	"github.com/sineycoder/go-bigger/types"
	"math"
	"strconv"
	"strings"
)

/**
 @author: nizhenxian
 @date: 2021/8/10 16:54:18
**/

const (
	MAX_INT32                              = types.Int(0x7fffffff)
	MAX_INT64                              = types.Long(0x7fffffffffffffff)
	MIN_INT32                              = ^MAX_INT32
	MIN_INT64                              = ^MAX_INT64
	pMAX_CONSTANT                          = 16
	p_KARATSUBA_SQUARE_THRESHOLD           = 128
	p_TOOM_COOK_SQUARE_THRESHOLD           = 216
	p_MULTIPLY_SQUARE_THRESHOLD            = 20
	p_SCHOENHAGE_BASE_CONVERSION_THRESHOLD = 20
	p_BURNIKEL_ZIEGLER_THRESHOLD           = 80
	p_BURNIKEL_ZIEGLER_OFFSET              = 40
	p_MAX_MAG_LENGTH                       = MAX_INT32/32 + 1
)

var (
	ZERO                = newBigInteger([]types.Int{0}, 0)
	ONE                 = BigIntegerValueOf(1)
	TWO                 = BigIntegerValueOf(2)
	TEN                 = BigIntegerValueOf(10)
	NEGATIVE_ONE        = BigIntegerValueOf(-1)
	p_LOG_TWO           = types.Double(math.Log(2.0))
	p_LONG_MASK         = types.Long(0xffffffff)
	posConst            = make([]*bigInteger, pMAX_CONSTANT+1)
	negConst            = make([]*bigInteger, pMAX_CONSTANT+1)
	logCache            = make([]types.Double, 32+1)
	powerCache          = make([][]*bigInteger, 32+1)
	zeros               = "000000000000000000000000000000000000000000000000000000000000000" // the length of zeros, length=63
	lowestSetBitPlusTwo types.Int
	intRadix            = []types.Int{0, 0,
		0x40000000, 0x4546b3db, 0x40000000, 0x48c27395, 0x159fd800,
		0x75db9c97, 0x40000000, 0x17179149, 0x3b9aca00, 0xcc6db61,
		0x19a10000, 0x309f1021, 0x57f6c100, 0xa2f1b6f, 0x10000000,
		0x18754571, 0x247dbc80, 0x3547667b, 0x4c4b4000, 0x6b5a6e1d,
		0x6c20a40, 0x8d2d931, 0xb640000, 0xe8d4a51, 0x1269ae40,
		0x17179149, 0x1cb91000, 0x23744899, 0x2b73a840, 0x34e63b41,
		0x40000000, 0x4cfa3cc1, 0x5c13d840, 0x6d91b519, 0x39aa400,
	}
	bitsPerDigit = []types.Int{0, 0,
		1024, 1624, 2048, 2378, 2648, 2875, 3072, 3247, 3402, 3543, 3672,
		3790, 3899, 4001, 4096, 4186, 4271, 4350, 4426, 4498, 4567, 4633,
		4696, 4756, 4814, 4870, 4923, 4975, 5025, 5074, 5120, 5166, 5210,
		5253, 5295}
	digitsPerInt = []types.Int{0, 0, 30, 19, 15, 13, 11,
		11, 10, 9, 9, 8, 8, 8, 8, 7, 7, 7, 7, 7, 7, 7, 6, 6, 6, 6,
		6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 5}
	digitsPerLong = []types.Int{0, 0,
		62, 39, 31, 27, 24, 22, 20, 19, 18, 18, 17, 17, 16, 16, 15, 15, 15, 14,
		14, 14, 14, 13, 13, 13, 13, 13, 13, 12, 12, 12, 12, 12, 12, 12, 12}
	longRadix = []*bigInteger{nil, nil,
		BigIntegerValueOf(0x4000000000000000), BigIntegerValueOf(0x383d9170b85ff80b),
		BigIntegerValueOf(0x4000000000000000), BigIntegerValueOf(0x6765c793fa10079d),
		BigIntegerValueOf(0x41c21cb8e1000000), BigIntegerValueOf(0x3642798750226111),
		BigIntegerValueOf(0x1000000000000000), BigIntegerValueOf(0x12bf307ae81ffd59),
		BigIntegerValueOf(0xde0b6b3a7640000), BigIntegerValueOf(0x4d28cb56c33fa539),
		BigIntegerValueOf(0x1eca170c00000000), BigIntegerValueOf(0x780c7372621bd74d),
		BigIntegerValueOf(0x1e39a5057d810000), BigIntegerValueOf(0x5b27ac993df97701),
		BigIntegerValueOf(0x1000000000000000), BigIntegerValueOf(0x27b95e997e21d9f1),
		BigIntegerValueOf(0x5da0e1e53c5c8000), BigIntegerValueOf(0xb16a458ef403f19),
		BigIntegerValueOf(0x16bcc41e90000000), BigIntegerValueOf(0x2d04b7fdd9c0ef49),
		BigIntegerValueOf(0x5658597bcaa24000), BigIntegerValueOf(0x6feb266931a75b7),
		BigIntegerValueOf(0xc29e98000000000), BigIntegerValueOf(0x14adf4b7320334b9),
		BigIntegerValueOf(0x226ed36478bfa000), BigIntegerValueOf(0x383d9170b85ff80b),
		BigIntegerValueOf(0x5a3c23e39c000000), BigIntegerValueOf(0x4e900abb53e6b71),
		BigIntegerValueOf(0x7600ec618141000), BigIntegerValueOf(0xaee5720ee830681),
		BigIntegerValueOf(0x1000000000000000), BigIntegerValueOf(0x172588ad4f5f0981),
		BigIntegerValueOf(0x211e44f7d02c1000), BigIntegerValueOf(0x2ee56725f06e5c71),
		BigIntegerValueOf(0x41c21cb8e1000000)}
)

type bigInteger struct {
	signum                    types.Int   // -1 for negative, 0 for zero, 1 for positive
	mag                       []types.Int // order
	firstNonzeroIntNumPlusTwo types.Int
	bitLengthPlusOne          types.Int
}

func init() {
	// cache
	for i := types.Int(1); i <= pMAX_CONSTANT; i++ {
		posConst[i] = &bigInteger{
			signum: 1,
			mag:    []types.Int{i},
		}
		negConst[i] = &bigInteger{
			signum: -1,
			mag:    []types.Int{i},
		}
	}

	for i := 2; i <= 32; i++ {
		powerCache[i] = []*bigInteger{
			BigIntegerValueOf(types.Long(i)),
		}
		logCache[i] = types.Double(math.Log(float64(i)))
	}
}

func destructiveMulAdd(x []types.Int, y, z types.Int) {
	ylong := y.ToLong() & p_LONG_MASK
	zlong := z.ToLong() & p_LONG_MASK
	length := types.Int(len(x))

	var product, carry types.Long
	for i := length - 1; i >= 0; i-- {
		product = ylong*(x[i].ToLong()&p_LONG_MASK) + carry
		x[i] = product.ToInt()
		carry = product.ShiftR(32)
	}

	sum := (x[length-1].ToLong() & p_LONG_MASK) + zlong
	x[length-1] = sum.ToInt()
	carry = sum.ShiftR(32)
	for i := length - 2; i >= 0; i-- {
		sum = (x[i].ToLong() & p_LONG_MASK) + carry
		x[i] = sum.ToInt()
		carry = sum.ShiftR(32)
	}
}

func subtract2(val types.Long, little []types.Int) []types.Int {
	highWord := val.ShiftR(32).ToInt()
	if highWord == 0 {
		result := make([]types.Int, 1)
		result[0] = (val - (little[0].ToLong() & p_LONG_MASK)).ToInt()
		return result
	} else {
		result := make([]types.Int, 2)
		if len(little) == 1 {
			difference := (val.ToInt().ToLong() & p_LONG_MASK) - (little[0].ToLong() - p_LONG_MASK)
			result[1] = difference.ToInt()
			borrow := (difference >> 32) != 0
			if borrow {
				result[0] = highWord - 1
			} else {
				result[0] = highWord
			}
			return result
		} else {
			difference := (val.ToInt().ToLong() & p_LONG_MASK) - (little[1].ToLong() & p_LONG_MASK)
			result[1] = difference.ToInt()
			difference = (highWord.ToLong() & p_LONG_MASK) - (little[0].ToLong() & p_LONG_MASK) + (difference >> 32)
			result[0] = difference.ToInt()
			return result
		}
	}
}

func subtract(big []types.Int, val types.Long) []types.Int {
	highWord := val.ShiftR(32).ToInt()
	bigIndex := types.Int(len(big))
	result := make([]types.Int, bigIndex)
	difference := types.Long(0)

	if highWord == 0 {
		bigIndex--
		difference = (big[bigIndex].ToLong() & p_LONG_MASK) - val
		result[bigIndex] = difference.ToInt()
	} else {
		bigIndex--
		difference = (big[bigIndex].ToLong() & p_LONG_MASK) - (val & p_LONG_MASK)
		result[bigIndex] = difference.ToInt()
		bigIndex--
		difference = (big[bigIndex].ToLong() & p_LONG_MASK) - (highWord.ToLong() & p_LONG_MASK) + (difference >> 32)
		result[bigIndex] = difference.ToInt()
	}

	borrow := (difference >> 32) != 0
	for bigIndex > 0 && borrow {
		bigIndex--
		result[bigIndex] -= 1
		borrow = result[bigIndex] == -1
	}

	for bigIndex > 0 {
		bigIndex--
		result[bigIndex] = big[bigIndex]
	}

	return result
}

func subtract_(big, little []types.Int) []types.Int {
	bigIndex := types.Int(len(big))
	result := make([]types.Int, bigIndex)
	littleIndex := types.Int(len(little))
	difference := types.Long(0)

	for littleIndex > 0 {
		bigIndex--
		littleIndex--
		difference = (big[bigIndex].ToLong() & p_LONG_MASK) -
			(little[littleIndex].ToLong() & p_LONG_MASK) +
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

func newBigInteger(magnitude []types.Int, signum types.Int) *bigInteger {
	bigInteger := &bigInteger{}
	if len(magnitude) == 0 {
		bigInteger.signum = 0
	} else {
		bigInteger.signum = signum
	}
	bigInteger.mag = magnitude
	return bigInteger
}

func bitCount(i types.Int) types.Int {
	i = i - (i.ShiftR(1) & 0x55555555)
	i = (i & 0x33333333) + (i.ShiftR(2) & 0x33333333)
	i = i + (i.ShiftR(4))&0x0f0f0f0f
	i = i + i.ShiftR(8)
	i = i + i.ShiftR(16)
	return i & 0x3f
}

func bitLengthForInt(n types.Int) types.Int {
	return 32 - NumberOfLeadingZeros(n)
}

/*
toString Converts the specified bigInteger to a string and appends to buf.
*/
func toString(u *bigInteger, buf *bytes.Buffer, radix types.Int, digits types.Int) {

	if len(u.mag) <= p_SCHOENHAGE_BASE_CONVERSION_THRESHOLD {
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
	b = u.BitLength()

	// Calculate a value for n in the equation radix^(2^n) = u
	n = types.Int(math.Round(math.Log(float64(types.Double(b)*p_LOG_TWO))/float64(p_LOG_TWO) - 1.0))
	v := getRadixConversionCache(radix, n)
	var result []*bigInteger
	result = u.DivideAndRemainder(v)

	expectedDigits := types.Int(1 << n)

	toString(result[0], buf, radix, digits-expectedDigits)
	toString(result[1], buf, radix, expectedDigits)
}

// getRadixConversionCache
// Returns the value radix^(2^exponent) from cache. If this value not exist, it is added.
func getRadixConversionCache(radix types.Int, exponent types.Int) *bigInteger {
	cacheLine := powerCache[radix]
	if exponent < types.Int(len(cacheLine)) {
		return cacheLine[exponent]
	}

	oldLength := types.Int(len(cacheLine))
	for i := oldLength; i <= exponent; i++ {
		cacheLine = append(cacheLine, cacheLine[i-1].Pow(2))
	}

	if exponent >= types.Int(len(powerCache[radix])) {
		powerCache[radix] = cacheLine
	}

	return cacheLine[exponent]
}

func (bi *bigInteger) getLowestSetBit() types.Int {
	lsb := lowestSetBitPlusTwo - 2
	if lsb == -2 { // lsb not initialized yet
		lsb = 0
		if bi.signum == 0 {
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

func (b *bigInteger) getInt(n types.Int) types.Int {
	if n < 0 {
		return 0
	}
	if n >= types.Int(len(b.mag)) {
		return b.sigInt()
	}
	magInt := b.mag[types.Int(len(b.mag))-n-1]
	if b.signum >= 0 {
		return magInt
	} else {
		if n <= b.firstNonzeroIntNum() {
			return -magInt
		} else {
			return ^magInt //
		}
	}
}

func (b *bigInteger) firstNonzeroIntNum() types.Int {
	fn := b.firstNonzeroIntNumPlusTwo - 2
	if fn == -2 {
		var i, mlen types.Int
		mlen = types.Int(len(b.mag))
		for i = mlen - 1; i >= 0 && b.mag[i] == 0; i-- {
		}
		fn = mlen - i - 1
		b.firstNonzeroIntNumPlusTwo = fn + 2 // offset by two to initialize
	}
	return fn
}

func (b *bigInteger) sigInt() types.Int {
	if b.signum < 0 {
		return -1
	} else {
		return 0
	}
}

// Returns a negative bigInteger
func (b *bigInteger) negate() *bigInteger {
	return newBigInteger(b.mag, -b.signum)
}

func add_(x []types.Int, val types.Long) []types.Int {
	var result []types.Int
	var sum types.Long
	xIndex := types.Int(len(x))
	highWord := val.ShiftR(32).ToInt()
	if highWord == 0 {
		result = make([]types.Int, xIndex)
		xIndex--
		sum = (x[xIndex].ToLong() & p_LONG_MASK) + val
		result[xIndex] = sum.ToInt()
	} else {
		if xIndex == 1 {
			result = make([]types.Int, 2)
			sum = val + (x[0].ToLong() & p_LONG_MASK)
			result[1] = sum.ToInt()
			result[0] = sum.ShiftR(32).ToInt()
			return result
		} else {
			result = make([]types.Int, xIndex)
			xIndex--
			sum = (x[xIndex].ToLong() & p_LONG_MASK) + (val & p_LONG_MASK)
			result[xIndex] = sum.ToInt()
			xIndex--
			sum = (x[xIndex].ToLong() & p_LONG_MASK) + (highWord.ToLong() & p_LONG_MASK) + sum.ShiftR(32)
			result[xIndex] = sum.ToInt()
		}
	}
	carry := sum.ShiftR(32) != 0
	for xIndex > 0 && carry {
		xIndex--
		result[xIndex] = x[xIndex] + 1
		carry = result[xIndex] == 0
	}

	for xIndex > 0 {
		xIndex--
		result[xIndex] = x[xIndex] + 1
	}

	if carry {
		bigger := make([]types.Int, len(result)+1)
		tool.Arraycopy(result, 0, bigger, 1, types.Int(len(result)))
		bigger[0] = 0x01
		return bigger
	}
	return result
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
		sum = (x[xIndex-1].ToLong() & p_LONG_MASK) + (y[0].ToLong() & p_LONG_MASK)
		xIndex--
		result[xIndex] = sum.ToInt()
	} else {
		// Add common parts of both numbers
		for yIndex > 0 {
			xIndex--
			yIndex--
			sum = (x[xIndex].ToLong() & p_LONG_MASK) +
				(y[yIndex].ToLong() & p_LONG_MASK) + (sum.ShiftR(32)) // make sure positive
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
		tool.Arraycopy(result, 0, bigger, 1, types.Int(len(result)))
		bigger[0] = 0x01
		return bigger
	}
	return result
}

func (b *bigInteger) Pow(exponent types.Int) *bigInteger {
	if exponent < 0 {
		panic(errors.New("Nagative exponent"))
	}

	if b.signum == 0 {
		if exponent == 0 {
			return ONE
		} else {
			return b
		}
	}

	partToSquare := b.Abs()
	powersOfTwo := partToSquare.getLowestSetBit()
	bitsToShiftLong := (powersOfTwo * exponent).ToLong()
	if bitsToShiftLong > p_LONG_MASK {
		panic(errors.New("overflow"))
	}
	bitsToShift := bitsToShiftLong.ToInt()

	var remainingBits types.Int
	if powersOfTwo > 0 {
		partToSquare = partToSquare.shiftRight(powersOfTwo)
		remainingBits = partToSquare.BitLength()
		if remainingBits == 1 {
			if b.signum < 0 && (exponent&1) == 1 {
				return NEGATIVE_ONE.shiftLeft(bitsToShift)
			} else {
				return ONE.shiftLeft(bitsToShift)
			}
		}
	} else {
		remainingBits = partToSquare.BitLength()
		if remainingBits == 1 {
			if b.signum < 0 && (exponent&1) == 1 {
				return NEGATIVE_ONE
			} else {
				return ONE
			}
		}
	}

	scaleFactor := (remainingBits * exponent).ToLong()
	if len(partToSquare.mag) == 1 && scaleFactor <= 62 {
		var newSign types.Int
		if b.signum < 0 && (exponent&1) == 1 {
			newSign = -1
		} else {
			newSign = 1
		}

		result := types.Long(1)
		baseToPow2 := partToSquare.mag[0].ToLong() & p_LONG_MASK

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
			if bitsToShift.ToLong()+scaleFactor <= 62 {
				return BigIntegerValueOf((result << bitsToShiftLong) * newSign.ToLong())
			} else {
				return BigIntegerValueOf(result * newSign.ToLong()).shiftLeft(bitsToShift)
			}
		} else {
			return BigIntegerValueOf(result * newSign.ToLong())
		}
	} else {
		if (b.BitLength().ToLong() * exponent.ToLong() / types.Long(32)).ToInt() > p_MAX_MAG_LENGTH {
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

		if b.signum < 0 && (exponent&1) == 1 {
			return answer.negate()
		} else {
			return answer
		}
	}
}

func (b *bigInteger) shiftRight(n types.Int) *bigInteger {
	if b.signum == 0 {
		return ZERO
	}
	if n > 0 {
		return b.shiftRightImpl(n)
	} else if n == 0 {
		return b
	} else {
		return newBigInteger(shiftLeft(b.mag, -n), b.signum)
	}
}

func shiftLeft(mag []types.Int, n types.Int) []types.Int {
	nInts := n.ShiftR(5)
	nBits := n & 0x1f
	magLen := types.Int(len(mag))
	var newMag []types.Int

	if nBits == 0 {
		newMag = make([]types.Int, magLen+nInts)
		tool.Arraycopy(mag, 0, newMag, 0, magLen)
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

func (b *bigInteger) shiftRightImpl(n types.Int) *bigInteger {
	nInts := n.ShiftR(5)
	nBits := n & 0x1f
	magLen := types.Int(len(b.mag))
	var newMag []types.Int

	if nInts >= magLen {
		if b.signum >= 0 {
			return ZERO
		} else {
			return negConst[1]
		}
	}
	if nBits == 0 {
		newMagLen := magLen - nInts
		copy(newMag, b.mag)
		newMag = tool.Copy(newMag, newMagLen)
	} else {
		i := 0
		highBits := b.mag[0].ShiftR(nBits)
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
			newMag[i] = (b.mag[j] << nBits2) | (b.mag[j+1].ShiftR(nBits))
			i++
			j++
		}
	}
	if b.signum < 0 {
		onesLost := false
		i := magLen - 1
		j := magLen - nInts
		for ; i >= j && !onesLost; i-- {
			onesLost = b.mag[i] != 0
		}

		if !onesLost && nBits != 0 {
			onesLost = b.mag[magLen-nInts-1]<<(32-nBits) != 0
		}

		if onesLost {
			newMag = increment(newMag)
		}
	}
	return newBigInteger(newMag, b.signum)
}

func (b *bigInteger) shiftLeft(n types.Int) *bigInteger {
	if b.signum == 0 {
		return ZERO
	}
	if n > 0 {
		return newBigInteger(shiftLeft(b.mag, n), b.signum)
	} else if n == 0 {
		return b
	} else {
		return b.shiftRightImpl(-n)
	}
}

func (b *bigInteger) square() *bigInteger {
	return b.squareRec(false)
}

func (b *bigInteger) squareRec(isRecursion bool) *bigInteger {
	if b.signum == 0 {
		return ZERO
	}
	length := types.Int(len(b.mag))

	if length < p_KARATSUBA_SQUARE_THRESHOLD {
		z := squareToLen(b.mag, length, nil)
		return newBigInteger(trustedStripLeadingZeroInts(z), 1)
	} else {
		if length < p_TOOM_COOK_SQUARE_THRESHOLD {
			return b.squareKaratsuba()
		} else {
			if !isRecursion {
				if bitLength(b.mag, types.Int(len(b.mag))).ToLong() > types.Long(16)*(p_MAX_MAG_LENGTH).ToLong() {
					panic(errors.New("overflow"))
				}
			}
			return b.squareToomCook3()
		}
	}
}

func (b *bigInteger) squareToomCook3() *bigInteger {
	length := types.Int(len(b.mag))
	k := (length + 2) / 3
	r := length - 2*k

	var a0, a1, a2 *bigInteger
	a2 = b.getToomSlice(k, r, 0, length)
	a1 = b.getToomSlice(k, r, 1, length)
	a0 = b.getToomSlice(k, r, 2, length)
	var v0, v1, v2, vm1, vinf, t1, t2, tm1, da1 *bigInteger

	v0 = a0.squareRec(true)
	da1 = a2.Add(a0)
	vm1 = da1.Subtract(a1).squareRec(true)
	da1 = da1.Add(a1)
	v1 = da1.squareRec(true)
	vinf = a2.squareRec(true)
	v2 = da1.Add(a2).shiftLeft(1).Subtract(a0).squareRec(true)

	t2 = v2.Subtract(vm1).exactDivideBy3()
	tm1 = v1.Subtract(vm1).shiftRight(1)
	t1 = v1.Subtract(v0)
	t2 = t2.Subtract(t1).shiftRight(1)
	t1 = t1.Subtract(tm1).Subtract(vinf)
	t2 = t2.Subtract(vinf.shiftLeft(1))
	tm1 = tm1.Subtract(t2)

	ss := k * 32
	return vinf.shiftLeft(ss).Add(t2).shiftLeft(ss).Add(t1).shiftLeft(ss).Add(tm1).shiftLeft(ss).Add(v0)
}

func (b *bigInteger) squareKaratsuba() *bigInteger {
	half := types.Int(len(b.mag)+1) / 2

	xl := b.getLower(half)
	xh := b.getUpper(half)
	xhs := xh.square() // xhs = xh ^ 2
	xls := xl.square() // xls = xl ^ 2

	return xhs.shiftLeft(half * 32).Add(xl.Add(xh).square().Subtract(xhs.Add(xls))).shiftLeft(half * 32).Add(xls)
}

func (b *bigInteger) getLower(n types.Int) *bigInteger {
	length := types.Int(len(b.mag))
	if length <= n {
		return b.Abs()
	}

	lowerInts := make([]types.Int, n)
	tool.Arraycopy(b.mag, length-n, lowerInts, 0, n)

	return newBigInteger(trustedStripLeadingZeroInts(lowerInts), 1)
}

func (b *bigInteger) getUpper(n types.Int) *bigInteger {
	length := types.Int(len(b.mag))
	if length <= n {
		return ZERO
	}

	upperLen := length - n
	upperInts := make([]types.Int, upperLen)
	tool.Arraycopy(b.mag, 0, upperInts, 0, upperLen)

	return newBigInteger(trustedStripLeadingZeroInts(upperInts), 1)
}

func (b *bigInteger) getToomSlice(lowerSize types.Int, upperSize types.Int, slice types.Int, fullsize types.Int) *bigInteger {
	var start, end, sliceSize, length, offset types.Int

	length = types.Int(len(b.mag))
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
		return b.Abs()
	}

	intSlice := make([]types.Int, sliceSize)
	tool.Arraycopy(b.mag, start, intSlice, 0, sliceSize)

	return newBigInteger(trustedStripLeadingZeroInts(intSlice), 1)

}

func (b *bigInteger) exactDivideBy3() *bigInteger {
	length := types.Int(len(b.mag))
	result := make([]types.Int, length)
	var x, w, q, borrow types.Long
	borrow = 0
	for i := length - 1; i >= 0; i-- {
		x = b.mag[i].ToLong() & p_LONG_MASK
		w = x - borrow
		if borrow > x {
			borrow = 1
		} else {
			borrow = 0
		}

		q = (w * 0xAAAAAAAB) & p_LONG_MASK
		result[i] = q.ToInt()

		if q >= 0x55555556 {
			borrow++
			if q >= 0xAAAAAAAB {
				borrow++
			}
		}
	}
	result = trustedStripLeadingZeroInts(result)
	return newBigInteger(result, b.signum)
}

func (b *bigInteger) multiplyRec(val *bigInteger, isRecursion bool) *bigInteger {
	if val.signum == 0 || b.signum == 0 {
		return ZERO
	}

	xlen := types.Int(len(b.mag))
	if val == b && xlen > p_MULTIPLY_SQUARE_THRESHOLD {
		return b.square()
	}

	ylen := types.Int(len(val.mag))

	if (xlen < p_KARATSUBA_SQUARE_THRESHOLD) || (ylen < p_KARATSUBA_SQUARE_THRESHOLD) {
		var resultSign types.Int
		if val.signum == b.signum {
			resultSign = 1
		} else {
			resultSign = -1
		}
		if len(val.mag) == 1 {
			return multiplyByInt(b.mag, val.mag[0], resultSign)
		}
		if len(b.mag) == 1 {
			return multiplyByInt(val.mag, b.mag[0], resultSign)
		}
		result := multiplyToLen(b.mag, xlen, val.mag, ylen, nil)
		result = trustedStripLeadingZeroInts(result)
		return newBigInteger(result, resultSign)
	} else {
		if (xlen < p_TOOM_COOK_SQUARE_THRESHOLD) && (ylen < p_TOOM_COOK_SQUARE_THRESHOLD) {
			return multiplyKaratsuba(b, val)
		} else {
			if !isRecursion {
				if (bitLength(b.mag, types.Int(len(b.mag))) + bitLength(val.mag, types.Int(len(val.mag)))).ToLong() > types.Long(32)*(p_MAX_MAG_LENGTH).ToLong() {
					panic("overflow")
				}
			}

			return multiplyToomCook3(b, val)
		}
	}

}

func (b *bigInteger) smallToString(radix types.Int) string {
	if b.signum == 0 {
		return "0"
	}

	maxNumDigitGroups := (4*len(b.mag) + 6) / 7
	digitGroup := make([]string, maxNumDigitGroups)
	tmp := b.Abs()
	numGroups := 0
	for tmp.signum != 0 {
		d := longRadix[radix]

		q, a, b2 := newMutableBigIntegerDefault(),
			newMutableBigIntegerArray(tmp.mag),
			newMutableBigIntegerArray(d.mag)

		r := a.Divide(b2, q)
		q2 := q.toBigInteger(tmp.signum * d.signum)
		r2 := r.toBigInteger(tmp.signum * d.signum)

		digitGroup[numGroups] = strconv.FormatInt(int64(r2.LongValue()), int(radix))
		numGroups++
		tmp = q2
	}

	var buf bytes.Buffer
	if b.signum < 0 {
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

func (b *bigInteger) divideAndRemainderKnuth(val *bigInteger) []*bigInteger {
	result := make([]*bigInteger, 2)
	q := newMutableBigIntegerDefault()
	a := newMutableBigIntegerArray(b.mag)
	bb := newMutableBigIntegerArray(val.mag)
	r := a.divideKnuth(bb, q, true)
	if b.signum == val.signum {
		result[0] = q.toBigInteger(1)
	} else {
		result[0] = q.toBigInteger(0)
	}
	result[1] = r.toBigInteger(b.signum)
	return result
}

func (b *bigInteger) divideAndRemainderBurnikelZiegler(val *bigInteger) []*bigInteger {
	q := newMutableBigIntegerDefault()
	r := newMutableBigIntegerByBigInteger(b).DivideAndRemainderBurnikelZiegler(newMutableBigIntegerByBigInteger(val), q)
	var qBigInt, rBigInt *bigInteger
	if q.IsZero() {
		qBigInt = ZERO
	} else {
		qBigInt = q.toBigInteger(b.signum * val.signum)
	}
	if r.IsZero() {
		rBigInt = ZERO
	} else {
		rBigInt = r.toBigInteger(b.signum)
	}
	return []*bigInteger{qBigInt, rBigInt}
}

func (b *bigInteger) compareMagnituteLong(val types.Long) types.Int {
	if val != MIN_INT64 {
		m1 := b.mag
		length := types.Int(len(m1))
		if length > 2 {
			return 1
		}
		if val < 0 {
			val = -val
		}
		highWord := val.ShiftR(32).ToInt()
		if highWord == 0 {
			if length < 1 {
				return -1
			}
			if length > 1 {
				return 1
			}
			a := m1[0]
			bb := val.ToInt()
			if a != bb {
				if (a.ToLong() & p_LONG_MASK) < (bb.ToLong() & p_LONG_MASK) {
					return -1
				}
				return 1
			}
			return 0
		} else {
			if length < 2 {
				return -1
			}
			a := m1[0]
			bb := highWord
			if a != bb {
				if (a.ToLong() & p_LONG_MASK) < (bb.ToLong() & p_LONG_MASK) {
					return -1
				}
				return 1
			}
			a = m1[1]
			bb = val.ToInt()
			if a != bb {
				if (a.ToLong() & p_LONG_MASK) < (bb.ToLong() & p_LONG_MASK) {
					return -1
				}
				return 1
			}
			return 0
		}
	}
	panic("illegal param")
}

func (b *bigInteger) compareMagnitute(val *bigInteger) types.Int {
	m1 := b.mag
	len1 := types.Int(len(m1))
	m2 := val.mag
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
			if (a.ToLong() & p_LONG_MASK) < (b2.ToLong() & p_LONG_MASK) {
				return -1
			} else {
				return 1
			}
		}
	}
	return 0
}

func (b *bigInteger) divideKnuth(val *bigInteger) *bigInteger {
	q := newMutableBigIntegerDefault()
	a := newMutableBigIntegerArray(b.mag)
	b2 := newMutableBigIntegerArray(val.mag)
	a.divideKnuth(b2, q, false)
	return q.toBigInteger(b.signum * val.signum)
}

func (b *bigInteger) divideBurnikelZiegler(val *bigInteger) *bigInteger {
	return b.divideAndRemainderBurnikelZiegler(val)[0]
}

func (b *bigInteger) Subtract(val *bigInteger) *bigInteger {
	if val.signum == 0 {
		return b
	}
	if b.signum == 0 {
		return val.negate()
	}
	if val.signum != b.signum {
		return newBigInteger(add(b.mag, val.mag), b.signum)
	}

	cmp := b.compareMagnitute(val)
	if cmp == 0 {
		return ZERO
	}
	var resultMag []types.Int
	if cmp > 0 {
		resultMag = subtract_(b.mag, val.mag)
	} else {
		resultMag = subtract_(val.mag, b.mag)
	}
	resultMag = trustedStripLeadingZeroInts(resultMag)
	if cmp == b.signum {
		return newBigInteger(resultMag, 1)
	} else {
		return newBigInteger(resultMag, -1)
	}
}
func (b *bigInteger) Multiply(val *bigInteger) *bigInteger {
	return b.multiplyRec(val, false)
}

// LongValue if this bigInteger is too bigger to fit in a long, only the low-order 64 bits are returned.
func (b *bigInteger) LongValue() types.Long {
	result := types.Long(0)
	for i := types.Int(1); i >= 0; i-- {
		result = (result << 32) + (b.getInt(i).ToLong() & p_LONG_MASK)
	}
	return result
}

func (b *bigInteger) DivideAndRemainder(val *bigInteger) []*bigInteger {
	if len(val.mag) < p_BURNIKEL_ZIEGLER_THRESHOLD || len(b.mag)-len(val.mag) < p_BURNIKEL_ZIEGLER_OFFSET {
		return b.divideAndRemainderKnuth(val)
	} else {
		return b.divideAndRemainderBurnikelZiegler(val)
	}
}

func (b *bigInteger) Divide(val *bigInteger) *bigInteger {
	if len(val.mag) < p_BURNIKEL_ZIEGLER_THRESHOLD ||
		len(b.mag)-len(val.mag) < p_BURNIKEL_ZIEGLER_OFFSET {
		return b.divideKnuth(val)
	} else {
		return b.divideBurnikelZiegler(val)
	}
}

func (b *bigInteger) Sqrt() *bigInteger {
	if b.signum < 0 {
		panic(errors.New("negative BigIntager"))
	}

	return newMutableBigIntegerArray(b.mag).sqrt().ToBigIntegerDefault()
}

// LongValueExact this bigInteger converted to a long. different from LongValue, this func will throw panic error
func (b *bigInteger) LongValueExact() types.Long {
	if len(b.mag) <= 2 && b.BitLength() <= 63 {
		return b.LongValue()
	} else {
		panic(errors.New("bigInteger out of long range"))
	}
}

func (b *bigInteger) checkRange() {
	if types.Int(len(b.mag)) > p_MAX_MAG_LENGTH || types.Int(len(b.mag)) == p_MAX_MAG_LENGTH && b.mag[0] < 0 {
		panic(errors.New("overflow"))
	}
}

func bigLengthForInt(n types.Int) types.Int {
	return 32 - NumberOfLeadingZeros(n)
}

func multiplyToomCook3(a *bigInteger, b *bigInteger) *bigInteger {
	alen, blen := types.Int(len(a.mag)), types.Int(len(b.mag))
	largest := types.Int(math.Max(float64(alen), float64(blen)))
	k := (largest + 2) / 3
	r := largest - 2*k

	var a0, a1, a2, b0, b1, b2 *bigInteger
	a2 = a.getToomSlice(k, r, 0, largest)
	a1 = a.getToomSlice(k, r, 1, largest)
	a0 = a.getToomSlice(k, r, 2, largest)
	b2 = b.getToomSlice(k, r, 0, largest)
	b1 = b.getToomSlice(k, r, 1, largest)
	b0 = b.getToomSlice(k, r, 2, largest)

	var v0, v1, v2, vm1, vinf, t1, t2, tm1, da1, db1 *bigInteger
	v0 = a0.multiplyRec(b0, true)
	da1 = a2.Add(a0)
	db1 = b2.Add(b0)
	vm1 = da1.Subtract(a1).multiplyRec(db1.Subtract(b1), true)
	da1 = da1.Add(a1)
	db1 = db1.Add(b1)
	v1 = da1.multiplyRec(db1, true)
	v2 = da1.Add(a2).shiftLeft(1).Subtract(a0).multiplyRec(
		db1.Add(b2).shiftLeft(1).Subtract(b0), true)
	vinf = a2.multiplyRec(b2, true)

	t2 = v2.Subtract(vm1).exactDivideBy3()
	tm1 = v1.Subtract(vm1).shiftRight(1)
	t1 = v1.Subtract(v0)
	t2 = t2.Subtract(t1).shiftRight(1)
	t1 = t1.Subtract(tm1).Subtract(vinf)
	t2 = t2.Subtract(vinf.shiftLeft(1))
	tm1 = tm1.Subtract(t2)

	ss := k * 32
	result := vinf.shiftLeft(ss).Add(t2).shiftLeft(ss).Add(t1).shiftLeft(ss).Add(tm1).shiftLeft(ss).Add(v0)

	if a.signum != b.signum {
		return result.negate()
	} else {
		return result
	}
}

func multiplyKaratsuba(x *bigInteger, y *bigInteger) *bigInteger {
	xlen, ylen := types.Int(len(x.mag)), types.Int(len(y.mag))

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
	result := p1.shiftLeft(32 * half).Add(p3.Subtract(p1).Subtract(p2)).shiftLeft(32 * half).Add(p2)

	if x.signum != y.signum {
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
		product := (y[j].ToLong()&p_LONG_MASK)*
			(x[xstart].ToLong()&p_LONG_MASK) + carry
		z[k] = product.ToInt()
		carry = product.ShiftR(32)
		k--
	}
	z[xstart] = carry.ToInt()

	for i := xstart - 1; i >= 0; i-- {
		carry = 0
		j, k = ystart, ystart+1+i
		for ; j >= 0; j-- {
			product := (y[j].ToLong()&p_LONG_MASK)*
				(x[i].ToLong()&p_LONG_MASK) +
				(z[k].ToLong() & p_LONG_MASK) + carry
			z[k] = product.ToInt()
			carry = product.ShiftR(32)
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

func multiplyByInt(x []types.Int, y, sign types.Int) *bigInteger {
	if bitCount(y) == 1 {
		return newBigInteger(shiftLeft(x, NumberOfTrailingZeros(y)), sign)
	}
	xlen := types.Int(len(x))
	rmag := make([]types.Int, xlen+1)
	carry, yl := types.Long(0), y.ToLong()&p_LONG_MASK
	rstart := types.Int(len(rmag)) - 1
	for i := xlen - 1; i >= 0; i-- {
		product := (x[i].ToLong()&p_LONG_MASK)*yl + carry
		rmag[rstart] = product.ToInt()
		rstart--
		carry = product.ShiftR(32)
	}
	if carry == 0 {
		rmag = tool.CopyRange(rmag, 1, types.Int(len(rmag)))
	} else {
		rmag[rstart] = carry.ToInt()
	}
	return newBigInteger(rmag, sign)
}

func bitLength(val []types.Int, length types.Int) types.Int {
	if length == 0 {
		return 0
	}
	return ((length - 1) << 5) + bitLengthForInt(val[0])
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
		piece := x[j].ToLong() & p_LONG_MASK
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
	t := (a[offset].ToLong() & p_LONG_MASK) + (carry.ToLong() & p_LONG_MASK)

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
	kLong := k.ToLong() & p_LONG_MASK
	carry := types.Long(0)

	offset = types.Int(len(out)) - offset - 1
	for j := length - 1; j >= 0; j-- {
		product := (in[j].ToLong()&p_LONG_MASK)*kLong +
			(out[offset].ToLong() & p_LONG_MASK) + carry
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
	if offset < 0 {
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

func increment(val []types.Int) []types.Int {
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

func newBigIntegerCharArray(val []uint8, sign, length types.Int) *bigInteger {
	b := &bigInteger{}
	var cursor, numDigits types.Int

	for cursor < length && tool.Digit(val[cursor], 10) == 0 {
		cursor++
	}
	if cursor == length {
		b.signum = 0
		b.mag = ZERO.mag
		return b
	}

	numDigits = length - cursor
	b.signum = sign
	var numWords types.Int
	if length < 10 {
		numWords = 1
	} else {
		numBits := (numDigits * bitsPerDigit[10]).ShiftR(10) + 1
		if (numBits + 31).ToLong() >= 1<<32 {
			panic(errors.New("overflow"))
		}
		numWords = (numBits + 31).ShiftR(5)
	}
	magnitude := make([]types.Int, numWords)

	firstGroupLen := numDigits % digitsPerInt[10]
	if firstGroupLen == 0 {
		firstGroupLen = digitsPerInt[10]
	}

	i, _ := strconv.ParseInt(string(val)[cursor:cursor+firstGroupLen], 10, 32)
	cursor += firstGroupLen
	magnitude[numWords-1] = types.Int(i)

	for cursor < length {
		i, _ = strconv.ParseInt(string(val)[cursor:cursor+digitsPerInt[10]], 10, 32)
		cursor += digitsPerInt[10]
		groupVal := types.Int(i)
		destructiveMulAdd(magnitude, intRadix[10], groupVal)
	}
	b.mag = trustedStripLeadingZeroInts(magnitude)
	if types.Int(len(b.mag)) >= p_MAX_MAG_LENGTH {
		b.checkRange()
	}
	return b
}

func (b *bigInteger) DoubleValue() types.Double {
	if b.signum == 0 {
		return 0.0
	}

	exponent := ((types.Int(len(b.mag)) - 1) << 5) + bigLengthForInt(b.mag[0]) - 1

	if exponent < 63 {
		return b.LongValue().ToDouble()
	} else if exponent > 1023 {
		if b.signum > 0 {
			return types.POSITIVE_INFINITY
		} else {
			return types.NEGATIVE_INFINITY
		}
	}

	shift := exponent - 53
	var twiceSignifFloor types.Long
	var nBits, nBits2, highBits, lowBits types.Int

	nBits = shift & 0x1f
	nBits2 = 32 - nBits

	if nBits == 0 {
		highBits = b.mag[0]
		lowBits = b.mag[1]
	} else {
		highBits = b.mag[0].ShiftR(nBits)
		lowBits = (b.mag[0] << nBits2) | b.mag[1].ShiftR(nBits)
		if highBits == 0 {
			highBits = lowBits
			lowBits = (b.mag[1] << nBits2) | b.mag[2].ShiftR(nBits)
		}
	}

	twiceSignifFloor = ((highBits.ToLong() & p_LONG_MASK) << 32) | (lowBits.ToLong() & p_LONG_MASK)

	signifFloor := twiceSignifFloor >> 1
	signifFloor &= 0x000FFFFFFFFFFFFF // remove the implied bit

	increment := (twiceSignifFloor&1) != 0 && ((signifFloor&1) != 0 || b.Abs().getLowestSetBit() < shift)
	signifRounded := types.Long(0)
	if increment {
		signifRounded = signifFloor + 1
	} else {
		signifRounded = signifFloor
	}
	bits := (exponent + 1023).ToLong() << 52

	bits += signifRounded
	bits |= b.signum.ToLong() & MIN_INT64
	return types.DoubleFromBits(bits)
}

func (b *bigInteger) SqrtAndRemainder() []*bigInteger {
	s := b.Sqrt()
	r := b.Subtract(s.square())
	if r.CompareTo(ZERO) < 0 {
		panic(errors.New("remainder value < 0"))
	}
	return []*bigInteger{s, r}
}

func (b *bigInteger) CompareTo(val *bigInteger) types.Int {
	if b.signum == val.signum {
		switch b.signum {
		case 1:
			return b.compareMagnitute(val)
		case -1:
			return val.compareMagnitute(b)
		default:
			return 0
		}
	}
	if b.signum > val.signum {
		return 1
	} else {
		return -1
	}
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

// BigIntegerValueOf, if |val| <= 16, return posConst cache
func BigIntegerValueOf(val types.Long) *bigInteger {
	if val == 0 {
		return ZERO
	}
	if val > 0 && val <= pMAX_CONSTANT {
		return posConst[val]
	} else if val < 0 && val >= -pMAX_CONSTANT {
		return negConst[-val]
	}
	return NewBigIntegerLong(val)
}

func NewBigIntegerLong(val types.Long) *bigInteger {
	bigInteger := &bigInteger{}
	if val < 0 {
		val = -val
		bigInteger.signum = -1 // set signum as -1, value is negative
	} else {
		bigInteger.signum = 1
	}

	highBit := (val >> 32).ToInt()
	if highBit == 0 {
		bigInteger.mag = []types.Int{val.ToInt()} // high 32 bits is all zero
	} else {
		bigInteger.mag = []types.Int{highBit, val.ToInt()}
	}
	return bigInteger
}

func NewBigIntegerString(val string) *bigInteger {
	return NewBigIntegerStringRadix(val, 10)
}

func NewBigIntegerStringRadix(val string, radix types.Int) *bigInteger {
	b := &bigInteger{}
	var cursor, numDigits types.Int
	length := types.Int(len(val))

	if radix < 2 || radix > 36 {
		panic(errors.New("Radix out of range"))
	}
	if length == 0 {
		panic(errors.New("Zero length bigInteger"))
	}

	sign := 1
	index1 := strings.LastIndex(val, "-")
	index2 := strings.LastIndex(val, "+")
	if index1 >= 0 {
		if index1 != 0 || index2 >= 0 {
			panic(errors.New("Illegal embedded sign character"))
		}
		sign = -1
		cursor = 1
	} else if index2 >= 0 {
		if index2 != 0 {
			panic(errors.New("Illegal embedded sign character"))
		}
		cursor = 1
	}
	if cursor == length {
		panic(errors.New("Zero length bigInteger"))
	}

	for cursor < length && tool.Digit(val[cursor], uint8(radix)) == 0 {
		cursor++
	}

	if cursor == length {
		b.signum = 0
		b.mag = ZERO.mag
		return b
	}

	numDigits = length - cursor
	b.signum = types.Int(sign)

	numBits := ((numDigits * bitsPerDigit[radix]).ShiftR(10) + 1).ToLong()
	if numBits+31 >= (types.Long(1) << 32) {
		panic(errors.New("overflow"))
	}
	numWords := (numBits + 31).ToInt().ShiftR(5)
	magnitude := make([]types.Int, numWords)

	firstGroupLen := numDigits % digitsPerInt[radix]
	if firstGroupLen == 0 {
		firstGroupLen = digitsPerInt[radix]
	}
	group := val[cursor : cursor+firstGroupLen]
	cursor += firstGroupLen
	res, err := strconv.ParseInt(group, int(radix), 32)
	if err != nil {
		panic(fmt.Sprintf("Input string error: %+v", err))
	}
	magnitude[numWords-1] = types.Int(res)
	if magnitude[numWords-1] < 0 {
		panic(errors.New("Illegal digit"))
	}

	superRadix := intRadix[radix]
	var groupVal types.Int
	for cursor < length {
		group = val[cursor : cursor+digitsPerInt[radix]]
		cursor += digitsPerInt[radix]
		res, _ = strconv.ParseInt(group, int(radix), 32)
		groupVal = types.Int(res)
		if groupVal < 0 {
			panic(errors.New("Illegal digit"))
		}
		destructiveMulAdd(magnitude, superRadix, groupVal)
	}
	b.mag = trustedStripLeadingZeroInts(magnitude)
	if types.Int(len(b.mag)) >= p_MAX_MAG_LENGTH {
		b.checkRange()
	}
	return b
}

func (b *bigInteger) Add(val *bigInteger) *bigInteger {
	if val.signum == 0 {
		return b
	}
	if b.signum == 0 {
		return val
	}
	if val.signum == b.signum {
		return newBigInteger(add(b.mag, val.mag), b.signum)
	}

	cmp := b.compareMagnitute(val)
	if cmp == 0 {
		return ZERO
	}
	var resultMag []types.Int
	if cmp > 0 {
		resultMag = subtract_(b.mag, val.mag)
	} else {
		resultMag = subtract_(val.mag, b.mag)
	}
	resultMag = trustedStripLeadingZeroInts(resultMag)
	if cmp == b.signum {
		return newBigInteger(resultMag, 1)
	} else {
		return newBigInteger(resultMag, -1)
	}
}

func (b *bigInteger) String() string {
	return b.StringRadix(10)
}

func (b *bigInteger) StringRadix(radix types.Int) string {
	if b.signum == 0 {
		return "0"
	}
	if radix < 2 || radix > 36 {
		radix = 10
	}

	if len(b.mag) <= p_SCHOENHAGE_BASE_CONVERSION_THRESHOLD {
		return b.smallToString(radix)
	}

	var buf bytes.Buffer
	if b.signum < 0 {
		toString(b.negate(), &buf, radix, 0)
		return "-" + buf.String()
	} else {
		toString(b, &buf, radix, 0)
	}
	return buf.String()
}

// BitLength Returns the number of bits in the minimal two's-complement representation of this bigInteger, excluding a sign bit.
func (b *bigInteger) BitLength() types.Int {
	n := b.bitLengthPlusOne - 1
	if n == -1 {
		var m []types.Int
		length := types.Int(len(m))
		if length == 0 {
			n = 0
		} else {
			magBitLength := ((length - 1) << 5) + bitLengthForInt(b.mag[0])
			if b.signum < 0 {
				pow2 := bitCount(b.mag[0]) == 1
				for i := types.Int(1); i < length && pow2; i++ {
					pow2 = b.mag[i] == 0
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

func NumberOfLeadingZerosForLong(i types.Long) types.Int {
	x := i.ShiftR(32).ToInt()
	if x == 0 {
		return 32 + NumberOfLeadingZeros(i.ToInt())
	} else {
		return NumberOfLeadingZeros(x)
	}
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

func (b *bigInteger) Abs() *bigInteger {
	if b.signum >= 0 {
		return b
	} else {
		return b.negate()
	}
}

func (b *bigInteger) add(val types.Long) *bigInteger {
	if val == 0 {
		return b
	}
	if b.signum == 0 {
		return BigIntegerValueOf(val)
	}
	if (val>>63)|(val.ShiftR(63)) == b.signum.ToLong() {
		return newBigInteger(add_(b.mag, types.Long(math.Abs(float64(val)))), b.signum)
	}
	cmp := b.compareMagnituteLong(val)
	if cmp == 0 {
		return ZERO
	}
	var resultMag []types.Int
	if cmp > 0 {
		resultMag = subtract(b.mag, val.Abs())
	} else {
		resultMag = subtract2(val.Abs(), b.mag)
	}
	resultMag = trustedStripLeadingZeroInts(resultMag)
	if cmp == b.signum {
		return newBigInteger(resultMag, 1)
	} else {
		return newBigInteger(resultMag, -1)
	}
}

func (b *bigInteger) multiplyLong(v types.Long) *bigInteger {
	if v == 0 || b.signum == 0 {
		return ZERO
	}
	if v == MIN_INT64 {
		return b.Multiply(BigIntegerValueOf(v))
	}
	var rsign types.Int
	if v > 0 {
		rsign = b.signum
	} else {
		rsign = -b.signum
	}
	if v < 0 {
		v = -v
	}
	dh := v.ShiftR(32)
	dl := v & p_LONG_MASK

	xlen := types.Int(len(b.mag))
	value := b.mag
	var rmag []types.Int
	if dh == 0 {
		rmag = make([]types.Int, xlen+1)
	} else {
		rmag = make([]types.Int, xlen+2)
	}
	var carry types.Long
	rstart := types.Int(len(rmag)) - 1
	for i := xlen - 1; i >= 0; i-- {
		product := (value[i].ToLong()&p_LONG_MASK)*dl + carry
		rmag[rstart] = product.ToInt()
		rstart--
		carry = product.ShiftR(32)
	}
	rmag[rstart] = carry.ToInt()
	if dh != 0 {
		carry = 0
		rstart = types.Int(len(rmag)) - 2
		for i := xlen - 1; i >= 0; i-- {
			product := (value[i].ToLong()&p_LONG_MASK)*dh + (rmag[rstart].ToLong() & p_LONG_MASK) + carry
			rmag[rstart] = product.ToInt()
			rstart--
			carry = product.ShiftR(32)
		}
		rmag[0] = carry.ToInt()
	}
	if carry == 0 {
		rmag = tool.CopyRange(rmag, 1, types.Int(len(rmag)))
	}
	return newBigInteger(rmag, rsign)
}
