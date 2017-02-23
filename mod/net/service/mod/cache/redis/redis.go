package cache

import (
	"fmt"
	"github.com/sail-services/sail-go/foundation/component/data/convert"
	"github.com/sail-services/sail-go/foundation/framework/ser/cache"
	"strings"
	"time"

	"gopkg.in/ini.v1"
	"gopkg.in/redis.v2"
)

var defaultHSetName = "CACHE"

type RedisCacher struct {
	c          *redis.Client
	prefix     string
	hsetName   string
	occupyMode bool
}

func (c *RedisCacher) Put(key string, val interface{}, expire int64) error {
	key = c.prefix + key
	if expire == 0 {
		if err := c.c.Set(key, convert.To_String(val)).Err(); err != nil {
			return err
		}
	} else {
		dur, err := time.ParseDuration(convert.To_String(expire) + "s")
		if err != nil {
			return err
		}
		if err = c.c.SetEx(key, dur, convert.To_String(val)).Err(); err != nil {
			return err
		}
	}
	if c.occupyMode {
		return nil
	}
	return c.c.HSet(c.hsetName, key, "0").Err()
}

func (c *RedisCacher) Get(key string) interface{} {
	val, err := c.c.Get(c.prefix + key).Result()
	if err != nil {
		return nil
	}
	return val
}

func (c *RedisCacher) Delete(key string) error {
	if err := c.c.Del(key).Err(); err != nil {
		return err
	}
	if c.occupyMode {
		return nil
	}
	return c.c.HDel(c.hsetName, key).Err()
}

func (c *RedisCacher) Incr(key string) error {
	if !c.Is_Exist(key) {
		return fmt.Errorf("key '%s' not exist", key)
	}
	return c.c.Incr(c.prefix + key).Err()
}

func (c *RedisCacher) Decr(key string) error {
	if !c.Is_Exist(key) {
		return fmt.Errorf("key '%s' not exist", key)
	}
	return c.c.Decr(c.prefix + key).Err()
}

func (c *RedisCacher) IsExist(key string) bool {
	if c.c.Exists(c.prefix + key).Val() {
		return true
	}
	if !c.occupyMode {
		c.c.HDel(c.hsetName, c.prefix+key)
	}
	return false
}

func (c *RedisCacher) Flush() error {
	if c.occupyMode {
		return c.c.FlushDb().Err()
	}
	keys, err := c.c.HKeys(c.hsetName).Result()
	if err != nil {
		return err
	}
	if err = c.c.Del(keys...).Err(); err != nil {
		return err
	}
	return c.c.Del(c.hsetName).Err()
}

// Conn: "network=tcp,addr=127.0.0.1:6379,password=,db=0,pool_size=100,idle_timeout=180"
func (c *RedisCacher) StartAndGC(opts cache.Options) error {
	c.hsetName = "MacaronCache"
	c.occupyMode = opts.Occupy_Mode
	cfg, err := ini.Load([]byte(strings.Replace(opts.Conn, ",", "\n", -1)))
	if err != nil {
		return err
	}
	opt := &redis.Options{
		Network: "tcp",
	}
	for k, v := range cfg.Section("").KeysHash() {
		switch k {
		case "network":
			opt.Network = v
		case "addr":
			opt.Addr = v
		case "password":
			opt.Password = v
		case "db":
			opt.DB = convert.String_To(v).Must_Int64()
		case "pool_size":
			opt.PoolSize = convert.String_To(v).Must_Int()
		case "idle_timeout":
			opt.IdleTimeout, err = time.ParseDuration(v + "s")
			if err != nil {
				return fmt.Errorf("error parsing idle timeout: %v", err)
			}
		case "hset_name":
			c.hsetName = v
		case "prefix":
			c.prefix = v
		default:
			return fmt.Errorf("session/redis: unsupported option '%s'", k)
		}
	}
	c.c = redis.NewClient(opt)
	if err = c.c.Ping().Err(); err != nil {
		return err
	}
	return nil
}

func init() {
	cache.Register("redis", &RedisCacher{})
}
