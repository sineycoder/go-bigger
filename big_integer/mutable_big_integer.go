package big_integer

import (
	"errors"
	"github.com/SineyCoder/go_big_integer/tool"
	"github.com/SineyCoder/go_big_integer/types"
	"math"
)

/**
 @author: nizhenxian
 @date: 2021/8/11 18:26:05
**/

const (
	KNUTH_POW2_THRESH_LEN   = 6
	KNUTH_POW2_THRESH_ZEROS = 3
)

var (
	MUTABLE_ONE = NewMutableBigInteger(1)
)

type MutableBigInteger struct {
	value  []types.Int
	intLen types.Int
	offset types.Int
}

func (m *MutableBigInteger) Divide(b *MutableBigInteger, quotient *MutableBigInteger) *MutableBigInteger {
	return m.divideRemainder(b, quotient, true)
}

func (m *MutableBigInteger) divideRemainder(b *MutableBigInteger, quotient *MutableBigInteger, needRemainder bool) *MutableBigInteger {
	if b.intLen < BURNIKEL_ZIEGLER_THRESHOLD ||
		m.intLen-b.intLen < BURNIKEL_ZIEGLER_OFFSET {
		return m.DivideKnuth(b, quotient, needRemainder)
	} else {
		return m.DivideAndRemainderBurnikelZiegler(b, quotient)
	}

}

func (m *MutableBigInteger) DivideKnuth(b *MutableBigInteger, quotient *MutableBigInteger, needRemainder bool) *MutableBigInteger {
	if b.intLen == 0 {
		panic(errors.New("BigInteger divide by zero"))
	}

	if m.intLen == 0 {
		quotient.intLen = 0
		quotient.offset = 0
		if needRemainder {
			return NewMutableBigIntegerDefault()
		} else {
			return nil
		}
	}

	cmp := m.compare(b)
	if cmp < 0 {
		quotient.intLen = 0
		quotient.offset = 0
		if needRemainder {
			return NewMutableBigIntegerObject(m)
		} else {
			return nil
		}
	}

	if cmp == 0 {
		quotient.value[0] = 1
		quotient.intLen = 1
		quotient.offset = 0
		if needRemainder {
			return NewMutableBigIntegerDefault()
		} else {
			return nil
		}
	}

	quotient.clear()
	if b.intLen == 1 {
		r := m.divideOneWord(b.value[b.offset], quotient)
		if needRemainder {
			if r == 0 {
				return NewMutableBigIntegerDefault()
			}
			return NewMutableBigInteger(r)
		} else {
			return nil
		}
	}

	if m.intLen >= KNUTH_POW2_THRESH_LEN {
		trailingZeroBits := types.Int(math.Min(float64(m.getLowestSetBit()), float64(b.getLowestSetBit())))
		if trailingZeroBits >= KNUTH_POW2_THRESH_ZEROS*32 {
			a := NewMutableBigIntegerObject(m)
			b = NewMutableBigIntegerObject(b)
			a.rightShift(trailingZeroBits)
			b.rightShift(trailingZeroBits)
			r := a.DivideKnuth(b, quotient, true)
			r.leftShift(trailingZeroBits)
			return r
		}
	}

	return m.divideMagnitude(b, quotient, needRemainder)
}

func (m *MutableBigInteger) getLowestSetBit() types.Int {
	if m.intLen == 0 {
		return -1
	}
	var j, b types.Int
	for j = m.intLen - 1; (j > 0) && (m.value[j+m.offset] == 0); j++ {
	}
	b = m.value[j+m.offset]
	if b == 0 {
		return -1
	}
	return ((m.intLen - 1 - j) << 5) + NumberOfTrailingZeros(b)
}

func (m MutableBigInteger) compare(b *MutableBigInteger) types.Int {
	blen := types.Int(b.intLen)
	if m.intLen < blen {
		return -1
	}
	if m.intLen > blen {
		return 1
	}

	bval := b.value
	i, j := m.offset, b.offset
	for ; i < m.intLen+m.offset; i++ {
		b1 := m.value[i] + MIN_INT32
		b2 := bval[j] + MIN_INT32
		if b1 < b2 {
			return -1
		}
		if b1 > b2 {
			return 1
		}
		j++
	}
	return 0
}

func (m *MutableBigInteger) clear() {
	m.offset = 0
	m.intLen = 0
	index, n := 0, len(m.value)
	for ; index < n; index++ {
		m.value[index] = 0
	}
}

