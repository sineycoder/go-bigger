package bigger

import (
	"errors"
	"github.com/sineycoder/go-bigger/tool"
	"github.com/sineycoder/go-bigger/types"
	"sync"
)

/**
 @author: nizhenxian
 @date: 2021/8/14 11:54:01
**/

type RoundingMode types.Int

var mu sync.Mutex

type BigDecimal struct {
	intVal     *BigInteger
	scale      types.Int
	precision  types.Int
	intCompact types.Long
}

type mathContext struct {
	precision    types.Int
	roundingMode RoundingMode
}

const (
	ROUND_UP = RoundingMode(iota)
	ROUND_DOWN
	ROUND_CEILING
	ROUND_FLOOR
	ROUDINGMODE_HALF_UP
	ROUND_HALF_DOWN
	ROUND_HALF_EVEN
	ROUND_UNNECESSARY
)

var (
	ZERO_THROUGH_TEN = []*BigDecimal{
		newBigDecimalByBigInteger(ZERO, 0, 0, 1),
		newBigDecimalByBigInteger(ONE, 1, 0, 1),
		newBigDecimalByBigInteger(TWO, 2, 0, 1),
		newBigDecimalByBigInteger(BigIntegerValueOf(3), 3, 0, 1),
		newBigDecimalByBigInteger(BigIntegerValueOf(4), 4, 0, 1),
		newBigDecimalByBigInteger(BigIntegerValueOf(5), 5, 0, 1),
		newBigDecimalByBigInteger(BigIntegerValueOf(6), 6, 0, 1),
		newBigDecimalByBigInteger(BigIntegerValueOf(7), 7, 0, 1),
		newBigDecimalByBigInteger(BigIntegerValueOf(8), 8, 0, 1),
		newBigDecimalByBigInteger(BigIntegerValueOf(9), 9, 0, 1),
		newBigDecimalByBigInteger(TEN, 10, 0, 1),
	}
	LONG_TEN_POWERS_TABLE = []types.Long{
		1,                   // 0 / 10^0
		10,                  // 1 / 10^1
		100,                 // 2 / 10^2
		1000,                // 3 / 10^3
		10000,               // 4 / 10^4
		100000,              // 5 / 10^5
		1000000,             // 6 / 10^6
		10000000,            // 7 / 10^7
		100000000,           // 8 / 10^8
		1000000000,          // 9 / 10^9
		10000000000,         // 10 / 10^10
		100000000000,        // 11 / 10^11
		1000000000000,       // 12 / 10^12
		10000000000000,      // 13 / 10^13
		100000000000000,     // 14 / 10^14
		1000000000000000,    // 15 / 10^15
		10000000000000000,   // 16 / 10^16
		100000000000000000,  // 17 / 10^17
		1000000000000000000, // 18 / 10^18
	}
	BIG_TEN_POWERS_TABLE = []*BigInteger{
		ONE,
		BigIntegerValueOf(10),
		BigIntegerValueOf(100),
		BigIntegerValueOf(1000),
		BigIntegerValueOf(10000),
		BigIntegerValueOf(100000),
		BigIntegerValueOf(1000000),
		BigIntegerValueOf(10000000),
		BigIntegerValueOf(100000000),
		BigIntegerValueOf(1000000000),
		BigIntegerValueOf(10000000000),
		BigIntegerValueOf(100000000000),
		BigIntegerValueOf(1000000000000),
		BigIntegerValueOf(10000000000000),
		BigIntegerValueOf(100000000000000),
		BigIntegerValueOf(1000000000000000),
		BigIntegerValueOf(10000000000000000),
		BigIntegerValueOf(100000000000000000),
		BigIntegerValueOf(1000000000000000000),
	}
)

func newBigDecimalByBigInteger(intVal *BigInteger, val types.Long, scale, prec types.Int) *BigDecimal {
	return &BigDecimal{
		scale:      scale,
		precision:  prec,
		intCompact: val,
		intVal:     intVal,
	}
}

