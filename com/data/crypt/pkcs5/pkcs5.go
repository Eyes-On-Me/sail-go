package pkcs5

import (
	"bytes"
	"errors"
)

func Padding(src []byte, block_size int) []byte {
	padding := block_size - len(src)%block_size
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func Unpadding(src []byte, block_size int) ([]byte, error) {
	length := len(src)
	padding := int(src[length-1])
	if padding < 0 || padding > length {
		return nil, errors.New("PKCS5 Unpadding Error")
	}
	return src[:length-padding], nil
}