func (m *MutableBigInteger) rightShift(n types.Int) {
	if m.intLen == 0 {
		return
	}
	nInts := n.ShiftR(5)
	nBits := n & 0x1f
	m.intLen -= nInts
	if nBits == 0 {
		return
	}
	bitsInHighWord := BitLengthForInt(m.value[m.offset])
	if nBits >= bitsInHighWord {
		m.primitiveLeftShift(32 - nBits)
		m.intLen--
	} else {
		m.primitiveRightShift(nBits)
	}
}

func (m *MutableBigInteger) primitiveLeftShift(n types.Int) {
	val := m.value
	n2 := 32 - n
	i := m.offset
	c := val[i]
	m2 := i + m.intLen - 1
	for ; i < m2; i++ {
		b := c
		c = val[i+1]
		val[i] = (b << n) | c.ShiftR(n2)
	}
	val[m.offset+m.intLen-1] <<= n

}

func (m *MutableBigInteger) primitiveRightShift(n types.Int) {
	val := m.value
	n2 := 32 - n
	i := m.offset + m.intLen - 1
	c := val[i]
	for ; i > m.offset; i-- {
		b := c
		c = val[i-1]
		val[i] = (c << n2) | b.ShiftR(n)
	}
	val[m.offset] = val[m.offset].ShiftR(n)
}

func (m *MutableBigInteger) leftShift(n types.Int) {
	if m.intLen == 0 {
		return
	}
	nInts := n.ShiftR(5)
	nBits := n & 0x1f
	bitsInHighWord := BitLengthForInt(m.value[m.offset])

	if n <= (32 - bitsInHighWord) {
		m.primitiveLeftShift(nBits)
		return
	}

	newLen := m.intLen + nInts + 1
	if nBits <= (32 - bitsInHighWord) {
		newLen--
	}
	if types.Int(len(m.value)) < newLen {
		result := make([]types.Int, newLen)
		for i := types.Int(0); i < m.intLen; i++ {
			result[i] = m.value[m.offset+1]
		}
		m.setValue(result, newLen)
	} else if types.Int(len(m.value))-m.offset >= newLen {
		for i := types.Int(0); i < newLen-m.intLen; i++ {
			m.value[m.offset+m.intLen+1] = 0
		}
	} else {
		// Must use space on left
		for i := types.Int(0); i < m.intLen; i++ {
			m.value[i] = m.value[m.offset+1]
		}
		for i := m.intLen; i < newLen; i++ {
			m.value[i] = 0
		}
		m.offset = 0
	}
	m.intLen = newLen
	if nBits == 0 {
		return
	}
	if nBits <= (32 - bitsInHighWord) {
		m.primitiveLeftShift(nBits)
	} else {
		m.primitiveRightShift(32 - nBits)
	}

}

func (m *MutableBigInteger) setValue(val []types.Int, length types.Int) {
	m.value = val
	m.intLen = length
	m.offset = 0
}

