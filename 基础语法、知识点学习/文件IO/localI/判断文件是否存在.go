package main

import (
	"fmt"
	"os"
)

/**
TODO Golang中判断文件是否存在，主要是通过os.Stat函数的返回错误值判断
*/
func judge(filePath string) {
	_, err := os.Stat(filePath)
	if err == nil {
		fmt.Println("文件存在")
	} else if os.IsNotExist(err) {
		fmt.Println("文件不存在")
	} else {
		fmt.Println("无法确定")
	}
}
