package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

//Redis redis cache
type Redis struct {
	conn *redis.Pool
}

//RedisOpts redis 连接属性
type RedisOpts struct {
	Host        string `yml:"host" json:"host"`
	Password    string `yml:"password" json:"password"`
	Database    int    `yml:"database" json:"database"`
	MaxIdle     int    `yml:"max_idle" json:"max_idle"`
	MaxActive   int    `yml:"max_active" json:"max_active"`
	IdleTimeout int32  `yml:"idle_timeout" json:"idle_timeout"` //second
}

//NewRedis 实例化
func NewRedis(opts *RedisOpts) *Redis {
	pool := &redis.Pool{
		MaxActive:   opts.MaxActive,
		MaxIdle:     opts.MaxIdle,
		IdleTimeout: time.Second * time.Duration(opts.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", opts.Host,
				redis.DialDatabase(opts.Database),
				redis.DialPassword(opts.Password),
			)
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := conn.Do("PING")
			return err
		},
	}
	return &Redis{pool}
}

//SetConn 设置conn
func (r *Redis) SetConn(conn *redis.Pool) {
	r.conn = conn
}

//Get 获取一个值
func (r *Redis) Get(key string) interface{} {
	conn := r.conn.Get()
	defer conn.Close()

	var data []byte
	var err error
	if data, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		return nil
	}
	var reply interface{}
	if err = json.Unmarshal(data, &reply); err != nil {
		return nil
	}

	return reply
}

//GetString 获取一个string值
func (r *Redis) GetString(key string) string {
	conn := r.conn.Get()
	defer conn.Close()

	var data []byte
	var err error
	if data, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		return ""
	}

	return string(data)
}

//Decr 获取一个string值
func (r *Redis) Decr(key string) int64 {
	conn := r.conn.Get()
	defer conn.Close()

	count, err := redis.Int64(conn.Do("DECR", key))
	if err != nil {
		fmt.Println("REDIS:::::", err)
		return -1
	}
	return count
}

//Incr 获取一个string值
func (r *Redis) Incr(key string) int64 {
	conn := r.conn.Get()
	defer conn.Close()

	count, err := redis.Int64(conn.Do("INCR", key))
	if err != nil {
		fmt.Println("REDIS:::::", err)
		return -1
	}
	return count
}

//SetString 设置一个值
func (r *Redis) SetString(key string, val string, timeout time.Duration) (err error) {
	conn := r.conn.Get()
	defer conn.Close()

	// fmt.Printf("[Redis-go] SET %s = %v\n", key, val)
	if timeout == 0*time.Second {
		_, err = conn.Do("SET", key, val)
		return
	}

	_, err = conn.Do("SETEX", key, int64(timeout/time.Second), val)

	return
}

//Set 设置一个值
func (r *Redis) Set(key string, val interface{}, timeout time.Duration) (err error) {
	conn := r.conn.Get()
	defer conn.Close()

	var data []byte
	if data, err = json.Marshal(val); err != nil {
		return
	}

	// fmt.Printf("[Redis-go] SET %s = %v\n", key, val)
	if timeout == 0*time.Second {
		_, err = conn.Do("SET", key, data)
		return
	}

	_, err = conn.Do("SETEX", key, int64(timeout/time.Second), data)

	return
}

// SetLock 设置Redis锁
func (r *Redis) SetLock(key string, val interface{}) (ok bool, err error) {
	conn := r.conn.Get()
	defer conn.Close()

	var (
		data  []byte
		reply int64
	)
	if data, err = json.Marshal(val); err != nil {
		return
	}

	// fmt.Printf("[Redis-go] SETNX %s = %v\n", key, val)
	reply, err = redis.Int64(conn.Do("SETNX", key, data))

	if reply == 1 {
		return true, err
	}
	return
}

//IsExist 判断key是否存在
func (r *Redis) IsExist(key string) bool {
	conn := r.conn.Get()
	defer conn.Close()

	a, _ := conn.Do("EXISTS", key)
	i := a.(int64)
	if i > 0 {
		return true
	}
	return false
}

//Delete 删除
func (r *Redis) Delete(key string) error {
	conn := r.conn.Get()
	defer conn.Close()
	// fmt.Println("REDIS DEL KEY")
	if _, err := conn.Do("DEL", key); err != nil {
		return err
	}

	return nil
}
