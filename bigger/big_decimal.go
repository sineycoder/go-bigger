package bigger

import (
	"errors"
	"github.com/sineycoder/go-bigger/tool"
	"github.com/sineycoder/go-bigger/types"
	"math"
	"sync"
)

/**
 @author: nizhenxian
 @date: 2021/8/14 11:54:01
**/

type RoundingMode types.Int

var mu sync.Mutex

type bigDecimal struct {
	intVal      *bigInteger
	scale       types.Int
	precision   types.Int
	intCompact  types.Long
	stringCache string
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
	ROUND_HALF_UP
	ROUND_HALF_DOWN
	ROUND_HALF_EVEN
	ROUND_UNNECESSARY
)

var (
	p_DIV_NUM_BASE     = types.Long(1) << 32
	p_ZERO_SCALED_BY   []*bigDecimal
	p_ZERO_THROUGH_TEN []*bigDecimal
	p_THRESHOLDS_TABLE = []types.Long{
		MAX_INT64,                       // 0
		MAX_INT64 / 10,                  // 1
		MAX_INT64 / 100,                 // 2
		MAX_INT64 / 1000,                // 3
		MAX_INT64 / 10000,               // 4
		MAX_INT64 / 100000,              // 5
		MAX_INT64 / 1000000,             // 6
		MAX_INT64 / 10000000,            // 7
		MAX_INT64 / 100000000,           // 8
		MAX_INT64 / 1000000000,          // 9
		MAX_INT64 / 10000000000,         // 10
		MAX_INT64 / 100000000000,        // 11
		MAX_INT64 / 1000000000000,       // 12
		MAX_INT64 / 10000000000000,      // 13
		MAX_INT64 / 100000000000000,     // 14
		MAX_INT64 / 1000000000000000,    // 15
		MAX_INT64 / 10000000000000000,   // 16
		MAX_INT64 / 100000000000000000,  // 17
		MAX_INT64 / 1000000000000000000, // 18
	}
	p_LONG_TEN_POWERS_TABLE = []types.Long{
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
	p_BIG_TEN_POWERS_TABLE []*bigInteger
)

func init() {
	Init()
	p_ZERO_THROUGH_TEN = []*bigDecimal{
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
	p_ZERO_SCALED_BY = []*bigDecimal{
		p_ZERO_THROUGH_TEN[0],
		newBigDecimalByBigInteger(ZERO, 0, 1, 1),
		newBigDecimalByBigInteger(ZERO, 0, 2, 1),
		newBigDecimalByBigInteger(ZERO, 0, 3, 1),
		newBigDecimalByBigInteger(ZERO, 0, 4, 1),
		newBigDecimalByBigInteger(ZERO, 0, 5, 1),
		newBigDecimalByBigInteger(ZERO, 0, 6, 1),
		newBigDecimalByBigInteger(ZERO, 0, 7, 1),
		newBigDecimalByBigInteger(ZERO, 0, 8, 1),
		newBigDecimalByBigInteger(ZERO, 0, 9, 1),
		newBigDecimalByBigInteger(ZERO, 0, 10, 1),
		newBigDecimalByBigInteger(ZERO, 0, 11, 1),
		newBigDecimalByBigInteger(ZERO, 0, 12, 1),
		newBigDecimalByBigInteger(ZERO, 0, 13, 1),
		newBigDecimalByBigInteger(ZERO, 0, 14, 1),
		newBigDecimalByBigInteger(ZERO, 0, 15, 1),
	}
	p_BIG_TEN_POWERS_TABLE = []*bigInteger{
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
}

func newBigDecimalByBigInteger(intVal *bigInteger, val types.Long, scale, prec types.Int) *bigDecimal {
	return &bigDecimal{
		scale:      scale,
		precision:  prec,
		intCompact: val,
		intVal:     intVal,
	}
}

func newBigDecimalByBigInteger2(unscaledVal *bigInteger, scale types.Int) *bigDecimal {
	return &bigDecimal{
		scale:      scale,
		intCompact: compactValFor(unscaledVal),
		intVal:     unscaledVal,
	}
}

func NewBigDecimalString(val string) *bigDecimal {
	if val == "" {
		panic("illegal value")
	}
	var offset, length, prec, scl types.Int
	var rs types.Long
	var rb *bigInteger
	var mc = &mathContext{roundingMode: ROUND_HALF_UP}
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
				rs = divideAndRound(rs, p_LONG_TEN_POWERS_TABLE[drop], mc.roundingMode)
				prec = longDigitLength(rs)
				drop = prec - mcp
			}
		}
	} else {
		coeff := make([]uint8, length)
		for ; length > 0; offset++ {
			c = val[offset]
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
				length--
				continue
			}
			if c == '.' {
				if dot {
					panic(errors.New("Character array contains more than one point"))
				}
				dot = true
				length--
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
					rs = divideAndRound(rs, p_LONG_TEN_POWERS_TABLE[drop], mc.roundingMode)
					prec = longDigitLength(rs)
					drop = prec - mcp
				}
				rb = nil
			}
		}
	}
	return &bigDecimal{
		scale:      scl,
		precision:  prec,
		intCompact: rs,
		intVal:     rb,
	}
}

