package log

import (
	"github.com/sail-services/sail-go/mod/net/service"
	"time"
)

func New() service.Handler {
	return func(con *service.Context) {
		start := time.Now()
		con.Next()
		if con.Opt.Log {
			con.Log.Infof(
				"%v %v %v (%v)\n",
				con.Req.Method,
				con.Resp.Status(),
				con.Req.RequestURI,
				time.Now().Sub(start),
			)
		}
	}
}