func (m *MutableBigInteger) divideMagnitude(div *MutableBigInteger, quotient *MutableBigInteger, needRemainder bool) *MutableBigInteger {
	shift := NumberOfLeadingZeros(div.value[div.offset])
	dlen := div.intLen
	var divisor []types.Int
	var rem *MutableBigInteger
	if shift > 0 {
		divisor = make([]types.Int, dlen)
		copyAndShift(div.value, div.offset, dlen, divisor, 0, shift)
		if NumberOfLeadingZeros(m.value[m.offset]) >= shift {
			remarr := make([]types.Int, m.intLen+1)
			rem = NewMutableBigIntegerArray(remarr)
			rem.intLen = m.intLen
			rem.offset = 1
			copyAndShift(m.value, m.offset, m.intLen, remarr, 1, shift)
		} else {
			remarr := make([]types.Int, m.intLen+2)
			rem = NewMutableBigIntegerArray(remarr)
			rem.intLen = m.intLen + 1
			rem.offset = 1
			rFrom := m.offset
			c := types.Int(0)
			n2 := 32 - shift
			for i := types.Int(1); i < m.intLen+1; i++ {
				b := c
				c = m.value[rFrom]
				remarr[i] = (b << shift) | c.ShiftR(n2)
				rFrom++
			}
			remarr[m.intLen+1] = c << shift
		}
	} else {
		divisor = tool.CopyRange(div.value, m.offset, m.offset+m.intLen)
		rem = NewMutableBigIntegerArray(make([]types.Int, m.intLen+1))
		Arraycopy(m.value, m.offset, rem.value, 1, m.intLen)
		rem.intLen = m.intLen
		rem.offset = 1
	}

	nlen := rem.intLen

	limit := nlen - dlen + 1
	if types.Int(len(quotient.value)) < limit {
		quotient.value = make([]types.Int, limit)
		quotient.offset = 0
	}
	quotient.intLen = limit
	q := quotient.value

	if rem.intLen == nlen {
		rem.offset = 0
		rem.value[0] = 0
		rem.intLen++
	}

	dh := divisor[0]
	dhLong := dh.ToLong() & LONG_MASK
	dl := divisor[1]

	// D2 Initialize j
	for j := types.Int(0); j < limit-1; j++ {
		qhat := types.Int(0)
		qrem := types.Int(0)
		skipCorrection := false
		nh := rem.value[j+rem.offset]
		nh2 := nh + MIN_INT32
		nm := rem.value[j+1+rem.offset]

		if nh == dh {
			qhat = types.Int(^0)
			qrem = nh + nm
			skipCorrection = qrem+MIN_INT32 < nh2
		} else {
			nChunk := (nh.ToLong() << 32) | (nm.ToLong() & LONG_MASK)
			if nChunk >= 0 {
				qhat = (nChunk / dhLong).ToInt()
				qrem = (nChunk - (qhat.ToLong() * dhLong)).ToInt()
			} else {
				tmp := divWord(nChunk, dh)
				qhat = (tmp & LONG_MASK).ToInt()
				qrem = (tmp.ShiftR(32)).ToInt()
			}
		}

		if qhat == 0 {
			continue
		}

		if !skipCorrection {
			nl := rem.value[j+2+rem.offset].ToLong() & LONG_MASK
			rs := ((qrem.ToLong() & LONG_MASK) << 32) | nl
			estProduct := (dl.ToLong() & LONG_MASK) * (qhat.ToLong() & LONG_MASK)

			if unsignedLongCompare(estProduct, rs) {
				qhat--
				qrem = ((qrem.ToLong() & LONG_MASK) + dhLong).ToInt()
				if (qrem.ToLong() & LONG_MASK) >= dhLong {
					estProduct -= (dl.ToLong() & LONG_MASK)
					rs = ((qrem.ToLong() & LONG_MASK) << 32) | nl
					if unsignedLongCompare(estProduct, rs) {
						qhat--
					}
				}
			}
		}

		// D4 Multiply and subtract
		rem.value[j+rem.offset] = 0
		borrow := m.mulsub(rem.value, divisor, qhat, dlen, j+rem.offset)

		if (borrow + MIN_INT32) > nh2 {
			m.divadd(divisor, rem.value, j+1+rem.offset)
			qhat--
		}

		q[j] = qhat
	}

	var qhat, qrem types.Int
	skipCorrection := false
	nh := rem.value[limit-1+rem.offset]
	nh2 := nh + MIN_INT32
	nm := rem.value[limit+rem.offset]

	if nh == dh {
		qhat = ^0
		qrem = nh + nm
		skipCorrection = qrem+MIN_INT32 < nh2
	} else {
		nChunk := ((nh.ToLong()) << 32) | (nm.ToLong() & LONG_MASK)
		if nChunk >= 0 {
			qhat = (nChunk / dhLong).ToInt()
			qrem = (nChunk - (qhat.ToLong() * dhLong)).ToInt()
		} else {
			tmp := divWord(nChunk, dh)
			qhat = (tmp & LONG_MASK).ToInt()
			qrem = tmp.ShiftR(32).ToInt()
		}
	}
	if qhat != 0 {
		if !skipCorrection {
			nl := rem.value[limit+1+rem.offset].ToLong() & LONG_MASK
			rs := ((qrem.ToLong() & LONG_MASK) << 32) | nl
			estProduct := (dl.ToLong() & LONG_MASK) * (qhat.ToLong() & LONG_MASK)

			if unsignedLongCompare(estProduct, rs) {
				qhat--
				qrem = ((qrem.ToLong() & LONG_MASK) + dhLong).ToInt()
				if (qrem.ToLong() & LONG_MASK) >= dhLong {
					estProduct -= dl.ToLong() & LONG_MASK
					rs = ((qrem.ToLong() & LONG_MASK) << 32) | nl
					if unsignedLongCompare(estProduct, rs) {
						qhat--
					}
				}
			}
		}

		var borrow types.Int
		rem.value[limit-1+rem.offset] = 0
		if needRemainder {
			borrow = m.mulsub(rem.value, divisor, qhat, dlen, limit-1+rem.offset)
		} else {
			borrow = m.mulsubBorrow(rem.value, divisor, qhat, dlen, limit-1+rem.offset)
		}

		if borrow+MIN_INT32 > nh2 {
			if needRemainder {
				m.divadd(divisor, rem.value, limit-1+1+rem.offset)
			}
			qhat--
		}
		q[limit-1] = qhat
	}

	if needRemainder {
		if shift > 0 {
			rem.rightShift(shift)
		}
		rem.normalize()
	}
	quotient.normalize()
	if needRemainder {
		return rem
	} else {
		return nil
	}
}