func bigDigitLength(b *bigInteger) types.Int {
	if b.signum == 0 {
		return 1
	}
	r := ((b.BitLength() + 1) * 646456993).ShiftR(31)
	if b.compareMagnitute(bigTenToThe(r)) < 0 {
		return r
	}
	return r + 1
}

func divideAndRoundByTenPow(intVal *bigInteger, tenPow types.Int, roundingMode RoundingMode) *bigInteger {
	if tenPow < types.Int(len(p_LONG_TEN_POWERS_TABLE)) {
		intVal = divideAndRoundByBigInteger(intVal, p_LONG_TEN_POWERS_TABLE[tenPow], roundingMode)
	} else {
		intVal = divideAndRoundByBigInteger2(intVal, bigTenToThe(tenPow), roundingMode)
	}
	return intVal
}

func divideAndRoundByBigInteger2(bdividend *bigInteger, bdivisor *bigInteger, roundingMode RoundingMode) *bigInteger {
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
	panic(errors.New("illegal param"))
}

func bigTenToThe(n types.Int) *bigInteger {
	if n < 0 {
		return ZERO
	}
	if n < types.Int(len(p_BIG_TEN_POWERS_TABLE))*16 {
		pows := p_BIG_TEN_POWERS_TABLE
		if n < types.Int(len(pows)) {
			return pows[n]
		} else {
			return expandBigIntegerTenPowers(n)
		}
	}
	return TEN.Pow(n)
}

func expandBigIntegerTenPowers(n types.Int) *bigInteger {
	mu.Lock()
	defer mu.Unlock()
	pows := p_BIG_TEN_POWERS_TABLE
	curLen := types.Int(len(pows))
	if curLen <= n {
		newLen := curLen << 1
		for newLen <= n {
			newLen <<= 1
		}
		temp := make([]*bigInteger, newLen)
		copy(temp, pows)
		pows = temp
		for i := curLen; i < newLen; i++ {
			pows[i] = pows[i-1].Multiply(TEN)
		}
		p_BIG_TEN_POWERS_TABLE = pows
	}
	return pows[n]
}

