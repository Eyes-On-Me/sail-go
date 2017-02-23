package convert

import (
	"crypto/rsa"
	"crypto/x509"
	"github.com/sail-services/sail-go/com/data/number"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type (
	STo    string
	argInt []int
)

func TimeToTimestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func RSAPubKeySToRSAPubKey(str string) *rsa.PublicKey {
	block, _ := pem.Decode([]byte(str))
	key, _ := x509.ParsePKIXPublicKey(block.Bytes)
	return key.(*rsa.PublicKey)
}

func RSAPrivKeySToRSAPrivKey(str string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(str))
	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	return key
}

func RSAPubKeyToS(title string, key *rsa.PublicKey) string {
	pkix_bytes, _ := x509.MarshalPKIXPublicKey(key)
	bytes := pem.EncodeToMemory(&pem.Block{
		Type:  title,
		Bytes: pkix_bytes,
	})
	return string(bytes)
}

func SToTCPAddr(str string) (addr *net.TCPAddr) {
	addr, _ = net.ResolveTCPAddr("tcp", str)
	return
}

// []byte 到 string
func BToS(bts []byte) string {
	return string(bts)
}

// string 到 []byte
func SToB(s string) []byte {
	return []byte(s)
}

// string 到 float64
func SToF64(s string) (f float64) {
	f, _ = strconv.ParseFloat(s, 64)
	return
}

// float64 到 string
func F64ToS(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 64)
}

// string 到 int64
func SToI64(s string) (f int64) {
	f, _ = strconv.ParseInt(s, 0, 64)
	return
}

// int64 到 string
func I64ToS(num int64) string {
	return strconv.FormatInt(num, 10)
}

// int 到 string
func IToS(num int) string {
	return strconv.Itoa(num)
}

// *http.Response 到 []byte
func RespToB(resp *http.Response) (bytes []byte) {
	bytes, _ = ioutil.ReadAll(resp.Body)
	return
}

// string 到 URL
func SToUrlS(str string) string {
	return url.QueryEscape(str)
}

// URL 到 string
func UrlSToS(str string) (string, error) {
	return url.QueryUnescape(str)
}

// int 到 hex string
func IToHexS(num int) (hex string) {
	if num == 0 {
		return "0"
	}
	for num > 0 {
		r := num % 16
		c := "?"
		if r >= 0 && r <= 9 {
			c = string(r + '0')
		} else {
			c = string(r + 'a' - 10)
		}
		hex = c + hex
		num = num / 16
	}
	return hex
}

// hex string 到 int
func HexSToI(hex string) (int, error) {
	num := 0
	length := len(hex)
	for i := 0; i < length; i++ {
		char := hex[length-i-1]
		factor := -1
		switch {
		case char >= '0' && char <= '9':
			factor = int(char) - '0'
		case char >= 'a' && char <= 'f':
			factor = int(char) - 'a' + 10
		default:
			return -1, fmt.Errorf("invalid hex: %s", string(char))
		}
		num += factor * number.Pow(16, i)
	}
	return num, nil
}

// ========================================================
// ToS
// ========================================================
func ToS(value interface{}, args ...int) (s string) {
	switch v := value.(type) {
	case bool:
		s = strconv.FormatBool(v)
	case float32:
		s = strconv.FormatFloat(float64(v), 'f', argInt(args).get(0, -1), argInt(args).get(1, 32))
	case float64:
		s = strconv.FormatFloat(v, 'f', argInt(args).get(0, -1), argInt(args).get(1, 64))
	case int:
		s = strconv.FormatInt(int64(v), argInt(args).get(0, 10))
	case int8:
		s = strconv.FormatInt(int64(v), argInt(args).get(0, 10))
	case int16:
		s = strconv.FormatInt(int64(v), argInt(args).get(0, 10))
	case int32:
		s = strconv.FormatInt(int64(v), argInt(args).get(0, 10))
	case int64:
		s = strconv.FormatInt(v, argInt(args).get(0, 10))
	case uint:
		s = strconv.FormatUint(uint64(v), argInt(args).get(0, 10))
	case uint8:
		s = strconv.FormatUint(uint64(v), argInt(args).get(0, 10))
	case uint16:
		s = strconv.FormatUint(uint64(v), argInt(args).get(0, 10))
	case uint32:
		s = strconv.FormatUint(uint64(v), argInt(args).get(0, 10))
	case uint64:
		s = strconv.FormatUint(v, argInt(args).get(0, 10))
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		s = fmt.Sprintf("%v", v)
	}
	return s
}

func (arg argInt) get(i int, args ...int) (r int) {
	if i >= 0 && i < len(arg) {
		r = arg[i]
	} else if len(args) > 0 {
		r = args[0]
	}
	return
}

// ========================================================
// STo
// ========================================================
func (st STo) Exist() bool {
	return string(st) != string(0x1E)
}

func (st STo) UI8() (uint8, error) {
	v, err := strconv.ParseUint(st.S(), 10, 8)
	return uint8(v), err
}

func (st STo) I() (int, error) {
	v, err := strconv.ParseInt(st.S(), 10, 32)
	return int(v), err
}

func (st STo) I64() (int64, error) {
	v, err := strconv.ParseInt(st.S(), 10, 64)
	return int64(v), err
}

func (st STo) MustUI8() uint8 {
	v, _ := st.UI8()
	return v
}

func (st STo) MustI() int {
	v, _ := st.I()
	return v
}

func (st STo) MustI64() int64 {
	v, _ := st.I64()
	return v
}

func (st STo) S() string {
	if st.Exist() {
		return string(st)
	}
	return ""
}
