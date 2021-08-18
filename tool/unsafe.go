package tool

import "unsafe"

/**
 @author: nizhenxian
 @date: 2021/8/17 10:13:29
**/

func StrToBytes(str string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&str)) // 内存结构中可以得知，string2个参数，byte3个参数，多了cap
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func BytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b)) // 忽略内存中的cap，可以直接转为string
}
