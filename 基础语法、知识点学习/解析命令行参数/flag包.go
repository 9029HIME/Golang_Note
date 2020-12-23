package main

import (
	"flag"
	"fmt"
)

//TODO Golang中获取命令行参数主要通过flag包（底层使用os.Args）
/**
TODO go build -o param.exe flag包.go
		param.exe -u root -p 123
*/
func main() {
	/**
	TODO flag包IntVar和StringVar参数介绍
		p :参数值要赋值的地方
		name : -后面的参数名
		value : 默认值
		usage : 说明
	*/
	var user string
	var password string
	flag.StringVar(&user, "u", "", "用户名的一个说明")
	flag.StringVar(&password, "p", "", "密码的一个说明")
	//TODO 这一步最重要，将os.Args中切片的值转换到p内
	flag.Parse()
	fmt.Printf("传入的用户名是：%v,传入的密码是：%v", user, password)
}
