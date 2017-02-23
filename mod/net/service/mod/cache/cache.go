package cache

import (
	"fmt"
	"github.com/sail-services/sail-go/mod/net/service"
)

type (
	Options struct {
		Adapter    string // 适配器 [nil]
		Conn       string // 适配器数据 [nil]
		Interval   int    // 数据回收间隔 [60]
		OccupyMode bool   // Redis: Occupy entire database [false]
	}
	Cache interface {
		Get(key string) interface{}                           // 获取
		Put(key string, val interface{}, timeout int64) error // 追加
		Delete(key string) error                              // 删除
		Incr(key string) error                                // 储存的数字值 +1
		Decr(key string) error                                // 储存的数字值 -1
		IsExist(key string) bool                              // Key 是否存在
		Flush() error                                         // 清空整个数据库
		StartAndGC(opt Options) error
	}
)

const (
	_DATA_CACHE = "_DATA_CACHE"
)

var (
	adapters = make(map[string]Cache)
)

func New(opts ...Options) service.Handler {
	opt := optPrepare(opts)
	cache, err := cacheHandler(opt.Adapter, opt)
	if err != nil {
		panic(err)
	}
	return cache
}

func Register(name string, adapter Cache) {
	if adapter == nil {
		panic("cache: Register adapter is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("cache: Register called twice for adapter " + name)
	}
	adapters[name] = adapter
}

func DataGetCache(con *service.Context) Cache {
	return con.DataMustGet(_DATA_CACHE).(Cache)
}

func cacheHandler(adapterName string, config Options) (service.Handler, error) {
	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("cache: unknown adapter name %q (forgot to import?)", adapterName)
	}
	if err := adapter.StartAndGC(config); err != nil {
		return nil, err
	}
	return func(con *service.Context) {
		con.DataSet(_DATA_CACHE, adapter)
	}, nil
}

func optPrepare(opts []Options) Options {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.Adapter == "" {
		panic("no adapter string")
	}
	if opt.Adapter == "memory" {
		if opt.Interval == 0 {
			opt.Interval = 60
		}
	} else {
		if opt.Conn == "" {
			panic("no connection string is given for non-memory cache adapter")
		}
	}
	return opt
}
