package cache

import (
	"crypto/md5"
	"database/sql"
	"github.com/sail-services/sail-go/foundation/framework/ser/cache"
	"encoding/hex"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type PostgresCacher struct {
	c        *sql.DB
	interval int
}

func NewPostgresCacher() *PostgresCacher {
	return &PostgresCacher{}
}

func (c *PostgresCacher) md5(key string) string {
	m := md5.Sum([]byte(key))
	return hex.EncodeToString(m[:])
}

func (c *PostgresCacher) Put(key string, val interface{}, expire int64) error {
	item := &cache.Item{Val: val}
	data, err := cache.EncodeGob(item)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	if c.Is_Exist(key) {
		_, err = c.c.Exec("UPDATE cache SET data=$1, created=$2, expire=$3 WHERE key=$4", data, now, expire, c.md5(key))
	} else {
		_, err = c.c.Exec("INSERT INTO cache(key,data,created,expire) VALUES($1,$2,$3,$4)", c.md5(key), data, now, expire)
	}
	return err
}

func (c *PostgresCacher) read(key string) (*cache.Item, error) {
	var (
		data    []byte
		created int64
		expire  int64
	)
	err := c.c.QueryRow("SELECT data,created,expire FROM cache WHERE key=$1", c.md5(key)).Scan(&data, &created, &expire)
	if err != nil {
		return nil, err
	}

	item := new(cache.Item)
	if err = cache.DecodeGob(data, item); err != nil {
		return nil, err
	}
	item.Created = created
	item.Expire = expire
	return item, nil
}

func (c *PostgresCacher) Get(key string) interface{} {
	item, err := c.read(key)
	if err != nil {
		return nil
	}
	if item.Expire > 0 &&
		(time.Now().Unix()-item.Created) >= item.Expire {
		c.Delete(key)
		return nil
	}
	return item.Val
}

func (c *PostgresCacher) Delete(key string) error {
	_, err := c.c.Exec("DELETE FROM cache WHERE key=$1", c.md5(key))
	return err
}

func (c *PostgresCacher) Incr(key string) error {
	item, err := c.read(key)
	if err != nil {
		return err
	}
	item.Val, err = cache.Incr(item.Val)
	if err != nil {
		return err
	}
	return c.Put(key, item.Val, item.Expire)
}

func (c *PostgresCacher) Decr(key string) error {
	item, err := c.read(key)
	if err != nil {
		return err
	}
	item.Val, err = cache.Decr(item.Val)
	if err != nil {
		return err
	}
	return c.Put(key, item.Val, item.Expire)
}

func (c *PostgresCacher) Is_Exist(key string) bool {
	var data []byte
	err := c.c.QueryRow("SELECT data FROM cache WHERE key=$1", c.md5(key)).Scan(&data)
	if err != nil && err != sql.ErrNoRows {
		panic("cache/postgres: error checking existence: " + err.Error())
	}
	return err != sql.ErrNoRows
}

func (c *PostgresCacher) Flush() error {
	_, err := c.c.Exec("DELETE FROM cache")
	return err
}

func (c *PostgresCacher) startGC() {
	if c.interval < 1 {
		return
	}
	if _, err := c.c.Exec("DELETE FROM cache WHERE EXTRACT(EPOCH FROM NOW()) - created >= expire"); err != nil {
		log.Printf("cache/postgres: error garbage collecting: %v", err)
	}
	time.AfterFunc(time.Duration(c.interval)*time.Second, func() { c.startGC() })
}

func (c *PostgresCacher) Start_And_GC(opt cache.Options) (err error) {
	c.interval = opt.Interval
	c.c, err = sql.Open("postgres", opt.Conn)
	if err != nil {
		return err
	} else if err = c.c.Ping(); err != nil {
		return err
	}
	go c.startGC()
	return nil
}

func init() {
	cache.Register("postgres", NewPostgresCacher())
}
