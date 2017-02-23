package i18n

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/sail-services/sail-go/mod/net/service"
	"github.com/sail-services/sail-go/mod/net/service/mod/i18n/i18n"
)

type (
	Options struct {
		Langs       []string // 语言 [nil]
		Names       []string // 语言名 [nil]
		Dir         string   // 目录 [lang]
		Tmpl        string   // 模版中的变量名 [lang]
		Url         bool     // 使用URL设定语言 [false]
		UrlVar      string   // URL设定语言 [lang] (/?lang=zh-CN)
		Cookie      string   // Cookie设定语言 [lang]
		format      string
		subUrl      string
		DefaultLang string
	}
	i18nLangs struct {
		i18n.Langs
	}
	LangType struct {
		Lang string
		Name string
	}
)

func New(opts ...Options) service.Handler {
	opt := optPrepare(opts)
	initLocales(opt)
	return func(con *service.Context) {
		if service.ModeIsDev() {
			i18n.LangsReload()
		}
		isNeedRedir := false
		hasCookie := false
		lang := con.Req.FormGet(opt.UrlVar)
		if lang == "" {
			lang = con.Req.CookieGet(opt.Cookie)
			hasCookie = true
		} else {
			isNeedRedir = true
		}
		if !i18n.IsExist(lang) {
			lang = ""
			isNeedRedir = false
			hasCookie = false
		}
		if lang == "" {
			al := con.Req.Header.Get("Accept-Language")
			if len(al) > 4 {
				al = al[:5]
				if i18n.IsExist(al) {
					lang = al
				}
			}
		}
		if lang == "" {
			lang = i18n.GetLangByIndex(0)
			isNeedRedir = false
		}
		curLang := LangType{
			Lang: lang,
		}
		if !hasCookie {
			con.Resp.CookieSet(opt.Cookie, curLang.Lang, 1<<31-1, "/"+strings.TrimPrefix(opt.subUrl, "/"))
		}
		restLangs := make([]LangType, 0, i18n.Count()-1)
		langs := i18n.ListLangs()
		names := i18n.ListLangDescs()
		for i, v := range langs {
			if lang != v {
				restLangs = append(restLangs, LangType{v, names[i]})
			} else {
				curLang.Name = names[i]
			}
		}
		lce := i18nLangs{i18n.Langs{lang}}
		con.Lang = lce
		con.Var[opt.Tmpl] = lce
		con.Var["P"] = i18n.P
		con.Var["LANG"] = lce.Lang
		con.Var["LANG_NAME"] = curLang.Name
		con.Var["LANGS"] = append([]LangType{curLang}, restLangs...)
		con.Var["LANGS_REST"] = restLangs
		if opt.Url && isNeedRedir {
			con.Ren.Redirect(http.StatusFound, opt.subUrl+con.Req.RequestURI[:strings.Index(con.Req.RequestURI, "?")])
		}
	}
}

func optPrepare(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}
	opt.subUrl = strings.TrimSuffix(opt.subUrl, "/")
	if len(opt.Langs) == 0 {
		panic("opt.Langs is nil")
	}
	if len(opt.Names) == 0 {
		panic("opt.Names is nil")
	}
	if len(opt.Langs) != len(opt.Names) {
		panic("len(opt.Langs) not equ len(opt.Names)")
	}
	i18n.SetDefaultLang(opt.DefaultLang)
	if opt.Cookie == "" {
		opt.Cookie = "LANG"
	}
	if opt.Dir == "" {
		opt.Dir = "lang"
	}
	if opt.format == "" {
		opt.format = "%s.ini"
	}
	if opt.UrlVar == "" {
		opt.UrlVar = "lang"
	}
	if opt.Tmpl == "" {
		opt.Tmpl = "lang"
	}
	return opt
}

func initLocales(opt Options) {
	for i, lang := range opt.Langs {
		fname := fmt.Sprintf(opt.format, lang)
		err := i18n.SetMessageWithDesc(lang, opt.Names[i], path.Join(opt.Dir, fname))
		if err != nil && err != i18n.ErrLangAlreadyExist {
			panic(fmt.Errorf("fail to set message file(%s): %v", lang, err))
		} else if err != nil {
			panic(err)
		}
	}
}

// ========================================================
// i18nLangs
// ========================================================
func (l i18nLangs) Language() string {
	return l.Lang
}
