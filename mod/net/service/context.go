package service

import (
	"github.com/sail-services/sail-go/mod/data/log"
	"math"
)

type (
	Context struct {
		Log  *log.Log
		Opt  *contextOpts
		Ren  *Render
		Req  Request
		Resp Response
		Var  map[string]interface{}
		Lang interface {
			Language() string
			P(string, ...interface{}) string
		}
		data         map[string]interface{}
		handlers     []Handler
		handlerIndex int8
		resp         response
	}
	contextOpts struct {
		Stop bool
		Log  bool
	}
)

// Context
func (con *Context) Next() {
	con.handlerIndex++
	s := int8(len(con.handlers))
	for ; con.handlerIndex < s; con.handlerIndex++ {
		con.handlers[con.handlerIndex](con)
		if con.Opt.Stop {
			return
		}
	}
}

func (con *Context) Abort(code int) {
	con.resp.WriteHeader(code)
	con.handlerIndex = math.MaxInt8 / 2
}

// Context - Data
func (con *Context) DataSet(key string, item interface{}) {
	if con.data == nil {
		con.data = make(map[string]interface{})
	}
	con.data[key] = item
}

func (con *Context) DataGet(key string) (value interface{}, success bool) {
	if con.data != nil {
		value, success = con.data[key]
	}
	return
}

func (con *Context) DataMustGet(key string) interface{} {
	if value, exists := con.DataGet(key); exists {
		return value
	} else {
		con.Log.Fatalf("Key %s does not exist\n", key)
	}
	return nil
}
