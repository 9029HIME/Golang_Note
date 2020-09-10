package main

import "fmt"

/**
Golang中两个结构体的强转前提：两个结构体必须有相同个数、相同名称、相同类型的属性
*/
func main() {
	var a A
	var b B
	a = A(b)
	fmt.Println(a)

}

/**
属性别名：
在Golang中变量名最好是大写开头，但这样JSON化的时候也是大写，因此需要给属性作一个别名
底层是通过反射转换
命名格式：`tag:value`
*/
type B struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type A struct {
	Name string
	Age  int
}
