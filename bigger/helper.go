package bigger

import (
	"errors"
	"github.com/sineycoder/go-bigger/types"
)

/**
 @author: nizhenxian
 @date: 2021/8/15 19:07:03
**/

var (
	DIGIT_TENS = []rune{
		'0', '0', '0', '0', '0', '0', '0', '0', '0', '0',
		'1', '1', '1', '1', '1', '1', '1', '1', '1', '1',
		'2', '2', '2', '2', '2', '2', '2', '2', '2', '2',
		'3', '3', '3', '3', '3', '3', '3', '3', '3', '3',
		'4', '4', '4', '4', '4', '4', '4', '4', '4', '4',
		'5', '5', '5', '5', '5', '5', '5', '5', '5', '5',
		'6', '6', '6', '6', '6', '6', '6', '6', '6', '6',
		'7', '7', '7', '7', '7', '7', '7', '7', '7', '7',
		'8', '8', '8', '8', '8', '8', '8', '8', '8', '8',
		'9', '9', '9', '9', '9', '9', '9', '9', '9', '9',
	}
	DIGIT_ONES = []rune{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	}
)

type StringBuilderHelper struct {
	buf          []rune
	cmpCharArray []rune
}

func NewStringBuilderHelper() *StringBuilderHelper {
	return &StringBuilderHelper{
		buf:          make([]rune, 1<<5),
		cmpCharArray: make([]rune, 19),
	}
}

func (h *StringBuilderHelper) PutIntCompact(intCompact types.Long) types.Int {
	if intCompact >= 0 {
		var q types.Long
		var r types.Int
		charPos := types.Int(len(h.cmpCharArray))

		for intCompact > MAX_INT32.ToLong() {
			q = intCompact / 100
			r = (intCompact - q*100).ToInt()
			intCompact = q
			charPos--
			h.cmpCharArray[charPos] = DIGIT_ONES[r]
			charPos--
			h.cmpCharArray[charPos] = DIGIT_TENS[r]
		}

		var q2, i2 types.Int
		i2 = intCompact.ToInt()
		for i2 >= 100 {
			q2 = i2 / 100
			r = i2 - q2*100
			i2 = q2
			charPos--
			h.cmpCharArray[charPos] = DIGIT_ONES[r]
			charPos--
			h.cmpCharArray[charPos] = DIGIT_TENS[r]
		}
		charPos--
		h.cmpCharArray[charPos] = DIGIT_ONES[i2]
		if i2 >= 10 {
			charPos--
			h.cmpCharArray[charPos] = DIGIT_TENS[i2]
		}
		return charPos
	}
	panic(errors.New("illegal param"))
}

func (h *StringBuilderHelper) GetCompactCharArray() []rune {
	return h.cmpCharArray
}

func (h *StringBuilderHelper) GetBuffer() []rune {
	return h.buf[:0]
}
