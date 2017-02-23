package pongo2

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"github.com/sail-services/sail-go/com/data/convert"
	"github.com/sail-services/sail-go/mod/net/service"
	"strings"
	"sync"

	"github.com/flosch/pongo2"
)

type (
	pongo2Render struct {
		lock sync.RWMutex
		http.ResponseWriter
		opt     *Options
		tpls    map[string]*pongo2.Template
		charset string
	}
	Options struct {
		Dir     string   // 文件夹 [pongo2]
		Ext     []string // 扩展名 [.html, .tpl]
		Charset string   // 字符集 [UTF-8]
	}
)

const (
	_CONTENT_TYPE = "Content-Type"
	_CONTENT_HTML = "text/html"
)

func New(opts ...Options) service.Handler {
	opt := optPrepare(opts)
	charset := service.CharsetGetHeader(opt.Charset)
	tpls := pongo2Compile(&opt)
	return func(con *service.Context) {
		ren := &pongo2Render{
			ResponseWriter: con.Resp,
			opt:            &opt,
			charset:        charset,
			tpls:           tpls,
		}
		con.Ren.RenderTpl = ren
	}
}

// ========================================================
// pongo2Render
// ========================================================
func (ren *pongo2Render) Tpl(status int, tpl_file string, data interface{}) {
	ren.do(status)
	err := ren.tpls[tpl_file].ExecuteWriter(dataToPongo2Context(data), ren)
	if err != nil {
		http.Error(ren, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ren *pongo2Render) TplS(status int, tpl_bytes []byte, data interface{}) {
	ren.do(status)
	t, err := pongo2.FromString(convert.BToS(tpl_bytes))
	if err != nil {
		http.Error(ren, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteWriter(dataToPongo2Context(data), ren)
	if err != nil {
		http.Error(ren, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ren *pongo2Render) do(status int) {
	if service.ModeIsDev() {
		tpl_map := pongo2Compile(ren.opt)
		ren.lock.Lock()
		ren.tpls = tpl_map
		ren.lock.Unlock()
	}
	ren.Header().Set(_CONTENT_TYPE, _CONTENT_HTML+ren.charset)
	ren.WriteHeader(status)
}

// --------------------------------------------------------
// FUN
// --------------------------------------------------------
func pongo2Compile(opt *Options) map[string]*pongo2.Template {
	tpl_map := make(map[string]*pongo2.Template)
	if err := filepath.Walk(opt.Dir, func(path string, info os.FileInfo, err error) error {
		r, err := filepath.Rel(opt.Dir, path)
		if err != nil {
			return err
		}
		ext := getExt(r)
		for _, extension := range opt.Ext {
			if ext == extension {
				name := (r[0 : len(r)-len(ext)])
				t, err := pongo2.FromFile(path)
				if err != nil {
					panic(fmt.Errorf("\"%s\": %v", path, err))
				}
				tpl_map[strings.Replace(name, "\\", "/", -1)] = t
				break
			}
		}
		return nil
	}); err != nil {
		panic("fail to walk templates directory: " + err.Error())
	}
	return tpl_map
}

func optPrepare(opts []Options) Options {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.Dir == "" {
		opt.Dir = "pongo2"
	}
	if len(opt.Ext) == 0 {
		opt.Ext = []string{".tpl", ".html"}
	}
	if opt.Charset == "" {
		opt.Charset = "UTF-8"
	}
	return opt
}

func dataToPongo2Context(data interface{}) pongo2.Context {
	return pongo2.Context(data.(map[string]interface{}))
}

func getExt(s string) string {
	index := strings.Index(s, ".")
	if index == -1 {
		return ""
	}
	return s[index:]
}
