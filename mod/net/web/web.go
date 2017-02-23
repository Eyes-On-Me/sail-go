package web

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/sail-services/sail-go/com/data/convert"
	"github.com/sail-services/sail-go/com/data/crypt/aes"
	rr_log "github.com/sail-services/sail-go/mod/data/log"
	"github.com/sail-services/sail-go/mod/net/service"
	"github.com/sail-services/sail-go/mod/net/service/mod/csrf"
	"github.com/sail-services/sail-go/mod/net/service/mod/gzip"
	"github.com/sail-services/sail-go/mod/net/service/mod/i18n"
	"github.com/sail-services/sail-go/mod/net/service/mod/log"
	"github.com/sail-services/sail-go/mod/net/service/mod/pongo2"
	"github.com/sail-services/sail-go/mod/net/service/mod/recovery"
	"github.com/sail-services/sail-go/mod/net/service/mod/session"
	_ "github.com/sail-services/sail-go/mod/net/service/mod/session/memory"
	"github.com/sail-services/sail-go/mod/net/service/mod/static"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type (
	Web struct {
		Ser     *service.Service
		DB      *gorm.DB
		Log     *rr_log.Log
		Project *Project
		Opt     *Opt
		Base    *Base
		Pro     *Pro
	}
	Project struct {
		Name    string
		Version string
		Url     string
	}
	Opt struct {
		DbLogDisabled bool
	}
	Base struct {
		Port            int
		SecretKey       string
		PathLangs       string
		PathTemplate    string
		PathRootDev     string
		PathRootRelease string
		DefaultLang     string
		I18nNames       []string
		I18nLangs       []string
	}
	Pro struct {
		PathLog     string
		PathStatics [][]string
		StaticFiles [][]string
		CSRF        string
		ConnDb      string
		ConnSession string
	}
)

const (
	_DB_TYPE      = "mysql"
	_SESSION_TYPE = "memory"
)

var (
	root_ string
)

func (web *Web) Init() {
	if web.Project.Name == "" ||
		web.Project.Version == "" ||
		web.Project.Url == "" ||
		web.Base.Port == 0 ||
		web.Base.PathRootDev == "" ||
		web.Base.PathRootRelease == "" ||
		web.Base.PathTemplate == "" ||
		web.Base.PathLangs == "" ||
		web.Base.DefaultLang == "" ||
		len(web.Base.I18nNames) == 0 ||
		len(web.Base.I18nLangs) == 0 {
		panic("Please Set Base Data")
	}
	service.EnvSet(strings.ToUpper(strings.Replace(web.Project.Name, " ", "_", -1)))
	if len(web.Base.SecretKey) != 0 {
		service.SecretKeySet(web.Base.SecretKey)
	}
	if service.ModeIsDev() && web.Log == nil {
		web.Log = rr_log.New(os.Stdout, rr_log.LEVEL_INFO, rr_log.DATA_BASIC)
	} else if web.Log == nil {
		web.Log = rr_log.NewFile(web.Pro.PathLog, rr_log.LEVEL_INFO, rr_log.DATA_ALL)
	}
	web.Ser = service.New(web.Log)
	if service.ModeIsDev() {
		root_ = web.Base.PathRootDev
	} else {
		root_ = web.Base.PathRootRelease
	}
	if len(web.Pro.ConnDb) != 0 {
		web.DB, _ = gorm.Open(_DB_TYPE, web.Pro.ConnDb)
		err := web.DB.DB().Ping()
		if err != nil {
			web.Ser.Log.Fatalln("Conn Database Error")
		}
		if service.ModeIsDev() && !web.Opt.DbLogDisabled {
			web.DB.LogMode(true)
		}
	}
	if len(web.Pro.StaticFiles) != 0 {
		for _, file := range web.Pro.StaticFiles {
			web.Ser.Rou.File(file[0], root_+file[1])
		}
	}
	if len(web.Pro.PathStatics) != 0 {
		opts := make([]static.Options, len(web.Pro.PathStatics))
		for i, s := range web.Pro.PathStatics {
			opts[i] = static.Options{
				Prefix: s[0],
				Dir:    root_ + s[1],
			}
		}
		web.Ser.Module(static.News(opts))
	}
	if service.ModeIsDev() {
		web.Ser.Module(log.New())
	} else {
		web.Ser.Module(gzip.New(gzip.LEVEL_DEFAULT))
	}
	web.Ser.Module(recovery.New())
	web.Ser.Module(pongo2.New(pongo2.Options{
		Dir: root_ + web.Base.PathTemplate,
	}))
	web.Ser.Module(i18n.New(i18n.Options{
		Dir:         root_ + web.Base.PathLangs,
		Langs:       web.Base.I18nLangs,
		Names:       web.Base.I18nNames,
		DefaultLang: web.Base.DefaultLang,
	}))
	if len(web.Pro.ConnSession) != 0 || len(web.Pro.CSRF) != 0 {
		web.Ser.Module(session.New(session.Options{
			Adapter:    _SESSION_TYPE,
			Conn:       web.Pro.ConnSession,
			Gclifetime: 60 * 60,
		}))
	}
	if len(web.Pro.CSRF) != 0 {
		web.Ser.Module(csrf.New(csrf.Options{
			SecretKey: web.Base.SecretKey,
			Session:   web.Pro.CSRF,
		}))
	}
}

func (web *Web) Run() {
	if service.ModeIsDev() {
		web.Ser.Run(web.Base.Port)
	} else {
		web.Ser.Run("127.0.0.1", web.Base.Port)
	}
}

func (web *Web) Tpl(tpl string, display int, is_pjax_page bool, con *service.Context) {
	con.Var["project"] = web.Project.Name
	con.Var["version"] = web.Project.Version
	con.Var["title"] = con.Lang.P(tpl + ".title")
	con.Var["tpl"] = tpl
	con.Var["display"] = display
	con.Var["is_dev"] = service.ModeIsDev()
	if len(web.Pro.CSRF) != 0 {
		cs := csrf.DataCSRFGet(con)
		con.Var["token"] = cs.TokenGet()
	}
	if is_pjax_page && len(con.Req.Header.Get("PJAX")) > 0 {
		con.Var["pjax"] = true
	} else {
		con.Var["pjax"] = false
	}
	status := 200
	if tpl == "404" {
		status = 404
	}
	con.Ren.Tpl(status, tpl)
}

func (web *Web) EnVar(v string) (ev string) {
	if web.Base.SecretKey == "" {
		web.Ser.Log.Fatalln("Not Set Secret Key")
	}
	crypt, _ := aes.Encrypt(convert.SToB(v), convert.SToB(web.Base.SecretKey))
	return base64.StdEncoding.EncodeToString(crypt)
}

func (web *Web) DeVar(ev string) (v string) {
	if web.Base.SecretKey == "" {
		web.Ser.Log.Fatalln("Not Set Secret Key")
	}
	ec, _ := base64.StdEncoding.DecodeString(ev)
	crypt, _ := aes.Decrypt(ec, convert.SToB(web.Base.SecretKey))
	return convert.BToS(crypt)
}
