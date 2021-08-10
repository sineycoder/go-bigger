package types

/**
 @author: nizhenxian
 @date: 2021/8/11 19:18:00
**/
type Int int32
type Long int64
type Float float32
type Double float64

// e.g. right shift, append 0 to high bit
func (i Int) ShiftR(val Int) Int {
	return Int(uint32(i) >> val)
}

func (i Int) ToLong() Long {
	return Long(i)
}

func (l Long) ShiftR(val Int) Long {
	return Long(uint64(l) >> val)
}

func (l Long) ToInt() Int {
	return Int(l)
}
