package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

/**
TODO 这里使用的是redigo
	redigo是面向redis命令形式向redis进行操作，主要调用conn.Do()这个api操作。
*/
func main() {
	//TODO 连接Redis
	pool := poolGet(10, 0, 10000, "tcp", "localhost:6379", "123")
	conn := pool.Get()

	//TODO 写字符串
	putString(conn, "name1", "hjy")
	getString(conn, "name1")
	//TODO 写Hash
	putHash(conn, "user01", "name", "pwy")
	getHash(conn, "user01", "name")
	//TODO 从连接池里拿出来的连接Close()后会回到池内
	defer conn.Close()
	//TODO 连接池关闭后，就再也拿不到连接了
	defer pool.Close()
}

func putString(conn redis.Conn, key, value string) {
	//TODO 添加的结果一般没用
	_, error := conn.Do("set", key, value)
	if error != nil {
		fmt.Printf("添加%v出错，错误信息：%v\n", key, error)
	}
}

func getString(conn redis.Conn, key string) {
	//TODO 查询的结果是空接口，需要强转
	result, error := conn.Do("get", key)
	if error != nil {
		fmt.Printf("查询%v出错，错误信息：%v\n", key, error)
	} else {
		//TODO 查询的结果本质是uint8，还要用内置api强转
		string, _ := redis.String(result, error)
		fmt.Printf("本次查询的结果是:%v\n", string)
	}
}

func putHash(conn redis.Conn, key, field, value string) {
	//TODO 添加的结果一般没用
	_, error := conn.Do("hset", key, field, value)
	if error != nil {
		fmt.Printf("添加%v出错，错误信息：%v\n", key, error)
	}
}

func getHash(conn redis.Conn, key, field string) {
	//TODO 查询的结果是空接口，需要强转
	result, error := conn.Do("hget", key, field)
	if error != nil {
		fmt.Printf("查询%v出错，错误信息：%v\n", key, error)
	} else {
		//TODO 查询的结果本质是uint8，还要用内置api强转
		string, _ := redis.String(result, error)
		fmt.Printf("本次查询的结果是:%v\n", string)
	}
}

//TODO 其他命令同理，只要将conn.Do(commandName,args ...interface{})中的commandName改成对应的redis命令
