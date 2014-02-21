package rkv

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

type RediStore struct {
	pool *redis.Pool
}

func NewRediStore(addr string, db int) (*RediStore, error) {
	pool := redis.NewPool(func() (redis.Conn, error) {
		conn, err := redis.DialTimeout("tcp", addr, 10*time.Second, 10*time.Second, 10*time.Second)
		if err == nil {
			//conn = redis.NewLoggingConn(conn, log.New(os.Stdout, "!", 0), fmt.Sprint(rand.Intn(900)+100))
			err = conn.Send("SELECT", db)
		}
		return conn, err
	}, 10)
	return &RediStore{pool}, nil
}

func (r *RediStore) Get(key string) (v Value, err error) {
	c := r.pool.Get()
	defer c.Close()

	var values []interface{}
	values, err = redis.Values(c.Do("HMGET", key, "rev", "ts", "c"))
	if err != nil {
		return
	}
	var timebuf int64
	_, err = redis.Scan(values, &v.Rev, &timebuf, &v.Content)
	v.Timestamp = time.Unix(0, timebuf)
	return
}

func (r *RediStore) Put(key string, p Value) error {
	c := r.pool.Get()
	defer c.Close()

	c.Send("WATCH", key)
	lastrev, err := redis.Int64(c.Do("HGET", key, "rev"))
	if err != nil && err != redis.ErrNil {
		c.Send("UNWATCH")
		return err
	}
	if lastrev+1 != p.Rev {
		c.Send("UNWATCH")
		return ErrRevNotMatch
	}
	c.Send("MULTI")
	c.Send("HMSET", key, "rev", p.Rev, "ts", time.Now().UnixNano(), "c", p.Content)
	_, err = redis.Values(c.Do("EXEC"))
	if err == redis.ErrNil {
		return ErrRevNotMatch
	}
	return err
}