func divideAndRoundByBigInteger(bdividend *bigInteger, ldivisor types.Long, roundingMode RoundingMode) *bigInteger {
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

func compactValFor(b *bigInteger) types.Long {
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
		tab := p_LONG_TEN_POWERS_TABLE
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
		panic(errors.New("illegal param"))
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
	panic(errors.New("illegal param"))
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
		if roundingMode >= ROUND_HALF_UP && roundingMode <= ROUND_HALF_EVEN {
			if cmpFracHalf < 0 {
				return false
			} else if cmpFracHalf > 0 {
				return true
			} else {
				if cmpFracHalf == 0 {
					switch roundingMode {
					case ROUND_HALF_DOWN:
						return false
					case ROUND_HALF_UP:
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
	if ads > MAX_INT32.ToLong() || ads < MIN_INT32.ToLong() {
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

func BigDecimalValueOf(val types.Long) *bigDecimal {
	if val >= 0 && val < types.Long(len(p_ZERO_THROUGH_TEN)) {
		return p_ZERO_THROUGH_TEN[val.ToInt()]
	} else if val != MIN_INT64 {
		return newBigDecimalByBigInteger(nil, val, 0, 0)
	}
	return newBigDecimalByBigInteger(BigIntegerValueOf(MIN_INT64), val, 0, 0)
}

func (b *bigDecimal) String() string {
	sc := b.stringCache
	if sc == "" {
		b.stringCache = b.layoutChars(true)
		sc = b.stringCache
	}
	return sc
}

func (b *bigDecimal) layoutChars(sci bool) string {
	if b.scale == 0 {
		if b.intCompact != MIN_INT64 {
			str := b.intCompact.String()
			return str
		} else {
			str := b.intVal.String()
			return str
		}
	}
	if b.scale == 2 && b.intCompact >= 0 && b.intCompact < MAX_INT32.ToLong() {
		lowInt := b.intCompact.ToInt() % 100
		highInt := b.intCompact.ToInt() / 100
		str := highInt.String() + "." + string(p_DIGIT_TENS[lowInt]) + string(p_DIGIT_ONES[lowInt])
		return str
	}

	sbHelper := newStringBuilderHelper()
	var coeff []rune
	var offset types.Int
	if b.intCompact != MIN_INT64 {
		offset = sbHelper.putIntCompact(types.Long(math.Abs(float64(b.intCompact))))
		coeff = sbHelper.getCompactCharArray()
	} else {
		offset = 0
		coeff = []rune(b.intVal.Abs().String())
	}

	buf := sbHelper.getBuffer()
	if b.signum() < 0 {
		buf = append(buf, '-')
	}
	coeffLen := types.Int(len(coeff)) - offset
	adjusted := -b.scale.ToLong() + (coeffLen - 1).ToLong()
	if (b.scale >= 0) && (adjusted >= -6) {
		pad := b.scale - coeffLen
		if pad >= 0 {
			buf = append(buf, '0')
			buf = append(buf, '.')
			for ; pad > 0; pad-- {
				buf = append(buf, '0')
			}
			buf = append(buf, coeff[offset:offset+coeffLen]...)
		} else {
			buf = append(buf, coeff[offset:offset-pad]...)
			buf = append(buf, '.')
			buf = append(buf, coeff[-pad+offset:-pad+offset+b.scale]...)
		}
	} else {
		if sci {
			buf = append(buf, coeff[offset])
			if coeffLen > 1 {
				buf = append(buf, '.')
				buf = append(buf, coeff[offset+1:offset+1+coeffLen-1]...)
			}
		} else {
			sig := (adjusted % 3).ToInt()
			if sig < 0 {
				sig += 3
			}
			adjusted -= sig.ToLong()
			sig++
			if b.signum() == 0 {
				switch sig {
				case 1:
					buf = append(buf, '0')
				case 2:
					buf = append(buf, []rune("0.00")...)
					adjusted += 3
				case 3:
					buf = append(buf, []rune("0.0")...)
					adjusted += 3
				default:
					panic(errors.New("Unexpected sig value"))
				}
			} else if sig >= coeffLen {
				buf = append(buf, coeff[offset:offset+coeffLen]...)
				for i := sig - coeffLen; i > 0; i-- {
					buf = append(buf, '0')
				}
			} else {
				buf = append(buf, coeff[offset:offset+sig]...)
				buf = append(buf, '.')
				buf = append(buf, coeff[offset+sig:offset+coeffLen]...)
			}
		}
		if adjusted != 0 {
			buf = append(buf, 'E')
			if adjusted > 0 {
				buf = append(buf, '+')
			}
			buf = append(buf, []rune(adjusted.String())...)
		}
	}
	return string(buf)
}

func (b *bigDecimal) signum() types.Int {
	if b.intCompact != MIN_INT64 {
		return ((b.intCompact >> 63) | (-b.intCompact).ShiftR(63)).ToInt()
	} else {
		return b.intVal.signum
	}
}

func (b *bigDecimal) Add(augend *bigDecimal) *bigDecimal {
	if b.intCompact != MIN_INT64 {
		if augend.intCompact != MIN_INT64 {
			return b.add(b.intCompact, b.scale, augend.intCompact, augend.scale)
		}
	}
	return nil
}

func (b *bigDecimal) add(xs types.Long, scale1 types.Int, ys types.Long, scale2 types.Int) *bigDecimal {
	sdiff := scale1.ToLong() - scale2.ToLong()
	if sdiff == 0 {
		return add3(xs, ys, scale1)
	} else if sdiff < 0 {
		raise := checkScale(xs, -sdiff)
		scaledX := longMultiplPowerTen(xs, raise)
		if scaledX != MIN_INT64 {
			return add3(scaledX, ys, scale2)
		} else {
			bigsum := bigMultiplyPowerTen(xs, raise)
			if (xs ^ ys) >= 0 {
				return newBigDecimalByBigInteger(bigsum, MIN_INT64, scale2, 0)
			} else {
				return valueOf3(bigsum, scale2, 0)
			}
		}
	} else {
		raise := checkScale(ys, sdiff)
		scaledY := longMultiplPowerTen(ys, raise)
		if scaledY != MIN_INT64 {
			return add3(xs, scaledY, scale1)
		} else {
			bigsum := bigMultiplyPowerTen(ys, raise).add(xs)
			if (xs ^ ys) >= 0 {
				return newBigDecimalByBigInteger(bigsum, MIN_INT64, scale1, 0)
			} else {
				return valueOf3(bigsum, scale1, 0)
			}
		}
	}
}

func (b *bigDecimal) SetScale(newScale types.Int, roundingMode RoundingMode) *bigDecimal {
	if roundingMode < ROUND_UP || roundingMode > ROUND_UNNECESSARY {
		panic(errors.New("Invalid rounding mode"))
	}

	oldScale := b.scale
	if newScale == oldScale {
		return b
	}
	if b.signum() == 0 {
		return zeroValueOf(newScale)
	}
	if b.intCompact != MIN_INT64 {
		rs := b.intCompact
		if newScale > oldScale {
			raise := b.checkScale(newScale.ToLong() - oldScale.ToLong())
			rs = longMultiplPowerTen(rs, raise)
			if rs != MIN_INT64 {
				return valueOf(rs, newScale)
			}
			rb := b.bigMultiplyPowerTen(raise)
			if b.precision > 0 {
				return newBigDecimalByBigInteger(rb, MIN_INT64, newScale, b.precision+raise)
			} else {
				return newBigDecimalByBigInteger(rb, MIN_INT64, newScale, 0)
			}
		} else {
			drop := b.checkScale(oldScale.ToLong() - newScale.ToLong())
			if drop < types.Int(len(p_LONG_TEN_POWERS_TABLE)) {
				return divideAndRound5(rs, p_LONG_TEN_POWERS_TABLE[drop], newScale, roundingMode, newScale)
			} else {
				return divideAndRoundByBigInteger5(b.inflated(), bigTenToThe(drop), newScale, roundingMode, newScale)
			}
		}
	} else {
		if newScale > oldScale {
			raise := b.checkScale(newScale.ToLong() - oldScale.ToLong())
			rb := bigMultiplyPowerTenByBigInteger(b.intVal, raise)
			if b.precision > 0 {
				return newBigDecimalByBigInteger(rb, MIN_INT64, newScale, b.precision+raise)
			} else {
				return newBigDecimalByBigInteger(rb, MIN_INT64, newScale, 0)
			}
		} else {
			drop := b.checkScale(oldScale.ToLong() - newScale.ToLong())
			if drop < types.Int(len(p_LONG_TEN_POWERS_TABLE)) {
				return divideAndRoundHalfByBigInteger5(b.intVal, p_LONG_TEN_POWERS_TABLE[drop], newScale, roundingMode, newScale)
			} else {
				return divideAndRoundByBigInteger5(b.intVal, bigTenToThe(drop), newScale, roundingMode, newScale)
			}
		}
	}
}

func divideAndRoundHalfByBigInteger5(bdividend *bigInteger, ldivisor types.Long, scale types.Int, roundingMode RoundingMode, preferredScale types.Int) *bigDecimal {
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
		if needIncrementByMutableBigInteger(ldivisor, roundingMode, qsign, mq, r) {
			mq.add(mutable_one)
		}
		return mq.toBigDecimal(qsign, scale)
	} else {
		if preferredScale != scale {
			compactVal := mq.toCompactValue(qsign)
			if compactVal != MIN_INT64 {
				return createAndStripZerosToMatchScale(compactVal, scale, preferredScale)
			}
			intVal := mq.toBigInteger(qsign)
			return createAndStripZerosToMatchScaleByBigInteger(intVal, scale, preferredScale)
		} else {
			return mq.toBigDecimal(qsign, scale)
		}
	}
}

func needIncrementByMutableBigInteger(ldivisor types.Long, roundingMode RoundingMode, qsign types.Int, mq *mutableBigInteger, r types.Long) bool {
	var cmpFracHalf types.Int
	if r <= MIN_INT64/2 || r > MAX_INT64/2 {
		cmpFracHalf = 1
	} else {
		cmpFracHalf = longCompareMagnitude(2*r, ldivisor)
	}
	return commonNeedIncrement(roundingMode, qsign, cmpFracHalf, mq.isOdd())
}

func bigMultiplyPowerTenByBigInteger(value *bigInteger, n types.Int) *bigInteger {
	if n <= 0 {
		return value
	}
	if n < types.Int(len(p_LONG_TEN_POWERS_TABLE)) {
		return value.multiplyLong(p_LONG_TEN_POWERS_TABLE[n])
	}
	return value.Multiply(bigTenToThe(n))
}

func divideAndRoundByBigInteger5(bdividend *bigInteger, bdivisor *bigInteger, scale types.Int, roundingMode RoundingMode, preferredScale types.Int) *bigDecimal {
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
		return mq.toBigDecimal(qsign, scale)
	} else {
		if preferredScale != scale {
			compactVal := mq.toCompactValue(qsign)
			if compactVal != MIN_INT64 {
				return createAndStripZerosToMatchScale(compactVal, scale, preferredScale)
			}
			intVal := mq.toBigInteger(qsign)
			return createAndStripZerosToMatchScaleByBigInteger(intVal, scale, preferredScale)
		} else {
			return mq.toBigDecimal(qsign, scale)
		}
	}
}

func divideAndRound5(ldividend types.Long, ldivisor types.Long, scale types.Int, roundingMode RoundingMode, preferredScale types.Int) *bigDecimal {
	var qsign types.Int
	q := ldividend / ldivisor
	if roundingMode == ROUND_DOWN && scale == preferredScale {
		return valueOf(q, scale)
	}
	r := ldividend % ldivisor
	if (ldividend < 0) == (ldivisor < 0) {
		qsign = 1
	} else {
		qsign = -1
	}
	if r != 0 {
		increment := needIncrement(ldivisor, roundingMode, qsign, q, r)
		if increment {
			return valueOf(q+qsign.ToLong(), scale)
		} else {
			return valueOf(q, scale)
		}
	} else {
		if preferredScale != scale {
			return createAndStripZerosToMatchScale(q, scale, preferredScale)
		} else {
			return valueOf(q, scale)
		}
	}
}

func createAndStripZerosToMatchScaleByBigInteger(intVal *bigInteger, scale types.Int, preferredScale types.Int) *bigDecimal {
	var qr []*bigInteger
	for intVal.compareMagnitute(TEN) >= 0 && scale > preferredScale {
		if intVal.testBit(0) {
			break
		}
		qr = intVal.DivideAndRemainder(TEN)
		if qr[1].signum != 0 {
			break
		}
		intVal = qr[0]
		scale = checkScaleByBigInteger(intVal, scale.ToLong()-1)
	}
	return valueOf3(intVal, scale, 0)
}

func checkScaleByBigInteger(intVal *bigInteger, val types.Long) types.Int {
	asInt := val.ToInt()
	if asInt.ToLong() != val {
		if val > MAX_INT32.ToLong() {
			asInt = MAX_INT32
		} else {
			asInt = MIN_INT32
		}
		if intVal.signum != 0 {
			if asInt > 0 {
				panic(errors.New("Underflow"))
			} else {
				panic(errors.New("Overflow"))
			}
		}
	}
	return asInt
}

func createAndStripZerosToMatchScale(compactVal types.Long, scale types.Int, preferredScale types.Int) *bigDecimal {
	for compactVal.Abs() >= 10 && scale > preferredScale {
		if (compactVal & 1) != 0 {
			break
		}
		r := compactVal % 10
		if r != 0 {
			break
		}
		compactVal /= 10
		scale = checkScale(compactVal, scale.ToLong()-1)
	}
	return valueOf(compactVal, scale)
}

func bigMultiplyPowerTen(value types.Long, n types.Int) *bigInteger {
	if n <= 0 {
		return BigIntegerValueOf(value)
	}
	return bigTenToThe(n).multiplyLong(value)
}

func longMultiplPowerTen(val types.Long, n types.Int) types.Long {
	if val == 0 || n <= 0 {
		return val
	}
	tab := p_LONG_TEN_POWERS_TABLE
	bounds := p_THRESHOLDS_TABLE
	if n < types.Int(len(tab)) && n < types.Int(len(bounds)) {
		tenpower := tab[n]
		if val == 1 {
			return tenpower
		}
		if val.Abs() <= bounds[n] {
			return val * tenpower
		}
	}
	return MIN_INT64
}

func (b *bigDecimal) checkScale(val types.Long) types.Int {
	asInt := val.ToInt()
	if asInt.ToLong() != val {
		if val > MAX_INT32.ToLong() {
			asInt = MAX_INT32
		} else {
			asInt = MIN_INT32
		}
		big := b.intVal
		if b.intCompact != 0 && (big == nil || big.signum != 0) {
			if asInt > 0 {
				panic(errors.New("Underflow"))
			} else {
				panic(errors.New("Overflow"))
			}
		}
	}
	return asInt
}

func (b *bigDecimal) bigMultiplyPowerTen(n types.Int) *bigInteger {
	if n <= 0 {
		return b.inflated()
	}

	if b.intCompact != MIN_INT64 {
		return bigTenToThe(n).multiplyLong(b.intCompact)
	} else {
		return b.intVal.Multiply(bigTenToThe(n))
	}
}

func (b *bigDecimal) inflated() *bigInteger {
	if b.intVal == nil {
		return BigIntegerValueOf(b.intCompact)
	}
	return b.intVal
}

func (b *bigDecimal) Subtract(subtrahend *bigDecimal) *bigDecimal {
	if b.intCompact != MIN_INT64 {
		if subtrahend.intCompact != MIN_INT64 {
			return add4(b.intCompact, b.scale, -subtrahend.intCompact, subtrahend.scale)
		} else {
			return add4_(b.intCompact, b.scale, subtrahend.intVal.negate(), subtrahend.scale)
		}
	} else {
		if subtrahend.intCompact != MIN_INT64 {
			return add4_(-subtrahend.intCompact, subtrahend.scale, b.intVal, b.scale)
		} else {
			return add4__(b.intVal, b.scale, subtrahend.intVal.negate(), subtrahend.scale)
		}
	}
}

func (b *bigDecimal) Multiply(multiplicand *bigDecimal) *bigDecimal {
	productScale := b.checkScale(b.scale.ToLong() + multiplicand.scale.ToLong())
	if b.intCompact != MIN_INT64 {
		if multiplicand.intCompact != MIN_INT64 {
			return multiply3(b.intCompact, multiplicand.intCompact, productScale)
		} else {
			return multiply3_(b.intCompact, multiplicand.intVal, productScale)
		}
	} else {
		if multiplicand.intCompact != MIN_INT64 {
			return multiply3_(multiplicand.intCompact, b.intVal, productScale)
		} else {
			return multiply3__(b.intVal, multiplicand.intVal, productScale)
		}
	}
}

func (b *bigDecimal) Divide(divisor *bigDecimal, scale types.Int, roundingMode RoundingMode) *bigDecimal {
	if roundingMode < ROUND_UP || roundingMode > ROUND_UNNECESSARY {
		panic(errors.New("invalid rounding mode"))
	}
	if b.intCompact != MIN_INT64 {
		if divisor.intCompact != MIN_INT64 {
			return divide6(b.intCompact, b.scale, divisor.intCompact, divisor.scale, scale, roundingMode)
		} else {
			return divide6_(b.intCompact, b.scale, divisor.intVal, divisor.scale, scale, roundingMode)
		}
	} else {
		if divisor.intCompact != MIN_INT64 {
			return divide6__(b.intVal, b.scale, divisor.intCompact, divisor.scale, scale, roundingMode)
		} else {
			return divide6___(b.intVal, b.scale, divisor.intVal, divisor.scale, scale, roundingMode)
		}
	}
}

func divide6___(dividend *bigInteger, dividendScale types.Int, divisor *bigInteger, divisorScale types.Int, scale types.Int, roundingMode RoundingMode) *bigDecimal {
	if checkScaleByBigInteger(dividend, scale.ToLong()+divisorScale.ToLong()) > dividendScale {
		newScale := scale + divisorScale
		raise := newScale - dividendScale
		scaledDividend := bigMultiplyPowerTenByBigInteger(dividend, raise)
		return divideAndRoundByBigInteger5(scaledDividend, divisor, scale, roundingMode, scale)
	} else {
		newScale := checkScaleByBigInteger(divisor, dividendScale.ToLong()-scale.ToLong())
		raise := newScale - divisorScale
		scaledDivisor := bigMultiplyPowerTenByBigInteger(divisor, raise)
		return divideAndRoundByBigInteger5(dividend, scaledDivisor, scale, roundingMode, scale)
	}
}

func divide6__(dividend *bigInteger, dividendScale types.Int, divisor types.Long, divisorScale types.Int, scale types.Int, roundingMode RoundingMode) *bigDecimal {
	if checkScaleByBigInteger(dividend, scale.ToLong()+divisorScale.ToLong()) > dividendScale {
		newScale := scale + divisorScale
		raise := newScale - dividendScale
		scaledDividend := bigMultiplyPowerTenByBigInteger(dividend, raise)
		return divideAndRoundHalfByBigInteger5(scaledDividend, divisor, scale, roundingMode, scale)
	} else {
		newscale := checkScale(divisor, dividendScale.ToLong()-scale.ToLong())
		raise := newscale - divisorScale
		if raise < types.Int(len(p_LONG_TEN_POWERS_TABLE)) {
			ys := divisor
			ys = longMultiplPowerTen(ys, raise)
			if ys != MIN_INT64 {
				return divideAndRoundHalfByBigInteger5(dividend, ys, scale, roundingMode, scale)
			}
		}
		scaledDivisor := bigMultiplyPowerTen(divisor, raise)
		return divideAndRoundByBigInteger5(dividend, scaledDivisor, scale, roundingMode, scale)
	}
}

func divide6_(dividend types.Long, dividendScale types.Int, divisor *bigInteger, divisorScale types.Int, scale types.Int, roundingMode RoundingMode) *bigDecimal {
	if checkScale(dividend, scale.ToLong()+divisorScale.ToLong()) > dividendScale {
		newScale := scale + divisorScale
		raise := newScale - dividendScale
		scaledDividend := bigMultiplyPowerTen(dividend, raise)
		return divideAndRoundByBigInteger5(scaledDividend, divisor, scale, roundingMode, scale)
	} else {
		newScale := checkScaleByBigInteger(divisor, dividendScale.ToLong()-scale.ToLong())
		raise := newScale - divisorScale
		scaledDivisor := bigMultiplyPowerTenByBigInteger(divisor, raise)
		return divideAndRoundByBigInteger5(BigIntegerValueOf(dividend), scaledDivisor, scale, roundingMode, scale)
	}
}

func divide6(dividend types.Long, dividendScale types.Int, divisor types.Long, divisorScale types.Int, scale types.Int, roundingMode RoundingMode) *bigDecimal {
	if checkScale(dividend, scale.ToLong()+divisorScale.ToLong()) > dividendScale {
		newScale := scale + divisorScale
		raise := newScale - dividendScale
		if raise < types.Int(len(p_LONG_TEN_POWERS_TABLE)) {
			xs := dividend
			xs = longMultiplPowerTen(xs, raise)
			if xs != MIN_INT64 {
				return divideAndRound5(xs, divisor, scale, roundingMode, scale)
			}
			q := multiplyDivideAndRound(p_LONG_TEN_POWERS_TABLE[raise], dividend, divisor, scale, roundingMode, scale)
			if q != nil {
				return q
			}
		}
		scaledDividend := bigMultiplyPowerTen(dividend, raise)
		return divideAndRoundHalfByBigInteger5(scaledDividend, divisor, scale, roundingMode, scale)
	} else {
		newScale := checkScale(divisor, divisorScale.ToLong()-scale.ToLong())
		raise := newScale - divisorScale
		if raise < types.Int(len(p_LONG_TEN_POWERS_TABLE)) {
			ys := divisor
			ys = longMultiplPowerTen(ys, raise)
			if ys != MIN_INT64 {
				return divideAndRound5(dividend, ys, scale, roundingMode, scale)
			}
		}
		scaledDivisor := bigMultiplyPowerTen(divisor, raise)
		return divideAndRoundByBigInteger5(BigIntegerValueOf(dividend), scaledDivisor, scale, roundingMode, scale)
	}
}

func multiplyDivideAndRound(dividend0 types.Long, dividend1 types.Long, divisor types.Long, scale types.Int, roundingMode RoundingMode, preferredScale types.Int) *bigDecimal {
	qsign := dividend0.Signum() * dividend1.Signum() * divisor.Signum()
	dividend0 = dividend0.Abs()
	dividend1 = dividend1.Abs()
	divisor = divisor.Abs()

	d0_hi := dividend0.ShiftR(32)
	d0_lo := dividend0 & p_LONG_MASK
	d1_hi := dividend1.ShiftR(32)
	d1_lo := dividend1 & p_LONG_MASK
	product := d0_lo * d1_lo
	d0 := product & p_LONG_MASK
	d1 := product.ShiftR(32)
	product = d0_hi*d1_lo + d1
	d1 = product & p_LONG_MASK
	d2 := product.ShiftR(32)
	product = d0_lo*d1_hi + d1
	d1 = product & p_LONG_MASK
	d2 += product.ShiftR(32)
	d3 := d2.ShiftR(32)
	d2 &= p_LONG_MASK
	product = d0_hi*d1_hi + d2
	d2 = product & p_LONG_MASK
	d3 = (product.ShiftR(32) + d3) & p_LONG_MASK
	dividendHi := make64(d3, d2)
	dividendLo := make64(d1, d0)
	//divide
	return divideAndRound128(dividendHi, dividendLo, divisor, qsign, scale, roundingMode, preferredScale)
}

func divideAndRound128(dividendHi types.Long, dividendLo types.Long, divisor types.Long, sign types.Int, scale types.Int, roundingMode RoundingMode, preferredScale types.Int) *bigDecimal {
	if dividendHi >= divisor {
		return nil
	}

	shift := NumberOfLeadingZerosForLong(divisor)
	divisor <<= shift

	v1 := divisor.ShiftR(32)
	v0 := divisor & p_LONG_MASK

	var tmp types.Long
	tmp = dividendLo << shift
	u1 := tmp.ShiftR(32)
	u0 := tmp & p_LONG_MASK

	tmp = (dividendHi << shift) | (dividendLo.ShiftR(64 - shift))
	u2 := tmp & p_LONG_MASK
	var q1, r_tmp types.Long
	if v1 == 1 {
		q1 = tmp
		r_tmp = 0
	} else if tmp >= 0 {
		q1 = tmp / v1
		r_tmp = tmp - q1*v1
	} else {
		rq := divRemNegativeLong(tmp, v1)
		q1 = rq[1]
		r_tmp = rq[0]
	}

	for q1 >= p_DIV_NUM_BASE || unsignedLongCompare(q1*v0, make64(r_tmp, u1)) {
		q1--
		r_tmp += v1
		if r_tmp >= p_DIV_NUM_BASE {
			break
		}
	}

	tmp = mulsub(u2, u1, v1, v0, q1)
	u1 = tmp & p_LONG_MASK
	var q0 types.Long
	if v1 == 1 {
		q0 = tmp
		r_tmp = 0
	} else if tmp >= 0 {
		q0 = tmp / v1
		r_tmp = tmp - q0*v1
	} else {
		rq := divRemNegativeLong(tmp, v1)
		q0 = rq[1]
		r_tmp = rq[0]
	}

	for q0 >= p_DIV_NUM_BASE || unsignedLongCompare(q0*v0, make64(r_tmp, u0)) {
		q0--
		r_tmp += v1
		if r_tmp >= p_DIV_NUM_BASE {
			break
		}
	}

	if q1.ToInt() < 0 {
		mq := newMutableBigIntegerArray([]types.Int{q1.ToInt(), q0.ToInt()})
		if roundingMode == ROUND_DOWN && scale == preferredScale {
			return mq.toBigDecimal(sign, scale)
		}
		r := mulsub(u1, u0, v1, v0, q0).ShiftR(shift)
		if r != 0 {
			if needIncrementMutableBigInteger(divisor.ShiftR(shift), roundingMode, sign, mq, r) {
				mq.add(mutable_one)
			}
			return mq.toBigDecimal(sign, scale)
		} else {
		}
		if preferredScale != scale {
			intVal := mq.toBigInteger(sign)
			return createAndStripZerosToMatchScaleByBigInteger(intVal, scale, preferredScale)
		} else {
			return mq.toBigDecimal(sign, scale)
		}
	}

	q := make64(q1, q0)
	q *= sign.ToLong()
	if roundingMode == ROUND_DOWN && scale == preferredScale {
		return valueOf(q, scale)
	}

	r := mulsub(u1, u0, v1, v0, q0).ShiftR(shift)
	if r != 0 {
		incr := needIncrement(divisor.ShiftR(shift), roundingMode, sign, q, r)
		if incr {
			return valueOf(q+sign.ToLong(), scale)
		} else {
			return valueOf(q, scale)
		}
	} else {
		if preferredScale != scale {
			return createAndStripZerosToMatchScale(q, scale, preferredScale)
		} else {
			return valueOf(q, scale)
		}
	}
}

func mulsub(u1 types.Long, u0 types.Long, v1 types.Long, v0 types.Long, q0 types.Long) types.Long {
	tmp := u0 - q0*v0
	return make64(u1+tmp.ShiftR(32)-q0*v1, tmp&p_LONG_MASK)
}

func divRemNegativeLong(n types.Long, d types.Long) []types.Long {
	if n >= 0 {
		panic(errors.New("Non-negative numberator"))
	}
	if d == 1 {
		panic(errors.New("Unity denominator"))
	}
	q := n.ShiftR(1) / d.ShiftR(1)
	r := n - q*d
	for r < 0 {
		r += d
		q--
	}
	for r >= d {
		r -= d
		q++
	}
	return []types.Long{r, q}
}

func make64(hi types.Long, lo types.Long) types.Long {
	return hi<<32 | lo
}

func multiply3__(x, y *bigInteger, scale types.Int) *bigDecimal {
	return newBigDecimalByBigInteger(x.Multiply(y), MIN_INT64, scale, 0)
}

func multiply3_(x types.Long, y *bigInteger, scale types.Int) *bigDecimal {
	if x == 0 {
		return zeroValueOf(scale)
	}
	return newBigDecimalByBigInteger(y.multiplyLong(x), MIN_INT64, scale, 0)
}

func multiply3(x, y types.Long, scale types.Int) *bigDecimal {
	product := multiply2(x, y)
	if product != MIN_INT64 {
		return valueOf(product, scale)
	}
	return newBigDecimalByBigInteger(BigIntegerValueOf(x).multiplyLong(y), MIN_INT64, scale, 0)
}

func multiply2(x, y types.Long) types.Long {
	product := x * y
	ax := x.Abs()
	ay := y.Abs()
	if (ax|ay).ShiftR(31) == 0 || y == 0 || (product/y == x) {
		return product
	}
	return MIN_INT64
}

func add4__(fst *bigInteger, scale1 types.Int, snd *bigInteger, scale2 types.Int) *bigDecimal {
	rscale := scale1
	sdiff := rscale.ToLong() - scale2.ToLong()
	if sdiff != 0 {
		if sdiff < 0 {
			raise := checkScaleByBigInteger(fst, -sdiff)
			rscale = scale2
			fst = bigMultiplyPowerTenByBigInteger(fst, raise)
		} else {
			raise := checkScaleByBigInteger(snd, sdiff)
			snd = bigMultiplyPowerTenByBigInteger(snd, raise)
		}
	}
	sum := fst.Add(snd)
	if fst.signum == snd.signum {
		return newBigDecimalByBigInteger(sum, MIN_INT64, rscale, 0)
	} else {
		return valueOf3(sum, rscale, 0)
	}
}

func add4_(xs types.Long, scale1 types.Int, snd *bigInteger, scale2 types.Int) *bigDecimal {
	rscale := scale1
	sdiff := rscale.ToLong() - scale2.ToLong()
	sameSigns := (((xs >> 63) | (-xs).ShiftR(63)).ToInt()) == snd.signum // (int) ((i >> 63) | (-i >>> 63))
	var sum *bigInteger
	if sdiff < 0 {
		raise := checkScale(xs, -sdiff)
		rscale = scale2
		scaledX := longMultiplPowerTen(xs, raise)
		if scaledX == MIN_INT64 {
			sum = snd.Add(bigMultiplyPowerTen(xs, raise))
		} else {
			sum = snd.add(scaledX)
		}
	} else {
		raise := checkScaleByBigInteger(snd, sdiff)
		snd = bigMultiplyPowerTenByBigInteger(snd, raise)
		sum = snd.add(xs)
	}
	if sameSigns {
		return newBigDecimalByBigInteger(sum, MIN_INT64, rscale, 0)
	} else {
		return valueOf3(sum, rscale, 0)
	}
}

func add4(xs types.Long, scale1 types.Int, ys types.Long, scale2 types.Int) *bigDecimal {
	sdiff := scale1.ToLong() - scale2.ToLong()
	if sdiff == 0 {
		return add3(xs, ys, scale1)
	} else if sdiff < 0 {
		raise := checkScale(xs, -sdiff)
		scaledX := longMultiplPowerTen(xs, raise)
		if scaledX != MIN_INT64 {
			return add3(scaledX, ys, scale2)
		} else {
			bigsum := bigMultiplyPowerTen(xs, raise).add(ys)
			if (xs ^ ys) >= 0 {
				return newBigDecimalByBigInteger(bigsum, MIN_INT64, scale2, 0)
			} else {
				return valueOf3(bigsum, scale2, 0)
			}
		}
	} else {
		raise := checkScale(ys, sdiff)
		scaledY := longMultiplPowerTen(ys, raise)
		if scaledY != MIN_INT64 {
			return add3(xs, scaledY, scale1)
		} else {
			bigsum := bigMultiplyPowerTen(ys, raise).add(xs)
			if (xs ^ ys) >= 0 {
				return newBigDecimalByBigInteger(bigsum, MIN_INT64, scale1, 0)
			} else {
				return valueOf3(bigsum, scale1, 0)
			}
		}
	}
}

func checkScale(intCompact types.Long, val types.Long) types.Int {
	asInt := val.ToInt()
	if asInt.ToLong() != val {
		if val > MAX_INT32.ToLong() {
			asInt = MAX_INT32
		} else {
			asInt = MIN_INT32
		}
		if intCompact != 0 {
			if asInt > 0 {
				panic(errors.New("Underflow"))
			} else {
				panic(errors.New("Overflow"))
			}
		}
	}
	return asInt
}

func add3(xs types.Long, ys types.Long, scale types.Int) *bigDecimal {
	sum := add2(xs, ys)
	if sum != MIN_INT64 {
		return valueOf(sum, scale)
	}
	return newBigDecimalByBigInteger2(BigIntegerValueOf(xs).add(ys), scale)
}

func add2(xs, ys types.Long) types.Long {
	sum := xs + ys
	if ((sum ^ xs) & (sum ^ ys)) >= 0 {
		return sum
	}
	return MIN_INT64
}

func valueOf(unscaleVal types.Long, scale types.Int) *bigDecimal {
	if scale == 0 {
		return BigDecimalValueOf(unscaleVal)
	} else if unscaleVal == 0 {
		return zeroValueOf(scale)
	}
	if unscaleVal == MIN_INT64 {
		return newBigDecimalByBigInteger(BigIntegerValueOf(MIN_INT64), unscaleVal, scale, 0)
	} else {
		return newBigDecimalByBigInteger(nil, unscaleVal, scale, 0)
	}
}

func valueOf3(intVal *bigInteger, scale, prec types.Int) *bigDecimal {
	val := compactValFor(intVal)
	if val == 0 {
		return zeroValueOf(scale)
	} else if scale == 0 && val >= 0 && val < types.Long(len(p_ZERO_THROUGH_TEN)) {
		return p_ZERO_THROUGH_TEN[val.ToInt()]
	}
	return newBigDecimalByBigInteger(intVal, val, scale, prec)
}

func zeroValueOf(scale types.Int) *bigDecimal {
	if scale >= 0 && scale < types.Int(len(p_ZERO_SCALED_BY)) {
		return p_ZERO_SCALED_BY[scale]
	} else {
		return newBigDecimalByBigInteger(ZERO, 0, scale, 1)
	}
}