func NewBigDecimalString(val string) *BigDecimal {
	if val == "" {
		panic("illegal value")
	}
	var offset, length, prec, scl types.Int
	var rs types.Long
	var rb *BigInteger
	var mc = &mathContext{roundingMode: ROUDINGMODE_HALF_UP}
	length = types.Int(len(val))

	isneg := false // whether positive
	if val[offset] == '-' {
		isneg = true
		offset++
		length--
	} else if val[offset] == '+' { // leading + allowed
		offset++
		length--
	}

	dot := false
	exp := types.Long(0)
	var c uint8
	isCompact := length <= 18

	idx := types.Int(0)
	if isCompact {
		for ; length > 0; offset++ {
			c = val[offset]
			if c == '0' {
				if prec == 0 {
					prec = 1
				} else if rs != 0 {
					rs *= 10
					prec++
				}
				if dot {
					scl++
				}
			} else if c >= '1' && c <= '9' {
				digit := types.Int(c - '0')
				if prec != 1 || rs != 0 {
					prec++
				}
				rs = rs*10 + digit.ToLong()
				if dot {
					scl++
				}
			} else if c == '.' {
				if dot {
					panic(errors.New("Character array contains more than one point"))
				}
				dot = true
			} else if c <= '9' && c >= '0' {
				digit := types.Long(tool.Digit(c, 10))
				if digit == 0 {
					if prec == 0 {
						prec = 1
					} else if rs != 0 {
						rs *= 10
						prec++
					}
				} else {
					if prec != 1 || rs != 0 {
						prec++
					}
					rs = rs*10 + digit
				}
				if dot {
					scl++
				}
			} else if c == 'e' || c == 'E' {
				exp = parseExp(val, offset, length)
				if exp.ToInt().ToLong() != exp {
					// overflow
					panic(errors.New("Exponent overflow"))
				}
				break
			} else {
				panic(errors.New("illegal character"))
			}
			length--
		}
		if prec == 0 {
			panic(errors.New("no digits found"))
		}
		if exp != 0 {
			scl = adjustScale(scl, exp)
		}
		if isneg {
			rs = -rs
		}
		mcp := mc.precision
		drop := prec - mcp

		if mcp > 0 && drop > 0 {
			for drop > 0 {
				scl = checkScaleNonZero(scl.ToLong() - drop.ToLong())
				rs = divideAndRound(rs, LONG_TEN_POWERS_TABLE[drop], mc.roundingMode)
				prec = longDigitLength(rs)
				drop = prec - mcp
			}
		}
	} else {
		coeff := make([]uint8, length)
		for ; length > 0; offset++ {
			if c >= '0' && c <= '9' {
				if c == '0' || tool.Digit(c, 10) == 0 {
					if prec == 0 {
						coeff[idx] = c
						prec = 1
					} else if idx != 0 {
						coeff[idx] = c
						idx++
						prec++
					}
				} else {
					if prec != 1 || idx != 0 {
						prec++
					}
					coeff[idx] = c
					idx++
				}
				if dot {
					scl++
				}
				continue
			}
			if c == '.' {
				if dot {
					panic(errors.New("Character array contains more than one point"))
				}
				dot = true
				continue
			}
			if c != 'e' && c != 'E' {
				panic(errors.New("Character array is missing exponent mark"))
			}
			exp = parseExp(val, offset, length)
			if exp.ToInt().ToLong() != exp {
				panic(errors.New("Exponent overflow"))
			}
			length--
			break
		}
		if prec == 0 {
			panic(errors.New("No digit found"))
		}
		if exp != 0 {
			scl = adjustScale(scl, exp)
		}
		if isneg {
			rb = newBigIntegerCharArray(coeff, -1, prec)
		} else {
			rb = newBigIntegerCharArray(coeff, 1, prec)
		}
		rs = compactValFor(rb)
		mcp := mc.precision
		if mcp > 0 && (prec > mcp) {
			if rs == MIN_INT64 {
				drop := prec - mcp
				for drop > 0 {
					scl = checkScaleNonZero(scl.ToLong() - drop.ToLong())
					rb = divideAndRoundByTenPow(rb, drop, mc.roundingMode)
					rs = compactValFor(rb)
					if rs != MIN_INT64 {
						prec = longDigitLength(rs)
						break
					}
					prec = bigDigitLength(rb)
					drop = prec - mcp
				}
			}
			if rs != MIN_INT64 {
				drop := prec - mcp
				for drop > 0 {
					scl = checkScaleNonZero(scl.ToLong() - drop.ToLong())
					rs = divideAndRound(rs, LONG_TEN_POWERS_TABLE[drop], mc.roundingMode)
					prec = longDigitLength(rs)
					drop = prec - mcp
				}
				rb = nil
			}
		}
	}
	return &BigDecimal{
		scale:      scl,
		precision:  prec,
		intCompact: rs,
		intVal:     rb,
	}
}

func bigDigitLength(b *BigInteger) types.Int {
	if b.signum == 0 {
		return 1
	}
	r := ((b.BitLength() + 1) * 646456993).ShiftR(31)
	if b.compareMagnitute(bigTenToThe(r)) < 0 {
		return r
	}
	return r + 1
}

func divideAndRoundByTenPow(intVal *BigInteger, tenPow types.Int, roundingMode RoundingMode) *BigInteger {
	if tenPow < types.Int(len(LONG_TEN_POWERS_TABLE)) {
		intVal = divideAndRoundByBigInteger(intVal, LONG_TEN_POWERS_TABLE[tenPow], roundingMode)
	} else {
		intVal = divideAndRoundByBigInteger2(intVal, bigTenToThe(tenPow), roundingMode)
	}
	return intVal
}

