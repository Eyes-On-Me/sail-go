package recovery

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/sail-services/sail-go/mod/net/service"
	"runtime"
)

const (
	panicHTML = `<html>
<head><title>PANIC: %s</title>
<meta charset="utf-8" />
<style type="text/css">
html, body {
	font-family: "Open sans", "Lucida Grande", Helvetica, sans-serif;
	background-color: #e4e4e6;
	margin: 0px;
}
h1 {
	color: #ea5343;
	background-color: #fff;
	padding: 20px;
}
pre {
    font-family: "Consolas", "Bitstream Vera Sans Mono", "Courier New", Courier, monospace;
    font-size: 12px;
	color: #333;
	margin: 20px;
	padding: 20px;
	background-color: #fff;
	white-space: pre-wrap;       /* css-3 */
	white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
	white-space: -pre-wrap;      /* Opera 4-6 */
	white-space: -o-pre-wrap;    /* Opera 7 */
	word-wrap: break-word;       /* Internet Explorer 5.5+ */
}
</style>
</head><body>
<h1>PANIC</h1>
<pre>%s</pre>
<pre>%s</pre>
</body>
</html>`
)

var (
	dunno      = []byte("???")
	center_dot = []byte("·")
	dot        = []byte(".")
	slash      = []byte("/")
)

func New() service.Handler {
	return func(con *service.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := stack(3)
				con.Log.Fatalf("<PANIC> %s\n%s\n", stack, err)
				con.Resp.WriteHeader(http.StatusInternalServerError)
				if service.ModeIsDev() {
					con.Resp.Header().Set("Content-Type", "text/html")
					body := []byte(fmt.Sprintf(panicHTML, err, err, stack))
					con.Resp.Write(body)
				}
			}
		}()
		con.Next()
	}
}

func stack(skip int) []byte {
	buf := new(bytes.Buffer)
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

func source(lines [][]byte, n int) []byte {
	n--
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, center_dot, dot, -1)
	return name
}
