package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	//TODO 使用TCP协议连接服务器
	conn, error := net.Dial("tcp", "127.0.0.1:9000")
	if error == nil {
		count := 1
		fmt.Printf("连接成功:%v\n", conn)
		for {
			//TODO 向服务器传输信息（持续性输出）
			//其实所有网络编程的IO，都是基于文件进行的，而非内存
			reader := bufio.NewReader(os.Stdin)
			content, error := reader.ReadString('\n')
			if error == nil {
				content = strings.Trim(content, " \r\n")
				if content == "exit" {
					fmt.Printf("bye~")
					return
				}
				//TODO 将从控制台输入的文字发送给服务端
				n, error := conn.Write([]byte(content))
				if error == nil {
					fmt.Printf("第%v次：已经成功发送内容:%v,长度为:%v字节\n", count, content, n)
					go getResponse(conn)
					count++
				} else {
					fmt.Printf("发生错误：%v", error)
				}
			}
		}
		defer conn.Close()
	}
}

func getResponse(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		response, responseError := conn.Read(buffer)
		if responseError == nil {
			fmt.Printf("接收到来自服务端的响应:%v\n", string(buffer[:response]))
		}
	}
}
