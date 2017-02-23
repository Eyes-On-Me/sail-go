package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"reflect"
	"strconv"
)

const (
	_VERSION = "2015.03.05"
	_URL     = "https://github.com/bitly/go-simplejson"
)

type Json struct {
	data interface{}
}

func New() *Json {
	return &Json{
		data: make(map[string]interface{}),
	}
}

func NewB(body []byte) (*Json, error) {
	j := new(Json)
	err := j.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func NewS(body string) (*Json, error) {
	j := new(Json)
	err := j.UnmarshalJSON([]byte(body))
	if err != nil {
		return nil, err
	}
	return j, nil
}

func NewIOReader(r io.Reader) (*Json, error) {
	j := new(Json)
	dec := json.NewDecoder(r)
	dec.UseNumber()
	err := dec.Decode(&j.data)
	return j, err
}

func (j *Json) Interface() interface{} {
	return j.data
}

func (j *Json) Encode() ([]byte, error) {
	return j.MarshalJSON()
}

func (j *Json) EncodePretty() ([]byte, error) {
	return json.MarshalIndent(&j.data, "", "  ")
}

func (j *Json) Set(key string, val interface{}) {
	m, err := j.Map()
	if err != nil {
		return
	}
	m[key] = val
}

func (j *Json) SetPath(branch []string, val interface{}) {
	if len(branch) == 0 {
		j.data = val
		return
	}
	if _, ok := (j.data).(map[string]interface{}); !ok {
		j.data = make(map[string]interface{})
	}
	curr := j.data.(map[string]interface{})
	for i := 0; i < len(branch)-1; i++ {
		b := branch[i]
		if _, ok := curr[b]; !ok {
			n := make(map[string]interface{})
			curr[b] = n
			curr = n
			continue
		}
		if _, ok := curr[b].(map[string]interface{}); !ok {
			n := make(map[string]interface{})
			curr[b] = n
		}
		curr = curr[b].(map[string]interface{})
	}
	curr[branch[len(branch)-1]] = val
}

func (j *Json) Del(key string) {
	m, err := j.Map()
	if err != nil {
		return
	}
	delete(m, key)
}

// js.Get("top_level").Get("dict").Get("value").Int()
func (j *Json) Get(key string) *Json {
	m, err := j.Map()
	if err == nil {
		if val, ok := m[key]; ok {
			return &Json{val}
		}
	}
	return &Json{nil}
}

// js.Get_Path("top_level", "dict")
func (j *Json) GetPath(branch ...string) *Json {
	jin := j
	for _, p := range branch {
		jin = jin.Get(p)
	}
	return jin
}

// js.Get("top_level").Get("array").Get_Index(1).Get("key").Int()
func (j *Json) GetIndex(index int) *Json {
	a, err := j.Array()
	if err == nil {
		if len(a) > index {
			return &Json{a[index]}
		}
	}
	return &Json{nil}
}

// if data, ok := js.Get("top_level").Check_Get("inner"); ok {
//     log.Println(data)
// }
func (j *Json) CheckGet(key string) (*Json, bool) {
	m, err := j.Map()
	if err == nil {
		if val, ok := m[key]; ok {
			return &Json{val}, true
		}
	}
	return nil, false
}

func (j *Json) Map() (map[string]interface{}, error) {
	if m, ok := (j.data).(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

func (j *Json) Array() ([]interface{}, error) {
	if a, ok := (j.data).([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

func (j *Json) F64() (float64, error) {
	switch j.data.(type) {
	case json.Number:
		return j.data.(json.Number).Float64()
	case float32, float64:
		return reflect.ValueOf(j.data).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

func (j *Json) I() (int, error) {
	switch j.data.(type) {
	case json.Number:
		i, err := j.data.(json.Number).Int64()
		return int(i), err
	case float32, float64:
		return int(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

func (j *Json) I64() (int64, error) {
	switch j.data.(type) {
	case json.Number:
		return j.data.(json.Number).Int64()
	case float32, float64:
		return int64(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(j.data).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

func (j *Json) UI64() (uint64, error) {
	switch j.data.(type) {
	case json.Number:
		return strconv.ParseUint(j.data.(json.Number).String(), 10, 64)
	case float32, float64:
		return uint64(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(j.data).Uint(), nil
	}
	return 0, errors.New("invalid value type")
}

func (j *Json) Bool() (bool, error) {
	if s, ok := (j.data).(bool); ok {
		return s, nil
	}
	return false, errors.New("type assertion to bool failed")
}

func (j *Json) S() (string, error) {
	if s, ok := (j.data).(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

func (j *Json) B() ([]byte, error) {
	if s, ok := (j.data).(string); ok {
		return []byte(s), nil
	}
	return nil, errors.New("type assertion to []byte failed")
}

func (j *Json) SArray() ([]string, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}
	retArr := make([]string, 0, len(arr))
	for _, a := range arr {
		if a == nil {
			retArr = append(retArr, "")
			continue
		}
		s, ok := a.(string)
		if !ok {
			return nil, err
		}
		retArr = append(retArr, s)
	}
	return retArr, nil
}

// for i, v := range js.Get("results").Must_Array() {
//     fmt.Println(i, v)
// }
func (j *Json) MustArray(args ...[]interface{}) []interface{} {
	var def []interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Must_Array() received too many arguments %d", len(args))
	}

	a, err := j.Array()
	if err == nil {
		return a
	}

	return def
}

// for k, v := range js.Get("dictionary").MustMap() {
//     fmt.Println(k, v)
// }
func (j *Json) MustMap(args ...map[string]interface{}) map[string]interface{} {
	var def map[string]interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustMap() received too many arguments %d", len(args))
	}

	a, err := j.Map()
	if err == nil {
		return a
	}

	return def
}

// myFunc(js.Get("param1").MustS(), js.Get("optional_param").MustS("my_default"))
func (j *Json) MustS(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustS() received too many arguments %d", len(args))
	}

	s, err := j.S()
	if err == nil {
		return s
	}

	return def
}

// myFunc(js.Get("param1").MustI(), js.Get("optional_param").MustI(5150))
func (j *Json) MustI(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustI() received too many arguments %d", len(args))
	}

	i, err := j.I()
	if err == nil {
		return i
	}

	return def
}

// myFunc(js.Get("param1").MustF64(), js.Get("optional_param").MustF64(5.150))
func (j *Json) MustF64(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustF64() received too many arguments %d", len(args))
	}

	f, err := j.F64()
	if err == nil {
		return f
	}

	return def
}

// myFunc(js.Get("param1").MustBool(), js.Get("optional_param").MustBool(true))
func (j *Json) MustBool(args ...bool) bool {
	var def bool
	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustBool() received too many arguments %d", len(args))
	}
	b, err := j.Bool()
	if err == nil {
		return b
	}
	return def
}

// myFunc(js.Get("param1").MustI64(), js.Get("optional_param").MustI64(5150))
func (j *Json) MustI64(args ...int64) int64 {
	var def int64
	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustI64() received too many arguments %d", len(args))
	}
	i, err := j.I64()
	if err == nil {
		return i
	}
	return def
}

// myFunc(js.Get("param1").MustUI64(), js.Get("optional_param").MustUI64(5150))
func (j *Json) MustUI64(args ...uint64) uint64 {
	var def uint64
	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustUI64() received too many arguments %d", len(args))
	}
	i, err := j.UI64()
	if err == nil {
		return i
	}
	return def
}

// GO
func (j *Json) UnmarshalJSON(p []byte) error {
	dec := json.NewDecoder(bytes.NewBuffer(p))
	dec.UseNumber()
	return dec.Decode(&j.data)
}

// GO
func (j *Json) MarshalJSON() ([]byte, error) {
	return json.Marshal(&j.data)
}
