package i18n

import (
	"errors"
	"fmt"
	"github.com/sail-services/sail-go/mod/data/ini"
	"reflect"
	"strings"
)

var (
	ErrLangAlreadyExist = errors.New("Lang already exists")
	locales             = &localeStore{store: make(map[string]*locale)}
)

type locale struct {
	id       int
	lang     string
	langDesc string
	message  *ini.File
}

type localeStore struct {
	langs       []string
	langDescs   []string
	store       map[string]*locale
	defaultLang string
}

type Langs struct {
	Lang string
}

func (l Langs) P(format string, args ...interface{}) string {
	return P(l.Lang, format, args...)
}

func (l Langs) Index() int {
	return IndexLang(l.Lang)
}

func LangsReload(langs ...string) error {
	return locales.Reload(langs...)
}

func (d *localeStore) Get(lang, section, format string) (string, bool) {
	if locale, ok := d.store[lang]; ok {
		if key, err := locale.message.Section(section).GetKey(format); err == nil {
			return key.Value(), true
		}
	}

	if len(d.defaultLang) > 0 && lang != d.defaultLang {
		return d.Get(d.defaultLang, section, format)
	}

	return "", false
}

func (d *localeStore) Add(lc *locale) bool {
	if _, ok := d.store[lc.lang]; ok {
		return false
	}

	lc.id = len(d.langs)
	d.langs = append(d.langs, lc.lang)
	d.langDescs = append(d.langDescs, lc.langDesc)
	d.store[lc.lang] = lc

	return true
}

func (d *localeStore) Reload(langs ...string) (err error) {
	if len(langs) == 0 {
		for _, lc := range d.store {
			if err = lc.message.Reload(); err != nil {
				return err
			}
		}
	} else {
		for _, lang := range langs {
			if lc, ok := d.store[lang]; ok {
				if err = lc.message.Reload(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func SetDefaultLang(lang string) {
	locales.defaultLang = lang
}

func ReloadLangs(langs ...string) error {
	return locales.Reload(langs...)
}

func Count() int {
	return len(locales.langs)
}

func ListLangs() []string {
	langs := make([]string, len(locales.langs))
	copy(langs, locales.langs)
	return langs
}

func ListLangDescs() []string {
	langDescs := make([]string, len(locales.langDescs))
	copy(langDescs, locales.langDescs)
	return langDescs
}

func IsExist(lang string) bool {
	_, ok := locales.store[lang]
	return ok
}

func IndexLang(lang string) int {
	if lc, ok := locales.store[lang]; ok {
		return lc.id
	}
	return -1
}

func GetLangByIndex(index int) string {
	if index < 0 || index >= len(locales.langs) {
		return ""
	}
	return locales.langs[index]
}

func GetDescriptionByIndex(index int) string {
	if index < 0 || index >= len(locales.langDescs) {
		return ""
	}

	return locales.langDescs[index]
}

func GetDescriptionByLang(lang string) string {
	return GetDescriptionByIndex(IndexLang(lang))
}

func SetMessageWithDesc(lang, langDesc string, localeFile interface{}, otherLocaleFiles ...interface{}) error {
	message, err := ini.Load(localeFile, otherLocaleFiles...)
	if err == nil {
		message.BlockMode = false
		lc := new(locale)
		lc.lang = lang
		lc.langDesc = langDesc
		lc.message = message

		if locales.Add(lc) == false {
			return ErrLangAlreadyExist
		}
	}
	return err
}

func SetMessage(lang string, localeFile interface{}, otherLocaleFiles ...interface{}) error {
	return SetMessageWithDesc(lang, lang, localeFile, otherLocaleFiles...)
}

func P(lang, format string, args ...interface{}) string {
	var section string
	parts := strings.SplitN(format, ".", 2)
	if len(parts) == 2 {
		section = parts[0]
		format = parts[1]
	}

	value, ok := locales.Get(lang, section, format)
	if ok {
		format = value
	}

	if len(args) > 0 {
		params := make([]interface{}, 0, len(args))
		for _, arg := range args {
			if arg != nil {
				val := reflect.ValueOf(arg)
				if val.Kind() == reflect.Slice {
					for i := 0; i < val.Len(); i++ {
						params = append(params, val.Index(i).Interface())
					}
				} else {
					params = append(params, arg)
				}
			}
		}
		return fmt.Sprintf(format, params...)
	}
	return format
}