func divideAndRoundByBigInteger2(bdividend *BigInteger, bdivisor *BigInteger, roundingMode RoundingMode) *BigInteger {
	var isRemainderZero bool
	var qsign types.Int
	mdividend := newMutableBigIntegerArray(bdividend.mag)
	mq := newMutableBigIntegerDefault()
	mdivisor := newMutableBigIntegerArray(bdivisor.mag)
	mr := mdividend.Divide(mdivisor, mq)
	isRemainderZero = mr.IsZero()
	if bdividend.signum != bdivisor.signum {
		qsign = -1
	} else {
		qsign = 1
	}
	if !isRemainderZero {
		if needIncrementMutableBigInteger2(mdivisor, roundingMode, qsign, mq, mr) {
			mq.add(mutable_one)
		}
	}
	return mq.toBigInteger(qsign)
}

func needIncrementMutableBigInteger2(mdivisor *mutableBigInteger, roundingMode RoundingMode, qsign types.Int, mq *mutableBigInteger, mr *mutableBigInteger) bool {
	if !mr.IsZero() {
		cmpFracHalf := mr.compareHalf(mdivisor)
		return commonNeedIncrement(roundingMode, qsign, cmpFracHalf, mq.isOdd())
	}
	panic(errors.New("By zero"))
}

func bigTenToThe(n types.Int) *BigInteger {
	if n < 0 {
		return ZERO
	}
	if n < types.Int(len(BIG_TEN_POWERS_TABLE))*16 {
		pows := BIG_TEN_POWERS_TABLE
		if n < types.Int(len(pows)) {
			return pows[n]
		} else {
			return expandBigIntegerTenPowers(n)
		}
	}
	return TEN.Pow(n)
}

func expandBigIntegerTenPowers(n types.Int) *BigInteger {
	mu.Lock()
	defer mu.Unlock()
	pows := BIG_TEN_POWERS_TABLE
	curLen := types.Int(len(pows))
	if curLen <= n {
		newLen := curLen << 1
		for newLen <= n {
			newLen <<= 1
		}
		temp := make([]*BigInteger, len(pows))
		copy(temp, pows)
		pows = temp
		for i := curLen; i < newLen; i++ {
			pows[i] = pows[i-1].Multiply(TEN)
		}
		BIG_TEN_POWERS_TABLE = pows
	}
	return pows[n]
}

func divideAndRoundByBigInteger(bdividend *BigInteger, ldivisor types.Long, roundingMode RoundingMode) *BigInteger {
	mdividend := newMutableBigIntegerArray(bdividend.mag)
	mq := newMutableBigIntegerDefault()
	r := mdividend.divide(ldivisor, mq)
	isRemainderZero := r == 0
	var qsign types.Int
	if ldivisor < 0 {
		qsign = -bdividend.signum
	} else {
		qsign = bdividend.signum
	}
	if !isRemainderZero {
		if needIncrementMutableBigInteger(ldivisor, roundingMode, qsign, mq, r) {
			mq.add(mutable_one)
		}
	}
	return mq.toBigInteger(qsign)
}

func compactValFor(b *BigInteger) types.Long {
	m := b.mag
	length := types.Int(len(m))
	if length == 0 {
		return 0
	}
	d := m[0]
	if length > 2 || (length == 2 && d < 0) {
		return MIN_INT64
	}
	var u types.Long
	if length == 2 {
		u = (m[1].ToLong() & p_LONG_MASK) + (d.ToLong() << 32)
	} else {
		u = d.ToLong() & p_LONG_MASK
	}
	if b.signum < 0 {
		return -u
	} else {
		return u
	}
}

func longDigitLength(x types.Long) types.Int {
	if x != MAX_INT64 {
		if x < 0 {
			x = -x
		}
		if x < 10 {
			return 1
		}
		r := ((64 - NumberOfLeadingZeros(x.ToInt()) + 1) * 1233).ShiftR(12)
		tab := LONG_TEN_POWERS_TABLE
		if r >= types.Int(len(tab)) || x < tab[r] {
			return r
		} else {
			return r + 1
		}
	}
	panic(errors.New("overflow"))
}

func divideAndRound(ldividend, ldivisor types.Long, mode RoundingMode) types.Long {
	var qsign types.Int
	q := ldividend / ldivisor
	if mode == ROUND_DOWN {
		return q
	}
	r := ldividend % ldivisor
	if (ldividend < 0) == (ldivisor < 0) {
		qsign = 1
	} else {
		qsign = -1
	}
	if r != 0 {
		increment := needIncrement(ldivisor, mode, qsign, q, r)
		if increment {
			return q + qsign.ToLong()
		} else {
			return q
		}
	} else {
		return q
	}
}