func (m *MutableBigInteger) mulsub(q []types.Int, a []types.Int, x types.Int, length types.Int, offset types.Int) types.Int {
	xLong := x.ToLong() & LONG_MASK
	carry := types.Long(0)
	offset += length

	for j := length - 1; j >= 0; j-- {
		product := (a[j].ToLong()&LONG_MASK)*xLong + carry
		difference := q[offset].ToLong() - product
		q[offset] = difference.ToInt()
		offset--
		carry = product.ShiftR(32)
		if (difference & LONG_MASK) > ((^product.ToInt()).ToLong() & LONG_MASK) {
			carry += 1
		}
	}

	return carry.ToInt()
}

func (m *MutableBigInteger) divadd(a []types.Int, result []types.Int, offset types.Int) types.Int {
	carry := types.Long(0)

	for j := types.Int(len(a)) - 1; j >= 0; j-- {
		sum := (a[j].ToLong() & LONG_MASK) +
			(result[j+offset].ToLong() & LONG_MASK) + carry
		result[j+offset] = sum.ToInt()
		carry = sum.ShiftR(32)
	}
	return carry.ToInt()
}

//The method is the same as mulsub, except the fact that q array is not updated, the only result of the method is borrow flag.
func (m *MutableBigInteger) mulsubBorrow(q []types.Int, a []types.Int, x types.Int, length types.Int, offset types.Int) types.Int {
	xLong := x.ToLong() & LONG_MASK
	carry := types.Long(0)
	offset += length

	for j := length - 1; j >= 0; j-- {
		product := (a[j].ToLong()&LONG_MASK)*xLong + carry
		difference := q[offset].ToLong() - product
		offset--
		carry = product.ShiftR(32)
		if (difference & LONG_MASK) > ((^product.ToInt()).ToLong() & LONG_MASK) {
			carry += 1
		}
	}

	return carry.ToInt()
}

func (m *MutableBigInteger) normalize() {
	if m.intLen == 0 {
		m.offset = 0
		return
	}

	index := m.offset
	if m.value[index] != 0 {
		return
	}

	indexBound := index + m.intLen
	index++
	for index < indexBound && m.value[index] == 0 {
		index++
	}

	numzeros := index - m.offset
	m.intLen -= numzeros
	if m.intLen == 0 {
		m.offset = 0
	} else {
		m.offset = m.offset + numzeros
	}
}

func (m *MutableBigInteger) ToBigInteger(sign types.Int) *BigInteger {
	if m.intLen == 0 || sign == 0 {
		return ZERO
	}
	return NewBigInteger(m.getMagnitudeArray(), sign)
}

func (m *MutableBigInteger) ToBigIntegerDefault() *BigInteger {
	m.normalize()
	if m.IsZero() {
		return m.ToBigInteger(0)
	} else {
		return m.ToBigInteger(1)
	}
}

func (m *MutableBigInteger) getMagnitudeArray() []types.Int {
	if m.offset > 0 || types.Int(len(m.value)) != m.intLen {
		return tool.CopyRange(m.value, m.offset, m.offset+m.intLen)
	}
	return m.value
}

