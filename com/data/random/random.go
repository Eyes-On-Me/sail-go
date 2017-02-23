package random

import (
	crand "crypto/rand"
	"io"
	"math/rand"
	"time"
)

const (
	RANDOM_ALL               = 0
	RANDOM_STRING            = 1
	RANDOM_STRING_UPPER      = 2
	RANDOM_STRING_LOWER      = 3
	RANDOM_STRING_AND_NUMBER = 4
	RANDOM_NUMBER            = 5
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func I(min, max int) int {
	return min + rand.Intn((max+1)-min)
}

func B(bytes_len int, rand_len int) []byte {
	array := make([]byte, bytes_len)
	io.ReadAtLeast(crand.Reader, array, rand_len)
	return array
}

func S(str_type int, length int) string {
	var char string
	switch str_type {
	case RANDOM_ALL:
		char = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!#$%&'_(})~|{*`+,-./:;<=>?@[]^\\\""
	case RANDOM_STRING:
		char = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case RANDOM_STRING_UPPER:
		char = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case RANDOM_STRING_LOWER:
		char = "abcdefghijklmnopqrstuvwxyz"
	case RANDOM_STRING_AND_NUMBER:
		char = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	case RANDOM_NUMBER:
		char = "0123456789"
	}
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = char[rand.Intn(len(char)-1)]
	}
	return string(buf)
}
