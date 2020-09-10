package main

import "fmt"

/**
Golang中方法的调用者是数据类型，可以是结构体，也可以是基本数据类型
不过基本数据类型需要指定别名
*/
func main() {
	var i Integer = 1
	i.Ind()
	fmt.Println(&i)
}

type Integer int

func (i *Integer) Ind() {
	if *i == 1 {
		fmt.Println("等于1")
	}
}

/**
结构体或别名如果有一个方法叫String() string。那么fmt.println()就会默认输出String()方法的返回值。
如果方法指向的是指针，this是对象，String()无效
如果方法指向的是指针，this是指针，String()有效
如果方法指向的是对象，this是对象，String()有效
如果方法指向的是对象，this是指针，String()有效
String()无效后会调用this的默认输出
*/
func (i *Integer) String() string {
	return "Integer的tostring()被调用啊了"
}
