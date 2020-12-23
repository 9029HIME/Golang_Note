package main

import (
	"fmt"
	"net"
)

/**
TODO Golang中网络相关的工具类都在net包中
*/
func main() {
	//返回值是一个Listener接口
	listener, error := net.Listen("tcp", "127.0.0.1:9000")
	if error == nil {
		//让监听者循环等待连接
		for {
			conn, error := listener.Accept()
			if error == nil {
				//TODO 起一个新的协程为请求服务
				fmt.Printf("%v客户端已经连接\n", conn.RemoteAddr().String())
				go doProcess(conn)
			}
		}
	} else {
		fmt.Println("监听失败")
	}
	defer listener.Close()
}

func doProcess(conn net.Conn) {

	for {
		buf := make([]byte, 1024)
		//TODO 如果客户端不发送消息(write)，则服务端会在这里阻塞，直到超时
		//这个n代表本次读到的数据长度
		n, error := conn.Read(buf)
		if error == nil {
			fmt.Printf("收到客户端%v的消息:%v\n", conn.RemoteAddr().String(), string(buf[:n]))
			//TODO 可以用同一个Connection给客户端发消息
			conn.Write([]byte("说完了吧？我透"))
		} else {
			fmt.Printf("客户端已退出\n")
			return
		}
	}

	defer conn.Close()
}
