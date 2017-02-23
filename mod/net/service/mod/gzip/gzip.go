package gzip

import (
	"compress/gzip"
	"github.com/sail-services/sail-go/mod/net/service"
	"strings"
)

type (
	gzipResponse struct {
		service.Response
		w *gzip.Writer
	}
)

const (
	LEVEL_BEST        = gzip.BestCompression
	LEVEL_BEST_SPEED  = gzip.BestSpeed
	LEVEL_DEFAULT     = gzip.DefaultCompression
	LEVEL_NO          = gzip.NoCompression
	_CONTENT_LENGTH   = "Content-Length"
	_CONTENT_ENCODING = "Content-Encoding"
	_ACCEPT_ENCODING  = "Accept-Encoding"
	_ENCODING_GZIP    = "gzip"
)

func New(level int) service.Handler {
	return func(con *service.Context) {
		if strings.Contains(con.Req.Header.Get("Connection"), "Upgrade") || !strings.Contains(con.Req.Header.Get(_ACCEPT_ENCODING), _ENCODING_GZIP) {
			con.Next()
			return
		}
		gz, err := gzip.NewWriterLevel(con.Resp, level)
		if err != nil {
			con.Next()
			return
		}
		defer gz.Close()
		hd := con.Resp.Header()
		hd.Set(_CONTENT_ENCODING, _ENCODING_GZIP)
		hd.Set("Vary", _ACCEPT_ENCODING)
		gr := &gzipResponse{con.Resp, gz}
		con.Resp = gr
		con.Next()
		gr.Header().Del(_CONTENT_LENGTH)
	}
}

// --------------------------------------------------------
// gzipResponse - GO
// --------------------------------------------------------
func (gr *gzipResponse) Write(p []byte) (int, error) {
	return gr.w.Write(p)
}
