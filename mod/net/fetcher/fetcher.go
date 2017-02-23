package fetcher

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type Fetcher struct {
	HTTPS     bool
	Host      string
	Referer   string
	CacheTime int64
	AutoHost  bool
	Cookies   []*http.Cookie
	Header    custom_header
	Client    *http.Client              `json:"-"`
	Cache     map[string]cache_response `json:"-"`
}

func New(host string) (f *Fetcher) {
	f = fetchNew(nil)
	f.Host = host
	return
}

func NewProxy(host, p string) (f *Fetcher) {
	u, _ := url.Parse(p)
	dialer, _ := proxy.FromURL(u,
		&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		})
	tr := &http.Transport{
		Proxy:               nil,
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	f = fetchNew(tr)
	f.Host = host
	return
}

func NewHTTPS(host string) (f *Fetcher) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            nil,
			InsecureSkipVerify: true,
		},
		DisableCompression: true,
	}
	f = fetchNew(tr)
	f.Host = host
	f.HTTPS = true
	return
}

func NewS(str string) (f *Fetcher, err error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return
	}
	f = fetchNew(nil)
	err = json.Unmarshal(data, f)
	return
}

func (f *Fetcher) GetBase64(path string) (data string, err error) {
	resp, err := f.Get(path)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode/100 != 2 {
		if err == nil {
			return "", errors.New("fetcher: error not excepted!")
		}
		err = errors.New(err.Error())
		return
	}
	data = base64.StdEncoding.EncodeToString(body)
	return
}

func (f *Fetcher) SaveFile(path, dstPath string) (err error) {
	resp, err := f.Get(path)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	ioutil.WriteFile(dstPath, body, os.ModePerm)
	return
}

func (f *Fetcher) RemoveGetCache(path string) {
	key := "get-" + "http"
	if f.HTTPS {
		key += "s"
	}
	key += "://" + f.Host + path
	_, ok := f.Cache[key]
	if !ok {
		return
	}
	delete(f.Cache, key)
}

func (f *Fetcher) RemovePostCache(path string, params url.Values) {
	key := "post-" + path + params.Encode()
	delete(f.Cache, key)
}

func (f *Fetcher) GetBody(path string) (body []byte, err error) {
	path = f.makeUrl(path)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return
	}
	key := "get-" + path
	resp, ok := f.loadCache(key)
	if ok {
		return
	}
	resp, err = f.request(req)
	if err != nil {
		return
	}
	if f.CacheTime > 0 {
		f.saveCache(key, resp)
	}
	body, err = ioutil.ReadAll(resp.Body)
	return
}

func (f *Fetcher) Get(path string) (resp *http.Response, err error) {
	path = f.makeUrl(path)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return
	}
	key := "get-" + path
	resp, ok := f.loadCache(key)
	if ok {
		return
	}
	resp, err = f.request(req)
	if err != nil {
		return
	}
	if f.CacheTime > 0 {
		f.saveCache(key, resp)
	}
	return
}

func (f *Fetcher) GetWithNoCache(path string) (resp *http.Response, err error) {
	path = f.makeUrl(path)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return
	}
	key := "get-" + path
	resp, err = f.request(req)
	if err != nil {
		return
	}
	if f.CacheTime > 0 {
		f.saveCache(key, resp)
	}
	return
}

func (f *Fetcher) Post(path, contentType string, content io.Reader) (resp *http.Response, err error) {
	path = f.makeUrl(path)
	req, err := http.NewRequest("POST", path, content)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", contentType)
	resp, err = f.request(req)
	if err != nil {
		return
	}
	f.Referer = path
	return
}

func (f *Fetcher) PostForm(path string, val url.Values) (resp *http.Response, err error) {
	contentType := "application/x-www-form-urlencoded"
	if val == nil {
		val = url.Values{}
	}
	resp, err = f.Post(path, contentType, strings.NewReader(val.Encode()))
	return
}

func (f *Fetcher) PostFormGetBody(path string, val url.Values) (body []byte, err error) {
	contentType := "application/x-www-form-urlencoded"
	if val == nil {
		val = url.Values{}
	}
	resp, err := f.Post(path, contentType, strings.NewReader(val.Encode()))
	body, err = ioutil.ReadAll(resp.Body)
	return
}

func (f *Fetcher) PostFormRetry(
	path string, val url.Values, tryTime int) (resp *http.Response, err error) {
	for i := 0; i < tryTime; i++ {
		resp, err = f.PostForm(path, val)
		if err == nil {
			break
		}
	}
	return
}

func (f *Fetcher) CallPostForm(v interface{}, path string, val url.Values) (err error) {
	resp, err := f.PostForm(path, val)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		err = errors.New("unmarshal fail: " + string(body) + ", " + err.Error())
		return
	}
	return
}