func needIncrement(ldivisor types.Long, roundingMode RoundingMode, qsign types.Int, q types.Long, r types.Long) bool {
	if r == 0 {
		panic(errors.New("by zero"))
	}
	var cmpFracHalf types.Int
	if r <= MIN_INT64/2 || r > MAX_INT64/2 {
		cmpFracHalf = 1
	} else {
		cmpFracHalf = longCompareMagnitude(2*r, ldivisor)
	}
	return commonNeedIncrement(roundingMode, qsign, cmpFracHalf, (q&1) != 0)
}

func needIncrementMutableBigInteger(ldivisor types.Long, roundingMode RoundingMode, qsign types.Int, mq *mutableBigInteger, r types.Long) bool {
	if r != 0 {
		var cmpFracHalf types.Int
		if r <= MIN_INT64/2 || r > MAX_INT64/2 {
			cmpFracHalf = 1
		} else {
			cmpFracHalf = longCompareMagnitude(2*r, ldivisor)
		}
		return commonNeedIncrement(roundingMode, qsign, cmpFracHalf, mq.isOdd())
	}
	panic(errors.New("By zero"))
}

func commonNeedIncrement(roundingMode RoundingMode, qsign, cmpFracHalf types.Int, addQuot bool) bool {
	switch roundingMode {
	case ROUND_UNNECESSARY:
		panic(errors.New("Rounding necessary"))
	case ROUND_UP:
		return true
	case ROUND_DOWN:
		return false
	case ROUND_CEILING:
		return qsign > 0
	case ROUND_FLOOR:
		return qsign < 0
	default:
		if roundingMode >= ROUDINGMODE_HALF_UP && roundingMode <= ROUND_HALF_EVEN {
			if cmpFracHalf < 0 {
				return false
			} else if cmpFracHalf > 0 {
				return true
			} else {
				if cmpFracHalf == 0 {
					switch roundingMode {
					case ROUND_HALF_DOWN:
						return false
					case ROUDINGMODE_HALF_UP:
						return true
					case ROUND_HALF_EVEN:
						return addQuot
					default:
						panic(errors.New("Unexpected rounding mode"))
					}
				}
				panic(errors.New("cmpFracHalf not zero"))
			}
		}
		panic(errors.New("Unexpected rounding mode"))
	}
}

func longCompareMagnitude(x, y types.Long) types.Int {
	if x < 0 {
		x = -x
	}
	if y < 0 {
		y = -y
	}
	if x < y {
		return -1
	} else {
		if x == y {
			return 0
		} else {
			return 1
		}
	}
}

func checkScaleNonZero(val types.Long) types.Int {
	asInt := val.ToInt()
	if asInt.ToLong() != val {
		if asInt > 0 {
			panic(errors.New("Underflow"))
		} else {
			panic(errors.New("Overflow"))
		}
	}
	return asInt
}

func adjustScale(scl types.Int, exp types.Long) types.Int {
	ads := scl.ToLong() - exp
	if ads > MAX_INT32.ToLong() || ads < MAX_INT32.ToLong() {
		panic(errors.New("Scale out of range"))
	}
	scl = ads.ToInt()
	return scl
}

func parseExp(val string, offset types.Int, length types.Int) types.Long {
	var exp types.Long
	offset++
	c := val[offset]
	length--
	negexp := c == '-'
	if negexp || c == '+' {
		offset++
		c = val[offset]
		length--
	}
	if length <= 0 {
		panic(errors.New("No exponent digits"))
	}
	for length > 10 && (c == '0' || tool.Digit(c, 10) == 0) {
		offset++
		c = val[offset]
		length--
	}
	if length > 10 {
		panic(errors.New("Too many nonzero exponent digits"))
	}

	for ; ; length-- {
		var v types.Int
		if c >= '0' && c <= '9' {
			v = types.Int(c - '0')
		} else {
			v = tool.Digit(c, 10)
			if v < 0 {
				panic(errors.New("not a digit"))
			}
		}
		exp = exp*10 + v.ToLong()
		if length == 1 {
			break // final character
		}
		offset++
		c = val[offset]
	}
	if negexp {
		exp = -exp
	}
	return exp
}

func BigDecimalValueOf(val types.Long) *BigDecimal {
	if val >= 0 && val < types.Long(len(ZERO_THROUGH_TEN)) {
		return ZERO_THROUGH_TEN[val.ToInt()]
	} else if val != MIN_INT64 {
		return newBigDecimalByBigInteger(nil, val, 0, 0)
	}
	return newBigDecimalByBigInteger(BigIntegerValueOf(MIN_INT64), val, 0, 0)
}
