package service

import (
	"github.com/sail-services/sail-go/com/data/convert"
	"encoding/json"
	"encoding/xml"
	"net/http"
)

type (
	Render struct {
		RenderTpl
		con *Context
	}
	RenderTpl interface {
		http.ResponseWriter
		Tpl(int, string, interface{})
		TplS(int, []byte, interface{})
	}
)

// ========================================================
// Render
// ========================================================
func (ren *Render) Redirect(status int, location string) {
	http.Redirect(ren.con.Resp, ren.con.Req.Request, location, status)
}

func (ren *Render) Data(status int, v []byte) {
	ren.data(status, "application/octet-stream", v)
}

func (ren *Render) S(status int, s string) {
	ren.data(status, "text/plain"+CharsetGetHeader(), []byte(s))
}

func (ren *Render) I(status int, i int) {
	ren.data(status, "text/plain"+CharsetGetHeader(), []byte(convert.IToS(i)))
}

func (ren *Render) B(status int, s []byte) {
	ren.data(status, "text/plain"+CharsetGetHeader(), s)
}

func (ren *Render) HTML(status int, v []byte) {
	ren.data(status, "text/html"+CharsetGetHeader(), v)
}

func (ren *Render) XML(status int, v interface{}) {
	var result []byte
	var err error
	if ModeIsDev() {
		result, err = xml.MarshalIndent(v, "", "  ")
	} else {
		result, err = xml.Marshal(v)
	}
	if err != nil {
		http.Error(ren, err.Error(), 500)
		return
	}
	ren.con.Resp.Header().Set("Content-Type", "text/xml"+CharsetGetHeader())
	ren.con.Resp.WriteHeader(status)
	ren.con.Resp.Write(result)
}

func (ren *Render) JSON(status int, v interface{}) {
	var result []byte
	var err error
	if ModeIsDev() {
		result, err = json.MarshalIndent(v, "", "  ")
	} else {
		result, err = json.Marshal(v)
	}
	if err != nil {
		http.Error(ren, err.Error(), 500)
		return
	}
	ren.con.Resp.Header().Set("Content-Type", "application/json"+CharsetGetHeader())
	ren.con.Resp.WriteHeader(status)
	ren.con.Resp.Write(result)
}

func (ren *Render) Tpl(status int, tpl_file string, data ...interface{}) {
	if ren.RenderTpl == nil {
		ren.con.Log.Panicln("Not Have Template Module")
	}
	ren.con.Resp.Header().Set("Cache-Control", "no-cache")
	if len(data) == 0 {
		ren.RenderTpl.Tpl(status, tpl_file, ren.con.Var)
	} else {
		ren.RenderTpl.Tpl(status, tpl_file, data[0])
	}
}

func (ren *Render) TplS(status int, tpl_bytes []byte, data ...interface{}) {
	if ren.RenderTpl == nil {
		ren.con.Log.Panicln("Not Have Template Module")
	}
	ren.con.Resp.Header().Set("Cache-Control", "no-cache")
	if len(data) == 0 {
		ren.RenderTpl.TplS(status, tpl_bytes, ren.con.Var)
	} else {
		ren.RenderTpl.TplS(status, tpl_bytes, data[0])
	}
}

func (ren *Render) data(status int, content_type string, v []byte) {
	if ren.con.Resp.Header().Get("Content-Type") == "" {
		ren.con.Resp.Header().Set("Content-Type", content_type)
	}
	ren.con.Resp.WriteHeader(status)
	ren.con.Resp.Write(v)
}
