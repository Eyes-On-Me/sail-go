package csrf

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sail-services/sail-go/com/data/convert"
	"github.com/sail-services/sail-go/mod/net/service"
	"github.com/sail-services/sail-go/mod/net/service/mod/session"
)

type (
	Options struct {
		SecretKey      string // 密钥 [nil]
		Header         string // Header 中的 Token 名 [X-CSRF]
		Form           string // Post 中的 Token 名 [CSRF]
		Cookie         string // Cookie 中的 Token 名 [CSRF]
		CookiePath     string // Cookie 的路径 [/]
		Session        string // 要处理的 Session 名 [nil]
		Origin         bool
		RespHaveHeader bool                        // 返回 Header 是否有密钥 [false]
		RespHaveCookie bool                        // 返回 Cookie 是否有密钥 [false]
		ErrorFunc      func(w http.ResponseWriter) // 设置出错 func
	}
	CSRF interface {
		HeaderGet() string        // 获取 Header 中的 Token 名
		FormGet() string          // 获取 Form 中的 Token 名
		CookieGet() string        // 获取 Cookie 中的 Token 名
		CookiePathGet() string    // 获取 Cookie 路径
		TokenGet() string         // 获取 Token
		TokenValid(t string) bool // 验证 Token
		IDGet() string            // 获取 ID
		Error(w http.ResponseWriter)
	}
	csrf struct {
		Header     string
		Form       string
		Token      string
		ID         string
		Cookie     string
		CookiePath string
		SecretKey  string
		ErrorFunc  func(w http.ResponseWriter)
	}
)

const (
	_DATA_CSRF = "_DATA_CSRF"
	_TIMEOUT   = 24 * time.Hour
)

func New(opts ...Options) service.Handler {
	var id_have bool
	opt := optPrepare(opts)
	return func(con *service.Context) {
		x := &csrf{
			SecretKey:  opt.SecretKey,
			Header:     opt.Header,
			Form:       opt.Form,
			Cookie:     opt.Cookie,
			CookiePath: opt.CookiePath,
			ErrorFunc:  opt.ErrorFunc}
		con.DataSet(_DATA_CSRF, x)
		sess := session.DataGetStore(con)
		id := sess.Get(opt.Session)
		if id == nil {
			id, id_have = con.Req.SecureCookieGet(opt.Session)
			if !id_have {
				id = "0"
				con.Resp.CookieSet(x.CookieGet(), "", -1, x.CookiePathGet())
			} else {
				sess.Set(opt.Session, id)
			}
		}
		switch id.(type) {
		case string:
			x.ID = id.(string)
		case int:
		case int64:
			x.ID = convert.ToS(id)
		default:
			return
		}
		if opt.Origin && len(con.Req.Header.Get("Origin")) > 0 {
			return
		}
		if val := con.Req.CookieGet(opt.Cookie); len(val) != 0 {
			x.Token = val
		} else {
			tm := time.Now()
			x.Token = generateTokenAtTime(x.SecretKey, x.ID, "POST", tm)
			if opt.RespHaveCookie && x.ID != "0" {
				con.Resp.CookieSet(opt.Cookie, x.Token, 0, opt.CookiePath)
			}
		}
		if opt.RespHaveHeader {
			con.Resp.Header().Add(opt.Header, x.Token)
		}
	}
}

func Validate(con *service.Context) {
	x := DataCSRFGet(con)
	token := con.Req.Header.Get(x.HeaderGet())
	if token == "" {
		token = con.Req.FormValue(x.FormGet())
	} else {
		if !x.TokenValid(token) {
			con.Resp.CookieSet(x.CookieGet(), "", -1, x.CookiePathGet())
			x.Error(con.Resp)
		}
		return
	}
	http.Error(con.Resp, "Bad Request: no CSRF token represnet", http.StatusBadRequest)
}

func ValidateWithSAndID(con *service.Context, token, id string) bool {
	x := DataCSRFGet(con)
	if x.IDGet() != id {
		return false
	}
	if !x.TokenValid(token) {
		return false
	}
	return true
}

func DataCSRFGet(con *service.Context) CSRF {
	return con.DataMustGet(_DATA_CSRF).(CSRF)
}

func optPrepare(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}
	if opt.SecretKey == "" {
		panic("Must have SecretKey")
	}
	if opt.Header == "" {
		opt.Header = "X-CSRF"
	}
	if opt.Form == "" {
		opt.Form = "CSRF"
	}
	if opt.Cookie == "" {
		opt.Cookie = "CSRF"
	}
	if opt.CookiePath == "" {
		opt.CookiePath = "/"
	}
	if opt.Session == "" {
		panic("Must have Session")
	}
	return opt
}

func generateTokenAtTime(secret_key, user_id, action_id string, now time.Time) string {
	h := hmac.New(sha1.New, []byte(secret_key))
	fmt.Fprintf(h, "%s:%s:%d", strings.Replace(user_id, ":", "_", -1), strings.Replace(action_id, ":", "_", -1), now.UnixNano())
	token := fmt.Sprintf("%s:%d", h.Sum(nil), now.UnixNano())
	return base64.URLEncoding.EncodeToString([]byte(token))
}

// ========================================================
// csrf
// ========================================================
func (c *csrf) HeaderGet() string {
	return c.Header
}

func (c *csrf) FormGet() string {
	return c.Form
}

func (c *csrf) CookieGet() string {
	return c.Cookie
}

func (c *csrf) IDGet() string {
	return c.ID
}

func (c *csrf) CookiePathGet() string {
	return c.CookiePath
}

func (c *csrf) TokenGet() string {
	return c.Token
}

func (c *csrf) TokenValid(token string) bool {
	now := time.Now()
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	sep := bytes.LastIndex(data, []byte{':'})
	if sep < 0 {
		return false
	}
	nanos, err := strconv.ParseInt(string(data[sep+1:]), 10, 64)
	if err != nil {
		return false
	}
	issueTime := time.Unix(0, nanos)
	if now.Sub(issueTime) >= _TIMEOUT {
		return false
	}
	if issueTime.After(now.Add(1 * time.Minute)) {
		return false
	}
	expected := generateTokenAtTime(c.SecretKey, c.ID, "POST", issueTime)
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}

func (c *csrf) Error(w http.ResponseWriter) {
	c.ErrorFunc(w)
}