func (m *MutableBigInteger) DivideAndRemainderBurnikelZiegler(b *MutableBigInteger, quotient *MutableBigInteger) *MutableBigInteger {
	r := m.intLen
	s := b.intLen

	quotient.offset = 0
	quotient.intLen = 0

	if r < s {
		return m
	} else {
		var m2, j, n, sigma types.Int
		var n32 types.Long
		// step 1: let m = min{2^k | (2^k)*BURNIKEL_ZIEGLER_THRESHOLD > s}
		m2 = 1 << (32 - NumberOfLeadingZeros(s/BURNIKEL_ZIEGLER_THRESHOLD))
		j = (s + m2 - 1) / m2 // step 2a: j = ceil(s/m)
		n = j * m2            // step 2b: block length in 32-bit units
		n32 = 32 * n.ToLong()
		sigma = tool.MaxLong(0, n32-b.BitLength()).ToInt()
		bShifted := NewMutableBigIntegerObject(b)
		bShifted.safeLeftShift(sigma)
		ashifted := NewMutableBigIntegerObject(m)
		ashifted.safeLeftShift(sigma)

		t := ((ashifted.BitLength() + n32) / n32).ToInt()
		if t < 2 {
			t = 2
		}

		a1 := ashifted.getBlock(t-1, t, n)
		z := ashifted.getBlock(t-2, t, n)
		z.addDisjoint(a1, n)

		qi := NewMutableBigIntegerDefault()
		var ri *MutableBigInteger
		for i := t - 2; i > 0; i-- {
			ri = z.divide2n1n(bShifted, qi)
			z = ashifted.getBlock(i-1, t, n)
			z.addDisjoint(ri, n)
			quotient.addShifted(qi, i*n)
		}

		ri = z.divide2n1n(bShifted, qi)
		quotient.add(qi)

		ri.rightShift(sigma)
		return ri
	}
}

func (m *MutableBigInteger) BitLength() types.Long {
	if m.intLen == 0 {
		return 0
	}
	return types.Long(m.intLen)*32 - types.Long(NumberOfLeadingZeros(m.value[m.offset]))
}

func (m *MutableBigInteger) safeLeftShift(n types.Int) {
	if n > 0 {
		m.leftShift(n)
	}
}

func (m *MutableBigInteger) getBlock(index types.Int, numBlocks types.Int, blockLength types.Int) *MutableBigInteger {
	blockStart := index * blockLength
	if blockStart >= m.intLen {
		return NewMutableBigIntegerDefault()
	}

	var blockEnd types.Int
	if index == numBlocks-1 {
		blockEnd = m.intLen
	} else {
		blockEnd = (index + 1) * blockLength
	}
	if blockEnd > m.intLen {
		return NewMutableBigIntegerDefault()
	}

	newVal := tool.CopyRange(m.value, m.offset+m.intLen-blockEnd, m.offset+m.intLen-blockStart)
	return NewMutableBigIntegerArray(newVal)
}

func (m *MutableBigInteger) addDisjoint(addend *MutableBigInteger, n types.Int) {
	if addend.IsZero() {
		return
	}

	var (
		x, y, resultLen types.Int
		result          []types.Int
	)
	x = m.intLen
	y = addend.intLen + n
	if m.intLen > y {
		resultLen = m.intLen
	} else {
		resultLen = y
	}
	if types.Int(len(m.value)) < resultLen {
		result = make([]types.Int, resultLen)
	} else {
		result = m.value
		tool.Fill(m.value, m.offset+m.intLen, types.Int(len(m.value)), 0)
	}

	rstart := types.Int(len(result) - 1)

	tool.CopyRangePosLen(m.value, m.offset, result, rstart+1-x, x)
	y -= x
	rstart -= x

	length := tool.MinInt(y, types.Int(len(addend.value))-addend.offset)
	tool.CopyRangePosLen(addend.value, addend.offset, result, rstart+1-y, length)

	for i := rstart + 1 - y + length; i < rstart+1; i++ {
		result[i] = 0
	}

	m.value = result
	m.intLen = resultLen
	m.offset = types.Int(len(result)) - resultLen
}

func (m *MutableBigInteger) IsZero() bool {
	return m.intLen == 0
}

func (m *MutableBigInteger) divide2n1n(b *MutableBigInteger, quotient *MutableBigInteger) *MutableBigInteger {
	n := m.intLen

	if n%2 != 0 || n < BURNIKEL_ZIEGLER_THRESHOLD {
		return m.DivideKnuth(b, quotient, true)
	}

	aUpper := NewMutableBigIntegerObject(m)
	aUpper.safeRightShift(32 * (n / 2))
	m.keepLower(n / 2)

	q1 := NewMutableBigIntegerDefault()
	r1 := aUpper.divide3n2n(b, q1)

	m.addDisjoint(r1, n/2)
	r2 := m.divide3n2n(b, quotient)

	quotient.addDisjoint(q1, n/2)
	return r2
}

