package template

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"github.com/sail-services/sail-go/com/data/buffer"
	"github.com/sail-services/sail-go/com/data/convert"
	"github.com/sail-services/sail-go/mod/net/service"
	"strings"
	"sync"
)

type (
	Options struct {
		Dir        string   // 文件夹 [tpl]
		Ext        []string // 扩展名 [.html, .tpl]
		Charset    string   // 字符集 [UTF-8]
		DelimLeft  string   // 模版左符号 [nil]
		DelimRight string   // 模版右符号 [nil]
	}
	renderTemplate struct {
		lock sync.RWMutex
		http.ResponseWriter
		tpl     *template.Template
		opt     *Options
		charset string
	}
)

const (
	contentType = "Content-Type"
	contentHTML = "text/html"
)

var (
	buffPool = buffer.New(64)
)

func New(opts ...Options) service.Handler {
	opt := optPrepare(opts)
	charset := service.CharsetGetHeader(opt.Charset)
	tpl := templateCompile(&opt)
	return func(con *service.Context) {
		r := &renderTemplate{
			ResponseWriter: con.Resp,
			tpl:            tpl,
			opt:            &opt,
			charset:        charset,
		}
		con.Ren.RenderTpl = r
	}
}

func templateCompile(opt *Options) *template.Template {
	t := template.New(opt.Dir)
	if len(opt.DelimLeft) != 0 && len(opt.DelimRight) != 0 {
		t.Delims(opt.DelimLeft, opt.DelimRight)
	}
	template.Must(t.Parse(service.ModeGet()))
	if err := filepath.Walk(opt.Dir, func(path string, info os.FileInfo, err error) error {
		r, err := filepath.Rel(opt.Dir, path)
		if err != nil {
			return err
		}
		ext := getExt(r)
		for _, extension := range opt.Ext {
			if ext == extension {
				if err != nil {
					panic(err)
				}
				break
			}
		}
		return nil
	}); err != nil {
		panic("fail to walk templates directory: " + err.Error())
	}
	return t
}

func optPrepare(opts []Options) Options {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.Dir == "" {
		opt.Dir = "tpl"
	}
	if len(opt.Ext) == 0 {
		opt.Ext = []string{".tpl", ".html"}
	}
	if opt.Charset == "" {
		opt.Charset = "UTF-8"
	}
	return opt
}

func getExt(s string) string {
	index := strings.Index(s, ".")
	if index == -1 {
		return ""
	}
	return s[index:]
}

// ========================================================
// renderTemplate
// ========================================================
func (ren *renderTemplate) Tpl(status int, tpl_file string, data interface{}) {
	ren.do(status)
	out := buffPool.Get()
	err := ren.tpl.ExecuteTemplate(out, tpl_file, data)
	if err != nil {
		http.Error(ren, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(ren, out)
	buffPool.Put(out)
}

func (ren *renderTemplate) TplS(status int, tpl_bytes []byte, data interface{}) {
	ren.do(status)
	t := template.Must(ren.tpl.Parse(convert.BToS(tpl_bytes)))
	out := buffPool.Get()
	err := t.Execute(out, data)
	if err != nil {
		http.Error(ren, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(ren, out)
	buffPool.Put(out)
}

func (ren *renderTemplate) do(status int) {
	if service.ModeIsDev() {
		tpl := templateCompile(ren.opt)
		ren.lock.Lock()
		ren.tpl = tpl
		ren.lock.Unlock()
	}
	ren.Header().Set(contentType, contentHTML+ren.charset)
	ren.WriteHeader(status)
}
