package md5

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 编码
func Encode(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