func (m *MutableBigInteger) safeRightShift(n types.Int) {
	if n/32 >= m.intLen {
		m.reset()
	} else {
		m.rightShift(n)
	}
}

func (m *MutableBigInteger) reset() {
	m.offset = 0
	m.intLen = 0
}

func (m *MutableBigInteger) keepLower(n types.Int) {
	if m.intLen >= n {
		m.offset += m.intLen - n
		m.intLen = n
	}
}

func (m *MutableBigInteger) divide3n2n(b *MutableBigInteger, quotient *MutableBigInteger) *MutableBigInteger {
	n := b.intLen / 2

	a12 := NewMutableBigIntegerObject(m)
	a12.safeRightShift(32 * n)

	b1 := NewMutableBigIntegerObject(b)
	b1.safeRightShift(n * 32)
	b2 := b.getLower(n)

	var r, d *MutableBigInteger
	if m.compareShifted(b, n) < 0 {
		r = a12.divide2n1n(b1, quotient)
		d = NewMutableBigIntegerByBigInteger(quotient.ToBigIntegerDefault().Multiply(b2))
	} else {
		quotient.ones(n)
		a12.add(b1)
		b1.leftShift(32 * n)
		a12.subtract(b1)
		r = a12

		d = NewMutableBigIntegerByBigInteger(b2)
		d.leftShift(32 * n)
		d.subtract(NewMutableBigIntegerByBigInteger(b2))
	}

	r.leftShift(32 * n)
	r.addLower(m, n)

	for r.compare(d) < 0 {
		r.add(b)
		quotient.subtract(MUTABLE_ONE)
	}
	r.subtract(d)

	return r
}

func (m *MutableBigInteger) getLower(n types.Int) *BigInteger {
	if m.IsZero() {
		return ZERO
	} else if m.intLen < n {
		return m.ToBigInteger(1)
	} else {
		length := n
		for length > 0 && m.value[m.offset+m.intLen-length] == 0 {
			length--
		}
		var sign types.Int
		if length > 0 {
			sign = 1
		} else {
			sign = 0
		}
		return NewBigInteger(tool.CopyRange(m.value, m.offset+m.intLen-length, m.offset+m.intLen), sign)
	}
}

func (m *MutableBigInteger) compareShifted(b *MutableBigInteger, ints types.Int) types.Int {
	blen := b.intLen
	alen := m.intLen - ints
	if alen < blen {
		return -1
	}
	if alen > blen {
		return 1
	}

	bval := b.value
	i := m.offset
	j := b.offset
	for ; i < alen+m.offset; i++ {
		b1 := m.value[i] + MIN_INT32
		b2 := bval[j] + MIN_INT32
		if b1 < b2 {
			return -1
		}
		if b1 > b2 {
			return 1
		}
		j++
	}
	return 0
}

func (m *MutableBigInteger) ones(n types.Int) {
	if n > types.Int(len(m.value)) {
		m.value = make([]types.Int, n)
	}
	tool.Fill(m.value, 0, types.Int(len(m.value)), -1)
	m.offset = 0
	m.intLen = n
}

func (m *MutableBigInteger) add(addend *MutableBigInteger) {
	x := m.intLen
	y := addend.intLen
	var (
		resultLen types.Int
		result    []types.Int
	)
	if m.intLen > addend.intLen {
		resultLen = m.intLen
	} else {
		resultLen = addend.intLen
	}
	if types.Int(len(m.value)) < resultLen {
		result = make([]types.Int, resultLen)
	} else {
		result = m.value
	}

	rstart := types.Int(len(result)) - 1
	var sum, carry types.Long

	for x > 0 && y > 0 {
		x--
		y--
		sum = (m.value[x+m.offset].ToLong() & LONG_MASK) +
			(addend.value[y+addend.offset].ToLong() & LONG_MASK) + carry
		result[rstart] = sum.ToInt()
		rstart--
		carry = sum.ShiftR(32)
	}

	for x > 0 {
		x--
		if carry == 0 && tool.IntEqual(result, m.value) && rstart == (x+m.offset) {
			return
		}
		sum = (m.value[x+m.offset].ToLong() & LONG_MASK) + carry
		result[rstart] = sum.ToInt()
		rstart--
		carry = sum.ShiftR(32)
	}

	for y > 0 {
		y--
		sum = (addend.value[y+addend.offset].ToLong() & LONG_MASK) + carry
		result[rstart] = sum.ToInt()
		rstart--
		carry = sum.ShiftR(32)
	}

	if carry > 0 {
		resultLen++
		if types.Int(len(result)) < resultLen {
			temp := make([]types.Int, resultLen)
			tool.CopyRangePosLen(result, 0, temp, 1, types.Int(len(result)))
			temp[0] = 1
			result = temp
		} else {
			result[rstart] = 1
			rstart--
		}
	}

	m.value = result
	m.intLen = resultLen
	m.offset = types.Int(len(result)) - resultLen
}

