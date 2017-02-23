package service

import (
	"net/http"
	"os"
	"sync"

	"github.com/sail-services/sail-go/com/data/convert"
	"github.com/sail-services/sail-go/mod/data/log"
)

type (
	Service struct {
		Rou  Router
		Log  *log.Log
		pool sync.Pool
		mods []Handler
	}
	Router interface {
		http.Handler
		init(ser *Service)
		Group(rpath string, function func(), hds ...Handler)
		Get(rpath string, hds ...Handler)
		Post(rpath string, hds ...Handler)
		Put(rpath string, hds ...Handler)
		Delete(rpath string, hds ...Handler)
		Patch(rpath string, hds ...Handler)
		Options(rpath string, hds ...Handler)
		Head(rpath string, hds ...Handler)
		Link(rpath string, hds ...Handler)
		Unlink(rpath string, hds ...Handler)
		Any(rpath string, hds ...Handler)
		File(rpath, fpath string)
		NotFound(hds ...Handler)
	}
	Handler func(*Context)
)

const (
	_VERSION   = "0.4 / 2015.12.1"
	_ENV       = "SERVICE"
	_HOST      = "0.0.0.0"
	_PORT      = 8080
	_E404      = "404 - Page Not Found"
	_CHARSET   = "UTF-8"
	_PATH_ROOT = "/"
	_MODE_DEV  = iota
	_MODE_RELEASE
	_FORM_MEMORY = int64(1024 * 1024 * 10)
)

var (
	_env        string = _ENV
	_mode       int    = _MODE_DEV
	_charset    string = _CHARSET
	_path       string
	_secret_key string
)

// ========================================================
// Service
// ========================================================
func New(l *log.Log) *Service {
	_path, _ = os.Getwd()
	if len(os.Getenv(_env)) != 0 {
		ModeSet(os.Getenv(_env))
	}
	ser := &Service{}
	ser.Log = l
	ser.Rou = new(routerPro)
	ser.Rou.init(ser)
	ser.mods = nil
	ser.pool.New = func() interface{} {
		con := &Context{Log: ser.Log}
		con.Ren = &Render{con: con}
		con.Opt = &contextOpts{Log: true}
		con.resp.con = con
		con.Resp = &con.resp
		con.Var = make(map[string]interface{})
		return con
	}
	return ser
}

// 空: 0.0.0.0:8080
// string: IP
// int: Port
// string, int: IP, Port
// string, int, string, string: IP, Prot, Cert, Key
func (ser *Service) Run(args ...interface{}) {
	host, port := _HOST, _PORT
	var cert, key string
	switch len(args) {
	case 1:
		switch arg := args[0].(type) {
		case string:
			host = arg
		case int:
			port = arg
		}
	case 2:
		if arg, ok := args[0].(string); ok {
			host = arg
		}
		if arg, ok := args[1].(int); ok {
			port = arg
		}
	case 4:
		if arg, ok := args[0].(string); ok {
			host = arg
		}
		if arg, ok := args[1].(int); ok {
			port = arg
		}
		if arg, ok := args[2].(string); ok {
			cert = arg
		}
		if arg, ok := args[3].(string); ok {
			key = arg
		}
	}
	addr := host + ":" + convert.ToS(port)
	if ModeIsDev() {
		ser.Log.Infof("RUN %v\n", addr)
	}
	if len(args) == 4 {
		ser.Log.Fatalln(http.ListenAndServeTLS(addr, cert, key, ser.Rou))
	} else {
		ser.Log.Fatalln(http.ListenAndServe(addr, ser.Rou))
	}
}

func (ser *Service) Module(hds ...Handler) {
	ser.mods = append(ser.mods, hds...)
}

func (ser *Service) modsCombine(hds []Handler) []Handler {
	final_size := len(ser.mods) + len(hds)
	merged_hds := make([]Handler, 0, final_size)
	merged_hds = append(merged_hds, ser.mods...)
	return append(merged_hds, hds...)
}

func (ser *Service) contextNew(resp http.ResponseWriter, req *http.Request, hds []Handler) *Context {
	con := ser.pool.Get().(*Context)
	con.resp.reset(resp, con)
	con.Resp = &con.resp
	con.Req.Request = req
	con.Req.con = con
	con.Ren.con = con
	con.Opt.Log = true
	con.Opt.Stop = false
	con.Var = make(map[string]interface{})
	con.data = make(map[string]interface{})
	con.handlers = hds
	con.handlerIndex = -1
	return con
}

// --------------------------------------------------------
// Charset
// --------------------------------------------------------
func CharsetSet(c string) {
	_charset = c
}

func CharsetGet() string {
	return _charset
}

func CharsetGetHeader(charset ...string) string {
	if len(charset) != 0 {
		return "; charset=" + charset[0]
	}
	return "; charset=" + _charset
}

// --------------------------------------------------------
// Secret Key
// --------------------------------------------------------
func SecretKeySet(key string) {
	_secret_key = key
}

func SecretKeyGet() string {
	return _secret_key
}

// --------------------------------------------------------
// Mode
// --------------------------------------------------------
func ModeSet(value string) {
	switch value {
	case "release":
		_mode = _MODE_RELEASE
	case "dev":
	default:
		_mode = _MODE_DEV
	}
}

func ModeGet() string {
	switch _mode {
	case _MODE_DEV:
		return "dev"
	case _MODE_RELEASE:
		return "release"
	}
	return "unknown"
}

func ModeIsDev() bool {
	return _mode == _MODE_DEV
}

// --------------------------------------------------------
// Other
// --------------------------------------------------------
func EnvSet(e string) {
	_env = e
}

func PathGet() string {
	return _path
}
