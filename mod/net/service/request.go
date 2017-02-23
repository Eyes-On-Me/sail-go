package service

import (
	"crypto/md5"
	"encoding/hex"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/sail-services/sail-go/com/data/convert"
	"github.com/sail-services/sail-go/com/data/crypt/aes"
)

type (
	Request struct {
		*http.Request
		params reqParams
		con    *Context
	}
	RequestBody struct {
		reader io.ReadCloser
	}
	reqParams map[string]string
)

// ========================================================
// Request
// ========================================================
func (req *Request) Body() *RequestBody {
	return &RequestBody{req.Request.Body}
}

func (req *Request) IP() string {
	ip := req.Header.Get("X-Real-IP")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = req.RemoteAddr
			if i := strings.LastIndex(ip, ":"); i > -1 {
				ip = ip[:i]
			}
		}
	}
	return ip
}

func (req *Request) FileGet(name string) (multipart.File, *multipart.FileHeader, error) {
	return req.FormFile(name)
}

// --------------------------------------------------------
// Request - Form
// --------------------------------------------------------
func (req *Request) FormGet(name string) string {
	req.formParse()
	return req.Form.Get(name)
}

func (req *Request) FormGetTrim(name string) string {
	return strings.TrimSpace(req.FormGet(name))
}

func (req *Request) FormGetEscape(name string) string {
	return template.HTMLEscapeString(req.FormGet(name))
}

func (req *Request) FormGetI(name string) int {
	return convert.STo(req.FormGet(name)).MustI()
}

func (req *Request) FormGetI64(name string) int64 {
	return convert.STo(req.FormGet(name)).MustI64()
}

func (req *Request) FormGetF64(name string) float64 {
	return convert.SToF64(req.FormGet(name))
}

func (req *Request) FormGets(name string) []string {
	req.formParse()
	vals, ok := req.Form[name]
	if !ok {
		return []string{}
	}
	return vals
}

func (req *Request) formParse() {
	if req.Form != nil {
		return
	}
	content_type := req.Header.Get("Content-Type")
	if (req.Method == "POST" || req.Method == "PUT") &&
		len(content_type) > 0 && strings.Contains(content_type, "multipart/form-data") {
		req.ParseMultipartForm(_FORM_MEMORY)
	} else {
		req.ParseForm()
	}
}

// --------------------------------------------------------
// Request - Param
// --------------------------------------------------------
func (req *Request) ParamSet(name, val string) {
	if !strings.HasPrefix(name, ":") {
		name = ":" + name
	}
	req.params[name] = val
}

// con.ParamGet(":uid")
func (req *Request) ParamGet(name string) string {
	return req.params[name]
}

// con.ParamGetEscape(":uname")
func (req *Request) ParamGetEscape(name string) string {
	return template.HTMLEscapeString(req.ParamGet(name))
}

// con.ParamGetI(":uid")
func (req *Request) ParamGetI(name string) int {
	return convert.STo(req.ParamGet(name)).MustI()
}

// con.ParamGetI64(":uid")
func (req *Request) ParamGetI64(name string) int64 {
	return convert.STo(req.ParamGet(name)).MustI64()
}

// con.ParamGetF64(":uid")
func (req *Request) ParamGetF64(name string) float64 {
	return convert.SToF64(req.ParamGet(name))
}

// --------------------------------------------------------
// Request - Cookie
// --------------------------------------------------------
func (req *Request) CookieGet(name string) string {
	cookie, err := req.Cookie(name)
	if err != nil {
		return ""
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val
}

func (req *Request) CookieGetI(name string) int {
	return convert.STo(req.CookieGet(name)).MustI()
}

func (req *Request) CookieGetI64(name string) int64 {
	return convert.STo(req.CookieGet(name)).MustI64()
}

func (req *Request) CookieGetF64(name string) float64 {
	return convert.SToF64(req.CookieGet(name))
}

func (req *Request) SecureCookieGet(name string) (string, bool) {
	if _secret_key == "" {
		req.con.Log.Fatalln("Not Set Secret Key")
	}
	val := req.CookieGet(name)
	if val == "" {
		return "", false
	}
	data, err := hex.DecodeString(val)
	if err != nil {
		return "", false
	}
	secret := _secret_key
	m := md5.Sum([]byte(secret))
	secret = hex.EncodeToString(m[:])
	text, err := aes.Decrypt(data, []byte(secret))
	return string(text), err == nil
}

// ========================================================
// RequestBody
// ========================================================
func (rb *RequestBody) B() ([]byte, error) {
	return ioutil.ReadAll(rb.reader)
}

func (rb *RequestBody) S() (string, error) {
	data, err := rb.B()
	return string(data), err
}

// --------------------------------------------------------
// RequestBody - Go
// --------------------------------------------------------
func (rb *RequestBody) ReadCloser() io.ReadCloser {
	return rb.reader
}