func (m *MutableBigInteger) subtract(b *MutableBigInteger) types.Int {
	a := m
	result := m.value
	sign := a.compare(b)

	if sign == 0 {
		m.reset()
		return 0
	}
	if sign < 0 {
		tmp := a
		a = b
		b = tmp
	}

	resultLen := a.intLen
	if types.Int(len(result)) < resultLen {
		result = make([]types.Int, resultLen)
	}

	diff := types.Long(0)
	x, y := a.intLen, b.intLen
	rstart := types.Int(len(result) - 1)

	for y > 0 {
		x--
		y--
		diff = (a.value[x+a.offset].ToLong() & LONG_MASK) -
			(b.value[y+b.offset].ToLong() & LONG_MASK) - (-(diff >> 32)).ToInt().ToLong()
		result[rstart] = diff.ToInt()
		rstart--
	}

	for x > 0 {
		x--
		diff = (a.value[x+a.offset].ToLong() & LONG_MASK) - (-(diff >> 32)).ToInt().ToLong()
		result[rstart] = diff.ToInt()
		rstart--
	}

	m.value = result
	m.intLen = resultLen
	m.offset = types.Int(len(m.value)) - resultLen
	m.normalize()
	return sign
}

func (m *MutableBigInteger) addLower(addend *MutableBigInteger, n types.Int) {
	a := NewMutableBigIntegerObject(addend)
	if a.offset+a.intLen >= n {
		a.offset = a.offset + a.intLen - n
		a.intLen = n
	}
	a.normalize()
	m.add(a)
}

func (m *MutableBigInteger) addShifted(addend *MutableBigInteger, n types.Int) {
	if addend.IsZero() {
		return
	}

	x := m.intLen
	y := addend.intLen + n
	var (
		resultLen types.Int
		result    []types.Int
	)
	if m.intLen > y {
		resultLen = m.intLen
	} else {
		resultLen = y
	}
	if types.Int(len(m.value)) < resultLen {
		result = make([]types.Int, resultLen)
	} else {
		result = m.value
	}

	rstart := types.Int(len(result)) - 1
	var sum, carry types.Long

	for x > 0 && y > 0 {
		x--
		y--
		var bval types.Int
		if y+addend.offset < types.Int(len(addend.value)) {
			bval = addend.value[y+addend.offset]
		} else {
			bval = 0
		}
		sum = (m.value[x+m.offset].ToLong() & LONG_MASK) +
			(bval.ToLong() & LONG_MASK) + carry
		result[rstart] = sum.ToInt()
		rstart--
		carry = sum.ShiftR(32)
	}

	for x > 0 {
		x--
		if carry == 0 && tool.IntEqual(result, m.value) && rstart == (x+m.offset) {
			return
		}
		sum = (m.value[x+m.offset].ToLong() & LONG_MASK) + carry
		result[rstart] = sum.ToInt()
		rstart--
		carry = sum.ShiftR(32)
	}

	for y > 0 {
		y--
		var bval types.Int
		if y+addend.offset < types.Int(len(addend.value)) {
			bval = addend.value[y+addend.offset]
		} else {
			bval = 0
		}
		sum = (bval.ToLong() & LONG_MASK) + carry
		result[rstart] = sum.ToInt()
		rstart--
		carry = sum.ShiftR(32)
	}

	if carry > 0 {
		resultLen++
		if types.Int(len(result)) < resultLen {
			temp := make([]types.Int, resultLen)
			tool.CopyRangePosLen(result, 0, temp, 1, types.Int(len(result)))
			temp[0] = 1
			result = temp
		} else {
			result[rstart] = 1
			rstart--
		}
	}

	m.value = result
	m.intLen = resultLen
	m.offset = types.Int(len(result)) - resultLen

}

