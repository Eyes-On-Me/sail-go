package static

import (
	"net/http"
	"path"
	"path/filepath"
	"github.com/sail-services/sail-go/mod/net/service"
	"strings"
)

type (
	Options struct {
		Prefix    string          // 前缀路径 [/]
		Dir       string          // 文件夹 [static]
		ShowLog   bool            // 显示日志 [false]
		FS        http.FileSystem // 文件系统接口 [可定义]
		IndexFile string          // 默认文件 [index.html]
	}
	staticFS struct {
		dir *http.Dir
	}
)

func New(opts ...Options) service.Handler {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	opt = optPrepare(opt)
	return func(con *service.Context) {
		if staticHandler(con, opt) {
			con.Opt.Stop = true
			return
		}
	}
}

func News(opts []Options) service.Handler {
	if len(opts) == 0 {
		panic("[Static] no static directory is given")
	}
	popts := make([]Options, len(opts))
	for i, opt := range opts {
		popts[i] = optPrepare(opt)
	}
	return func(con *service.Context) {
		for i := range popts {
			if staticHandler(con, popts[i]) {
				con.Opt.Stop = true
				return
			}
		}
	}
}

func staticHandler(con *service.Context, opt Options) bool {
	if con.Req.Method != "GET" && con.Req.Method != "HEAD" {
		return false
	}
	file := con.Req.URL.Path
	if opt.Prefix != "" {
		if !strings.HasPrefix(file, opt.Prefix) {
			return false
		}
		file = file[len(opt.Prefix):]
		if file != "" && file[0] != '/' {
			return false
		}
	}
	f, err := opt.FS.Open(file)
	if err != nil {
		return false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return true
	}
	if fi.IsDir() {
		if !strings.HasSuffix(con.Req.URL.Path, "/") {
			http.Redirect(con.Resp, con.Req.Request, con.Req.URL.Path+"/", http.StatusFound)
			return true
		}
		file = path.Join(file, opt.IndexFile)
		f, err = opt.FS.Open(file)
		if err != nil {
			return false
		}
		defer f.Close()
		fi, err = f.Stat()
		if err != nil || fi.IsDir() {
			return true
		}
	}
	if opt.ShowLog && service.ModeIsDev() {
		con.Log.Println("[Static] " + file)
	}
	con.Opt.Log = false
	http.ServeContent(con.Resp, con.Req.Request, file, fi.ModTime(), f)
	return true
}

func optPrepare(opt Options) Options {
	if opt.Dir == "" {
		opt.Dir = "static"
	}
	if opt.IndexFile == "" {
		opt.IndexFile = "index.html"
	}
	if opt.Prefix != "" {
		if opt.Prefix[0] != '/' {
			opt.Prefix = "/" + opt.Prefix
		}
		opt.Prefix = strings.TrimRight(opt.Prefix, "/")
	}
	if opt.FS == nil {
		opt.FS = staticFSNew(opt.Dir)
	}
	return opt
}

// ========================================================
// staticFS
// ========================================================
func staticFSNew(directory string) staticFS {
	if !filepath.IsAbs(directory) {
		directory = filepath.Join(service.PathGet(), directory)
	}
	dir := http.Dir(directory)
	return staticFS{&dir}
}

func (fs staticFS) Open(name string) (http.File, error) {
	return fs.dir.Open(name)
}