func (f *Fetcher) Store() (ret string, err error) {
	data, err := json.Marshal(f)
	if err != nil {
		return
	}
	ret = base64.StdEncoding.EncodeToString(data)
	return
}

func (f *Fetcher) getCacheKey(req *http.Request) string {
	q := req.URL.Query()
	key := req.URL.Path + q.Encode()
	return key
}

func (f *Fetcher) loadCache(key string) (resp *http.Response, ok bool) {
	r, ok := f.Cache[key]
	if !ok {
		return
	}
	ok = false
	if time.Now().Unix()-r.CacheTime > f.CacheTime {
		delete(f.Cache, key)
		return
	}
	resp = r.Resp
	ok = true
	return
}

func (f *Fetcher) saveCache(key string, resp *http.Response) {
	r := cache_response{
		resp, time.Now().Unix(),
	}
	f.Cache[key] = r
}

func (f *Fetcher) request(req *http.Request) (resp *http.Response, err error) {
	resp, err = f.Client.Do(req)
	return
}

func (f *Fetcher) insertReferer(req *http.Request) (err error) {
	if f.Referer != "" {
		req.Header.Set("Referer", f.Referer)
	}
	return
}

func (f *Fetcher) insertCookie(req *http.Request) (err error) {
	for _, cookie := range f.Cookies {
		req.AddCookie(cookie)
	}
	return
}

func (f *Fetcher) mergeCookie(resp *http.Response) (err error) {
	cookies := resp.Cookies()
	newCookies := make([]*http.Cookie, len(cookies))
	length := 0
	for _, c := range cookies {
		for idx, cs := range f.Cookies {
			if c.Name == cs.Name {
				f.Cookies[idx] = c
				goto next
			}
		}
		newCookies[length] = c
		length++
	next:
		continue
	}
	f.Cookies = append(f.Cookies, newCookies[:length]...)
	return
}

func (f *Fetcher) makeUrl(path string) string {
	u := path
	idx := strings.Index(path, "://")
	if idx <= 0 {
		if f.Host != "" {
			u = f.Host + u
		}
		prefix := "http"
		if f.HTTPS {
			prefix = "https"
		}
		u = prefix + "://" + u
	} else if uu, err := url.Parse(path); err != nil && f.AutoHost && uu.Host != "" {
		f.Host = uu.Host
	}
	return u
}

func (f *Fetcher) makeOtherHeader(req *http.Request) (err error) {
	accept := "application/json, text/javascript, */*; q=0.01"
	//		accept_encoding := "deflate, sdch"
	accept_encoding := "none"
	accept_language := "en-US,en;q=0.8"
	origin := f.makeUrl("")
	user_agent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/27.0.1453.116 Safari/537.36"
	if f.Header.Agent != "" {
		user_agent = f.Header.Agent
	}
	//	x_request_with := "XMLHttpRequest"
	req.Header.Set("Accept", accept)
	req.Header.Set("Accept-Encoding", accept_encoding)
	req.Header.Set("Accept-Language", accept_language)
	req.Header.Set("Origin", origin)
	req.Header.Set("User-Agent", user_agent)
	//	req.Header.Set("X-Requested-With", x_request_with)
	for key, val := range f.Header.Custom {
		req.Header.Set(key, val)
	}
	return
}

func fetchNew(tr http.RoundTripper) (f *Fetcher) {
	f = &Fetcher{}
	newTr := transportNew(tr)
	newTr.AfterReq = func(resp *http.Response, req *http.Request) {
		f.mergeCookie(resp)
		f.Referer = req.URL.String()
	}
	newTr.BeforeReq = func(req *http.Request) {
		f.makeOtherHeader(req)
		f.insertCookie(req)
		f.insertReferer(req)
	}
	f.Client = &http.Client{
		Transport: newTr,
	}
	f.Cache = make(map[string]cache_response)
	return
}

type cache_response struct {
	Resp      *http.Response
	CacheTime int64
}

type transport struct {
	tr        http.RoundTripper
	BeforeReq func(req *http.Request)
	AfterReq  func(resp *http.Response, req *http.Request)
}

func transportNew(tr http.RoundTripper) *transport {
	t := &transport{}
	if tr == nil {
		tr = http.DefaultTransport
	}
	t.tr = tr
	return t
}
func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.BeforeReq(req)
	resp, err = t.tr.RoundTrip(req)
	if err != nil {
		return
	}
	t.AfterReq(resp, req)
	return
}

type custom_header struct {
	Agent  string
	Custom map[string]string
}

func (c *custom_header) Set(field string, value string) {
	c.Custom[field] = value
}
