package main

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

/**
TODO 减少连接的开启与关闭次数，此连接池基于redigo
*/
func poolGet(maxIdle, maxActive int, idleTimeout time.Duration, protocol, address, password string) *redis.Pool {
	pool := &redis.Pool{
		MaxIdle:     maxIdle,     //最大空闲连接数
		MaxActive:   maxActive,   //最大连接数
		IdleTimeout: idleTimeout, //最大空闲时间
		Dial: func() (redis.Conn, error) {
			return redis.Dial(protocol, address, redis.DialPassword(password))
		},
	}
	return pool
}
