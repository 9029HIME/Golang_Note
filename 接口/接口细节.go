package main

import "fmt"

/**
TODO:Golang接口也支持继承，方式与结构体继承一样
*/
type C interface {
	//TODO:从目前的版本来说，接口都是没有方法体和变量的（不知道后面会不会像Java8有静态方法、静态变量）
	method1(var1 string) string
	method2(var2 string) string
}

func test() {
	//TODO Golang接口和Java类似，可以用多态指向变量实例
	var c C = new(InterfaceUser)
	c.method1("a")
	//TODO Golang所有结构体都实现了空接口interface{}
	var d interface{} = new(InterfaceUser)
	fmt.Println(d)
}

//TODO 基本数据类型也可以实现接口，前提是指定别名
type Integer int

func (i *Integer) method1(var1 string) string {
	return "Integer的1方法"
}

func (i *Integer) method2(var1 string) string {
	return "Integer的1方法"
}