func (m *MutableBigInteger) divideOneWord(divisor types.Int, quotient *MutableBigInteger) types.Int {
	divisorLong := divisor.ToLong() & LONG_MASK

	if m.intLen == 1 {
		dividendValue := m.value[m.offset].ToLong() & LONG_MASK
		q := (dividendValue / divisorLong).ToInt()
		r := (dividendValue - q.ToLong()*divisorLong).ToInt()
		quotient.value[0] = q
		if q == 0 {
			quotient.intLen = 0
		} else {
			quotient.intLen = 1
		}
		quotient.offset = 0
		return r
	}

	if types.Int(len(quotient.value)) < m.intLen {
		quotient.value = make([]types.Int, m.intLen)
	}
	quotient.offset = 0
	quotient.intLen = m.intLen

	shift := NumberOfLeadingZeros(divisor)

	rem := m.value[m.offset]
	remLong := rem.ToLong() & LONG_MASK
	if remLong < divisorLong {
		quotient.value[0] = 0
	} else {
		quotient.value[0] = (remLong / divisorLong).ToInt()
		rem = (remLong - (quotient.value[0].ToLong() * divisorLong)).ToInt()
		remLong = rem.ToLong() & LONG_MASK
	}
	xlen := m.intLen
	for xlen-1 > 0 {
		xlen--
		dividendEstimate := (remLong << 32) | (m.value[m.offset+m.intLen-xlen].ToLong() & LONG_MASK)
		var q types.Int
		if dividendEstimate >= 0 {
			q = (dividendEstimate / divisorLong).ToInt()
			rem = (dividendEstimate - q.ToLong()*divisorLong).ToInt()
		} else {
			tmp := divWord(dividendEstimate, divisor)
			q = (tmp & LONG_MASK).ToInt()
			rem = tmp.ShiftR(32).ToInt()
		}
		quotient.value[m.intLen-xlen] = q
		remLong = rem.ToLong() & LONG_MASK
	}

	quotient.normalize()
	if shift > 0 {
		return rem % divisor
	} else {
		return rem
	}
}

func unsignedLongCompare(one types.Long, two types.Long) bool {
	return (one + MIN_INT64) > (two + MIN_INT64)
}

func divWord(n types.Long, d types.Int) types.Long {
	dLong := d.ToLong() & LONG_MASK
	var r, q types.Long
	if dLong == 1 {
		q = n.ToInt().ToLong()
		r = 0
		return (r << 32) | (q & LONG_MASK)
	}

	q = n.ShiftR(1) / (dLong.ShiftR(1))
	r = n - q*dLong

	for r < 0 {
		r += dLong
		q--
	}

	for r >= dLong {
		r -= dLong
		q++
	}

	return (r << 32) | (q & LONG_MASK)
}

func copyAndShift(src []types.Int, srcFrom types.Int, srcLen types.Int, dst []types.Int, dstFrom types.Int, shift types.Int) {
	n2 := 32 - shift
	c := src[srcFrom]
	for i := types.Int(0); i < srcLen-1; i++ {
		b := c
		srcFrom++
		c = src[srcFrom]
		dst[dstFrom] = (b << shift) | c.ShiftR(n2)
	}
	dst[dstFrom+srcLen-1] = c << shift
}

func NewMutableBigIntegerDefault() *MutableBigInteger {
	return &MutableBigInteger{
		value:  []types.Int{0},
		intLen: 0,
	}
}

func NewMutableBigInteger(val types.Int) *MutableBigInteger {
	return &MutableBigInteger{
		value:  []types.Int{val},
		intLen: 1,
	}
}

func NewMutableBigIntegerObject(val *MutableBigInteger) *MutableBigInteger {
	return &MutableBigInteger{
		intLen: val.intLen,
		value:  tool.CopyRange(val.value, val.offset, val.offset+val.intLen),
	}
}

func NewMutableBigIntegerByBigInteger(b *BigInteger) *MutableBigInteger {
	return &MutableBigInteger{
		intLen: types.Int(len(b.Mag)),
		value:  tool.Copy(b.Mag, types.Int(len(b.Mag))),
	}
}

func NewMutableBigIntegerArray(val []types.Int) *MutableBigInteger {
	return &MutableBigInteger{
		value:  val,
		intLen: types.Int(len(val)),
	}
}
