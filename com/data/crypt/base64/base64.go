package base64

import (
	"encoding/base64"
)

// Base64 编码
func Encode(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// Base64 字符串编码
func EncodeS(str string) string {
	return Encode([]byte(str))
}

// Base64 解码
func Decode(str string) (string, error) {
	s, e := base64.StdEncoding.DecodeString(str)
	return string(s), e
}
