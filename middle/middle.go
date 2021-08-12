package middle

import "github.com/SineyCoder/go_big_integer/types"

/**
 @author: nizhenxian
 @date: 2021/8/12 17:57:41
**/

type MutableBigInteger struct {
	value  []types.Int
	intLen types.Int
	offset types.Int
}
