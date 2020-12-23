package main

import (
	"fmt"
	"io/ioutil"
)

/**
TODO 常用小文件IO，关键API：ioutil.ReadFile()
*/
func smallRead() {
	filePath := "C:\\Users\\Administrator\\Desktop\\Golang文件.txt"
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("出错信息：%v", err)
	} else {
		fmt.Printf("内容是：%v", string(bytes))
	}
}
