package service

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"github.com/sail-services/sail-go/com/data/crypt/aes"
	"time"
)

type (
	Response interface {
		http.ResponseWriter
		http.Hijacker
		http.Flusher
		http.CloseNotifier
		Status() int
		Size() int
		IsWritten() bool
		Before(func(Response))
		CookieSet(name, value string, others ...interface{})
		SecureCookieSet(name, value string, others ...interface{})
		writeHeader()
	}
	response struct {
		http.ResponseWriter
		con         *Context
		size        int
		status      int
		beforeFuncs []func(Response)
	}
)

// ========================================================
// response
// ========================================================
func (resp *response) Status() int {
	return resp.status
}

func (resp *response) Size() int {
	return resp.size
}

func (resp *response) IsWritten() bool {
	return resp.size != -1
}

func (resp *response) Before(before func(Response)) {
	resp.beforeFuncs = append(resp.beforeFuncs, before)
}

func (resp *response) reset(http_resp http.ResponseWriter, con *Context) {
	resp.ResponseWriter = http_resp
	resp.con = con
	resp.size = -1
	resp.status = 200
	resp.beforeFuncs = nil
}

func (resp *response) callBefore() {
	for i := len(resp.beforeFuncs) - 1; i >= 0; i-- {
		resp.beforeFuncs[i](resp)
	}
}

func (resp *response) writeHeader() {
	if !resp.IsWritten() {
		resp.size = 0
		resp.ResponseWriter.WriteHeader(resp.status)
	}
}

// --------------------------------------------------------
// response - Cookie
// --------------------------------------------------------
func (resp *response) CookieSet(name, value string, others ...interface{}) {
	cookie := http.Cookie{}
	cookie.Name = name
	cookie.Value = url.QueryEscape(value)
	if len(others) > 0 {
		switch v := others[0].(type) {
		case int:
			cookie.MaxAge = v
			cookie.Expires = time.Now().Add(time.Duration(v) * time.Second)
		case int64:
			cookie.MaxAge = int(v)
			cookie.Expires = time.Now().Add(time.Duration(v) * time.Second)
		case int32:
			cookie.MaxAge = int(v)
			cookie.Expires = time.Now().Add(time.Duration(v) * time.Second)
		}
	}
	cookie.Path = "/"
	if len(others) > 1 {
		if v, ok := others[1].(string); ok && len(v) > 0 {
			cookie.Path = v
		}
	}
	if len(others) > 2 {
		if v, ok := others[2].(string); ok && len(v) > 0 {
			cookie.Domain = v
		}
	}
	if len(others) > 3 {
		switch v := others[3].(type) {
		case bool:
			cookie.Secure = v
		default:
			if others[3] != nil {
				cookie.Secure = true
			}
		}
	}
	if len(others) > 4 {
		if v, ok := others[4].(bool); ok && v {
			cookie.HttpOnly = true
		}
	}
	resp.Header().Add("Set-Cookie", cookie.String())
}

func (resp *response) SecureCookieSet(name, value string, others ...interface{}) {
	if _secret_key == "" {
		resp.con.Log.Fatalln("Not Set Secret Key")
	}
	if value == "" {
		resp.CookieSet(name, value, others...)
		return
	}
	secret := _secret_key
	m := md5.Sum([]byte(secret))
	secret = hex.EncodeToString(m[:])
	text, err := aes.Encrypt([]byte(value), []byte(secret))
	if err != nil {
		resp.con.Log.Panic("error encrypting cookie: " + err.Error())
	}
	resp.CookieSet(name, hex.EncodeToString(text), others...)
}

// --------------------------------------------------------
// response - GO
// --------------------------------------------------------
func (resp *response) Write(data []byte) (n int, err error) {
	resp.writeHeader()
	n, err = resp.ResponseWriter.Write(data)
	resp.size += n
	return
}

func (resp *response) WriteHeader(code int) {
	if code > 0 {
		resp.callBefore()
		resp.status = code
		if resp.IsWritten() {
			resp.con.Log.Errorln("WriteHeader Error")
		}
	}
}

func (resp *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	resp.size = 0
	return resp.ResponseWriter.(http.Hijacker).Hijack()
}

func (resp *response) CloseNotify() <-chan bool {
	return resp.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (resp *response) Flush() {
	resp.ResponseWriter.(http.Flusher).Flush()
}
