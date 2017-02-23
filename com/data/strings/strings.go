package strings

import (
	"fmt"
)

// 最后的一个字符
func LastChar(str string) uint8 {
	size := len(str)
	if size == 0 {
		panic("The length of the string can't be 0")
	}
	return str[size-1]
}

// 居中
func Center(s string, length int) string {
	l := (length - len(s)) / 2
	return fmt.Sprintf("%*s%s%*s", l, "", s, l, "")
}

// 反转
func Reverse(s string) string {
	n := len(s)
	runes := make([]rune, n)
	for _, rune := range s {
		n--
		runes[n] = rune
	}
	return string(runes[n:])
}
